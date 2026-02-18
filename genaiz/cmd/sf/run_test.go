package sf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/docker"
)

func TestRunExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: newRunTestOptions(),
	}
	var expectedDockerContext = "RunDockerContext"
	var expectedDockerFile = "RunDockerfile"
	var expectedDockerTag = "RunDockerTag"
	var expectedDockerVersion = "RunDockerVersion"
	var expectedEnvFile = "RunEnvFile"
	var expectedEnvVars = "RunEnvVars"
	var expectedMountInput = "RunMountInput"
	var expectedMountOutput = "RunMountOutput"
	var expectedMountLog = "RunMountLog"
	var expectedMountVar = "RunMountVar"
	var expectedRunImage = "RunRunImage"
	var expectedRunPrefix = "RunRunPrefix"

	testViper.Set(testCli.optionDockerContext.Key, expectedDockerContext)
	testViper.Set(testCli.optionDockerFile.Key, expectedDockerFile)
	testViper.Set(testCli.optionDockerTag.Key, expectedDockerTag)
	testViper.Set(testCli.optionDockerVersion.Key, expectedDockerVersion)
	testViper.Set(testExecutor.optionEnvFile.Key, expectedEnvFile)
	testViper.Set(testExecutor.optionEnvVars.Key, expectedEnvVars)
	testViper.Set(testExecutor.optionMountInput.Key, expectedMountInput)
	testViper.Set(testExecutor.optionMountOutput.Key, expectedMountOutput)
	testViper.Set(testExecutor.optionMountLog.Key, expectedMountLog)
	testViper.Set(testExecutor.optionMountVar.Key, expectedMountVar)
	testViper.Set(testExecutor.optionRunImage.Key, expectedRunImage)
	testViper.Set(testExecutor.optionRunPrefix.Key, expectedRunPrefix)
	testExecutor.Display()

	if out := testOutput.String(); out != "" {
		assert.Contains(t, out, expectedDockerContext)
		assert.Contains(t, out, expectedDockerFile)
		assert.Contains(t, out, expectedDockerTag)
		assert.Contains(t, out, expectedDockerVersion)
		assert.Contains(t, out, expectedEnvFile)
		assert.Contains(t, out, expectedEnvVars)
		assert.Contains(t, out, expectedMountInput)
		assert.Contains(t, out, expectedMountOutput)
		assert.Contains(t, out, expectedMountLog)
		assert.Contains(t, out, expectedMountVar)
		assert.Contains(t, out, expectedRunImage)
		assert.Contains(t, out, expectedRunPrefix)
	} else {
		assert.Fail(t, "output is empty")
	}
}

func TestRunExecutor_Pretend(t *testing.T) {
	var calledBuild, calledRun bool
	var capturedDataLinkParams broker.DataLinkParams
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SyncExecutor: SyncExecutor{
			innerSources:           &config.ListOption{Option: config.Option{Key: "innerStores"}},
			innerStores:            &config.ListOption{Option: config.Option{Key: "innerSources"}},
			dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink}),
			exportTaskFactory:      newExportLinkPretendCapture(&capturedDataLinkParams),
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		runTaskFactory:   newRunTaskPretendStub(&calledRun),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testViper.Set(testExecutor.innerSources.Key, []string{fmt.Sprintf("%s/%s:%s",
			testDataLink.Oem, testDataLink.Handle, testDataLink.Version)})
		testLedger.InitLogging()
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.True(t, calledRun)
		assert.Equal(t, testDataLink, capturedDataLinkParams.DataLink)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunExecutor_Pretend_DataLinkError(t *testing.T) {
	var calledBuild, calledRun bool
	var capturedDataLinkParams broker.DataLinkParams
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SyncExecutor: SyncExecutor{
			innerSources:           &config.ListOption{Option: config.Option{Key: "innerStores"}},
			innerStores:            &config.ListOption{Option: config.Option{Key: "innerSources"}},
			dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{}),
			exportTaskFactory:      newExportLinkPretendCapture(&capturedDataLinkParams),
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		runTaskFactory:   newRunTaskPretendStub(&calledRun),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		var patch = mock.Patches{T: t}.OsExit(func(int) {})

		defer patch.Unpatch()
		defer filez.CloseSilently(fd)
		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testViper.Set(testExecutor.innerSources.Key, []string{"notValid/noVersion"})
		testLedger.InitLogging()
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.False(t, calledRun)
		assert.NotEmpty(t, patch.CalledWith)
		assert.EqualValues(t, 1, patch.CalledWith)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunExecutor_Pretend_EnvMapError(t *testing.T) {
	var calledBuild, calledRun bool
	var testFile = filepath.Join(t.TempDir(), ".env-test")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		runTaskFactory:   newRunTaskPretendStub(&calledRun),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)

		if _, err = fd.Write([]byte("syntax_error")); err == nil {
			var patch = mock.Patches{T: t}.OsExit(func(int) {})

			defer patch.Unpatch()
			testViper.Set(testExecutor.optionEnvFile.Key, testFile)
			testExecutor.Pretend()
			assert.False(t, calledBuild)
			assert.False(t, calledRun)
			assert.True(t, patch.Called)
			assert.EqualValues(t, 1, patch.CalledWith)
			return
		}
	}

	assert.NoError(t, err)
}

func TestRunExecutor_Pretend_NoSync(t *testing.T) {
	var calledBuild, calledRun bool
	var capturedDataLinkParams broker.DataLinkParams
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SyncExecutor: SyncExecutor{
			innerSources:           &config.ListOption{Option: config.Option{Key: "innerStores"}},
			innerStores:            &config.ListOption{Option: config.Option{Key: "innerSources"}},
			collectTaskFactory:     newCollectLinkPretendCapture(&capturedDataLinkParams),
			dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink}),
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		runTaskFactory:   newRunTaskPretendStub(&calledRun),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testViper.Set(testExecutor.optionNoPropSync.Key, "True")
		testViper.Set(testExecutor.innerSources.Key, []string{fmt.Sprintf("%s/%s:%s",
			testDataLink.Oem, testDataLink.Handle, testDataLink.Version)})
		testLedger.InitLogging()
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.True(t, calledRun)
		assert.Equal(t, testDataLink, capturedDataLinkParams.DataLink)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunExecutor_Pretend_RebuildImage(t *testing.T) {
	var calledBuild, calledRun bool
	var capturedDataLinkParams broker.DataLinkParams
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SyncExecutor: SyncExecutor{
			innerSources:           &config.ListOption{Option: config.Option{Key: "innerStores"}},
			innerStores:            &config.ListOption{Option: config.Option{Key: "innerSources"}},
			collectTaskFactory:     newCollectLinkPretendCapture(&capturedDataLinkParams),
			dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{}),
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		runTaskFactory:   newRunTaskPretendStub(&calledRun),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testViper.Set(testExecutor.optionNoPropSync.Key, "True")
		testLedger.InitLogging()
		testExecutor.rebuildImage = true
		testExecutor.Pretend()
		assert.True(t, calledBuild)
		assert.True(t, calledRun)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunExecutor_Proceed(t *testing.T) {
	var calledBuild, calledRun bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SyncExecutor: SyncExecutor{
			innerSources:           &config.ListOption{Option: config.Option{Key: "innerStores"}},
			innerStores:            &config.ListOption{Option: config.Option{Key: "innerSources"}},
			dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{}),
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		runTaskFactory:   newRunTaskCompleteStub(&calledRun),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.InitLogging()
		testExecutor.rebuildImage = true
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.True(t, calledRun)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunExecutor_Proceed_EnvMapError(t *testing.T) {
	var calledBuild, calledRun bool
	var testFile = filepath.Join(t.TempDir(), ".env-test")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		runTaskFactory:   newRunTaskPretendStub(&calledRun),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)

		if _, err = fd.Write([]byte("syntax_error")); err == nil {
			var patch = mock.Patches{T: t}.OsExit(func(int) {})

			defer patch.Unpatch()
			testViper.Set(testExecutor.optionEnvFile.Key, testFile)
			testExecutor.Proceed()
			assert.False(t, calledBuild)
			assert.False(t, calledRun)
			assert.True(t, patch.Called)
			assert.EqualValues(t, 1, patch.CalledWith)
			return
		}
	}

	assert.NoError(t, err)
}

func TestRunExecutor_Proceed_NoSync(t *testing.T) {
	var calledBuild, calledRun bool
	var capturedDataLinkParams broker.DataLinkParams
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SyncExecutor: SyncExecutor{
			innerSources:           &config.ListOption{Option: config.Option{Key: "innerStores"}},
			innerStores:            &config.ListOption{Option: config.Option{Key: "innerSources"}},
			collectTaskFactory:     newCollectLinkCompleteCapture(&capturedDataLinkParams),
			dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink}),
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		runTaskFactory:   newRunTaskCompleteStub(&calledRun),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testViper.Set(testExecutor.optionNoPropSync.Key, "True")
		testViper.Set(testExecutor.innerStores.Key, []string{fmt.Sprintf("%s/%s:%s",
			testDataLink.Oem, testDataLink.Handle, testDataLink.Version)})
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.False(t, calledBuild)
		assert.True(t, calledRun)
		assert.Equal(t, testDataLink, capturedDataLinkParams.DataLink)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunExecutor_Proceed_SyncSpecs(t *testing.T) {
	var calledBuild bool
	var capturedContainerParams docker.ContainerParams
	var capturedDataLinkParams broker.DataLinkParams
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testPropSpec = &broker.PropSpec{
		Key:   "key",
		Value: "value",
	}
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SyncExecutor: SyncExecutor{
			innerSources:           &config.ListOption{Option: config.Option{Key: "innerStores"}},
			innerStores:            &config.ListOption{Option: config.Option{Key: "innerSources"}},
			dataLinksWriterFactory: newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink}),
			exportTaskFactory:      newExportLinkCompleteCapture(&capturedDataLinkParams),
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		runTaskFactory:   newRunTaskCompleteCapture(&capturedContainerParams),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testViper.Set(testExecutor.innerStores.Key, []string{fmt.Sprintf("%s/%s:%s",
			testDataLink.Oem, testDataLink.Handle, testDataLink.Version)})
		testViper.Set(schema.Genaiz.Function.Publish.PropSpecs.Doc, []broker.PropSpec{*testPropSpec})
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.False(t, calledBuild)
		assert.Equal(t, testPropSpec.Key, capturedContainerParams.VarSpecs[0].GetKey())
		assert.Equal(t, testPropSpec.Value, capturedContainerParams.VarSpecs[0].GetDefaultValue())
		assert.Equal(t, testDataLink, capturedDataLinkParams.DataLink)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunOptions_allDefiners(t *testing.T) {
	var testOptions = newRunTestOptions()
	var definers = testOptions.allDefiners()

	assert.Contains(t, definers, testOptions.optionEnvFile)
	assert.Contains(t, definers, testOptions.optionEnvVars)
	assert.Contains(t, definers, testOptions.optionMountInput)
	assert.Contains(t, definers, testOptions.optionMountOutput)
	assert.Contains(t, definers, testOptions.optionMountLog)
	assert.Contains(t, definers, testOptions.optionMountVar)
	assert.Contains(t, definers, testOptions.optionNoPropSync)
	assert.Contains(t, definers, testOptions.optionRunImage)
	assert.Contains(t, definers, testOptions.optionRunPrefix)
	assert.Equal(t, 9, len(definers))
}

func TestNewRunOptions(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewRunOptions(testCli)

	assert.NotEmpty(t, testOptions.optionEnvFile)
	assert.NotEmpty(t, testOptions.optionEnvVars)
	assert.NotEmpty(t, testOptions.optionMountInput)
	assert.NotEmpty(t, testOptions.optionMountOutput)
	assert.NotEmpty(t, testOptions.optionMountLog)
	assert.NotEmpty(t, testOptions.optionMountVar)
	assert.NotEmpty(t, testOptions.optionNoPropSync)
	assert.NotEmpty(t, testOptions.optionRunImage)
	assert.NotEmpty(t, testOptions.optionRunPrefix)
	assert.False(t, testOptions.rebuildImage)
}

func TestNewRun(t *testing.T) {
	var testDir = t.TempDir()
	var runCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithOutput(io.Writer(testOutput)).WithViper(testViper).Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testRun = NewRun(testLedger, testCli)
	var expectedTag = "dockerTag"

	testRun.PostRun = func(cmd *cobra.Command, args []string) {
		runCompleted = true
	}

	testViper.Set(testCli.optionDockerTag.Key, expectedTag)
	testViper.Set(schema.Genaiz.Function.Run.MountOutput.Doc, testDir)
	assert.NoError(t, testRun.Execute())
	assert.True(t, runCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedTag)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func newRunTaskPretendStub(flag *bool) RunTaskFactory {
	return func() *task.Task[docker.ContainerParams] {
		return &task.Task[docker.ContainerParams]{
			OnPrepare: func(params *docker.ContainerParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *docker.ContainerParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newRunTaskCompleteStub(flag *bool) RunTaskFactory {
	return func() *task.Task[docker.ContainerParams] {
		return &task.Task[docker.ContainerParams]{
			Name: "run_test",
			OnPrepare: func(params *docker.ContainerParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *docker.ContainerParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newRunTaskCompleteCapture(capture *docker.ContainerParams) RunTaskFactory {
	return func() *task.Task[docker.ContainerParams] {
		return &task.Task[docker.ContainerParams]{
			Name: "run_test",
			OnPrepare: func(params *docker.ContainerParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *docker.ContainerParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newRunTestOptions() *RunOptions {
	return &RunOptions{
		EnvOptions: EnvOptions{
			optionEnvFile: cli.Options.Docker.EnvFile().
				WithKeys(&schema.Genaiz.Function.Run.EnvFile).
				BuildStringOption(),
			optionEnvVars: cli.Options.Docker.EnvVar().
				WithKeys(&schema.Genaiz.Function.Run.EnvVars).
				BuildListOption(),
		},
		optionMountInput: cli.Options.Functions.MountInput().
			WithKeys(&schema.Genaiz.Function.Run.MountInput).
			BuildStringOption(),
		optionMountOutput: cli.Options.Functions.MountOutput().
			WithKeys(&schema.Genaiz.Function.Run.MountOutput).
			BuildStringOption(),
		optionMountLog: cli.Options.Functions.MountLog().
			WithKeys(&schema.Genaiz.Function.Run.MountLog).
			BuildStringOption(),
		optionMountVar: cli.Options.Functions.MountVar().
			WithKeys(&schema.Genaiz.Function.Run.MountVar).
			BuildStringOption(),
		optionRunImage: cli.Options.Docker.Image().
			WithKeys(&schema.Genaiz.Function.Run.Image).
			BuildStringOption(),
		optionNoPropSync: cli.Options.Functions.NoPropSync().
			WithKeys(&schema.Genaiz.Function.Run.NoPropSync).
			BuildBoolOption(),
		optionRunPrefix: cli.Options.Docker.ContainerPrefix().
			WithKeys(&schema.Genaiz.Function.Run.Prefix).
			BuildStringOption(),
	}
}

func newCollectLinkCompleteCapture(capture *broker.DataLinkParams) CollectTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newCollectLinkPretendCapture(capture *broker.DataLinkParams) CollectTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataLinkParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newExportLinkCompleteCapture(capture *broker.DataLinkParams) ExportTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newExportLinkPretendCapture(capture *broker.DataLinkParams) ExportTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataLinkParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
