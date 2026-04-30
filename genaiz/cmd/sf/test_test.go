package sf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

func TestTestExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testExecutor = &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: newTestTestOptions(),
	}
	var expectedDockerContext = "TestDockerContext"
	var expectedDockerFile = "TestDockerfile"
	var expectedDockerRepo = "TestDockerRepo"
	var expectedDockerVersion = "TestDockerVersion"
	var expectedEnvFile = "TestEnvFile"
	var expectedEnvVars = "TestEnvVars"
	var expectedMountInput = "TestMountInput"
	var expectedMountOutput = "TestMountOutput"
	var expectedMountLog = "TestMountLog"
	var expectedMountVar = "TestMountVar"
	var expectedRunImage = "TestRunImage"
	var expectedRunPrefix = "TestContainerPrefix"

	testViper.Set(testCli.optionDockerContext.Key, expectedDockerContext)
	testViper.Set(testCli.optionDockerFile.Key, expectedDockerFile)
	testViper.Set(testCli.optionDockerRepo.Key, expectedDockerRepo)
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

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedDockerContext)
		assert.Contains(t, actual, expectedDockerFile)
		assert.Contains(t, actual, expectedDockerRepo)
		assert.Contains(t, actual, expectedDockerVersion)
		assert.Contains(t, actual, expectedEnvFile)
		assert.Contains(t, actual, expectedEnvVars)
		assert.Contains(t, actual, expectedMountInput)
		assert.Contains(t, actual, expectedMountOutput)
		assert.Contains(t, actual, expectedMountLog)
		assert.Contains(t, actual, expectedMountVar)
		assert.Contains(t, actual, expectedRunImage)
		assert.Contains(t, actual, expectedRunPrefix)
	} else {
		assert.Fail(t, "output is empty")
	}
}

func TestTestExecutor_Pretend(t *testing.T) {
	var calledBuild bool
	var calledTest int
	var capturedDataLinkParams broker.DataLinkParams
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testExecutor = &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SyncBridge: dk.NewSyncBridgeBuilder().
			WithDataLinksWriterFactory(newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink})).
			WithExportLinkTaskFactory(newExportLinkPretendCapture(&capturedDataLinkParams)).
			Build(),
		RunOptions: newTestTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		testTaskFactory:  newContainerTaskPretendStub(&calledTest),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.EqualValues(t, 1, calledTest)
		assert.Empty(t, capturedDataLinkParams)
	} else {
		assert.NoError(t, err)
	}
}

func TestTestExecutor_Pretend_EnvMapError(t *testing.T) {
	var calledBuild bool
	var calledTest int
	var capturedDataLinkParams broker.DataLinkParams
	var testFile = filepath.Join(t.TempDir(), ".env-test")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testExecutor = &TestExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SyncBridge: dk.NewSyncBridgeBuilder().
			WithDataLinksWriterFactory(newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink})).
			WithExportLinkTaskFactory(newExportLinkPretendCapture(&capturedDataLinkParams)).
			Build(),
		RunOptions: newTestTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		testTaskFactory:  newContainerTaskPretendStub(&calledTest),
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
			assert.EqualValues(t, 0, calledTest)
			assert.True(t, patch.Called)
			assert.EqualValues(t, 1, patch.CalledWith)
			assert.Empty(t, capturedDataLinkParams)
			return
		}
	}

	assert.NoError(t, err)
}

func TestTestExecutor_Pretend_WithBuild(t *testing.T) {
	var calledBuild bool
	var calledTest int
	var capturedDataLinkParams broker.DataLinkParams
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testExecutor = &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SyncBridge: dk.NewSyncBridgeBuilder().
			WithDataLinksWriterFactory(newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink})).
			WithExportLinkTaskFactory(newExportLinkPretendCapture(&capturedDataLinkParams)).
			Build(),
		RunOptions: newTestTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		testTaskFactory:  newContainerTaskPretendStub(&calledTest),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.rebuildImage = true
		testExecutor.Pretend()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledTest)
		assert.Empty(t, capturedDataLinkParams)
	} else {
		assert.NoError(t, err)
	}
}

func TestTestExecutor_Proceed(t *testing.T) {
	var calledBuild bool
	var calledTest int
	var capturedDataLinkParams broker.DataLinkParams
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testExecutor = &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		SyncBridge: dk.NewSyncBridgeBuilder().
			WithDataLinksWriterFactory(newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink})).
			WithExportLinkTaskFactory(newExportLinkPretendCapture(&capturedDataLinkParams)).
			Build(),
		RunOptions: newTestTestOptions(),

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		testTaskFactory:  newContainerTaskCompleteStub(&calledTest),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.rebuildImage = true
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledTest)
		assert.Empty(t, capturedDataLinkParams)
	} else {
		assert.NoError(t, err)
	}
}

func TestTestExecutor_Proceed_EnvMapError(t *testing.T) {
	var calledBuild bool
	var calledTest int
	var capturedDataLinkParams broker.DataLinkParams
	var testFile = filepath.Join(t.TempDir(), ".env-test")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testDataLink = &broker.DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "0.1.1",
	}
	var testExecutor = &TestExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SyncBridge: dk.NewSyncBridgeBuilder().
			WithDataLinksWriterFactory(newDataLinksWriterTestFactory([]broker.DataLink{*testDataLink})).
			WithExportLinkTaskFactory(newExportLinkPretendCapture(&capturedDataLinkParams)).
			Build(),
		RunOptions: newTestTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		testTaskFactory:  newContainerTaskPretendStub(&calledTest),
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
			assert.EqualValues(t, 0, calledTest)
			assert.True(t, patch.Called)
			assert.EqualValues(t, 1, patch.CalledWith)
			assert.Empty(t, capturedDataLinkParams)
			return
		}
	}

	assert.NoError(t, err)
}

func TestNewTestOptions(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewTestOptions(testCli)

	assert.NotEmpty(t, testOptions.optionEnvFile)
	assert.NotEmpty(t, testOptions.optionEnvVars)
	assert.NotEmpty(t, testOptions.optionMountInput)
	assert.NotEmpty(t, testOptions.optionMountOutput)
	assert.NotEmpty(t, testOptions.optionMountLog)
	assert.NotEmpty(t, testOptions.optionMountVar)
	assert.NotEmpty(t, testOptions.optionRunImage)
	assert.NotEmpty(t, testOptions.optionRunPrefix)
}

func TestNewTestOptions_DefaultRunMounts(t *testing.T) {
	var testDir = t.TempDir()
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewTestOptions(testCli)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()

	if err := os.MkdirAll(filepath.Join(testDir, "/run/in"), 0755); err == nil {
		testLedger.WorkDir = testDir
		actualIn := testLedger.GetString(testOptions.optionMountInput)
		assert.True(t, strings.HasPrefix(actualIn, testDir))
		assert.True(t, strings.HasSuffix(actualIn, "/in"))
		actualOut := testLedger.GetString(testOptions.optionMountOutput)
		assert.True(t, strings.HasPrefix(actualOut, testDir))
		assert.True(t, strings.HasSuffix(actualOut, "/out"))
		actualLog := testLedger.GetString(testOptions.optionMountLog)
		assert.True(t, strings.HasPrefix(actualLog, testDir))
		assert.True(t, strings.HasSuffix(actualLog, "/log"))
		actualVar := testLedger.GetString(testOptions.optionMountVar)
		assert.True(t, strings.HasPrefix(actualVar, testDir))
		assert.True(t, strings.HasSuffix(actualVar, "/var"))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestNewTest(t *testing.T) {
	var testDir = t.TempDir()
	var testCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(testViper).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testTest = NewTest(testLedger, testCli)
	var expectedImage = "dockerImage"

	testTest.PostRun = func(cmd *cobra.Command, args []string) {
		testCompleted = true
	}

	testViper.Set(testCli.optionDockerRepo.Key, "namespace/repo")
	testViper.Set(schema.Genaiz.Function.Test.MountInput.Doc, testDir)
	testViper.Set(schema.Genaiz.Function.Test.MountOutput.Doc, testDir)
	testViper.Set(schema.Genaiz.Function.Test.Image.Doc, expectedImage)
	testLedger.WorkDir = testDir
	assert.NoError(t, testTest.Execute())
	assert.True(t, testCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedImage)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func newTestTestOptions() *RunOptions {
	return &RunOptions{
		EnvOptions: EnvOptions{
			optionEnvFile: cli.Options.Docker.EnvFile().
				WithKeys(&schema.Genaiz.Function.Test.EnvFile).
				BuildStringOption(),
			optionEnvVars: cli.Options.Docker.EnvVar().
				WithKeys(&schema.Genaiz.Function.Test.EnvVars).
				BuildListOption(),
		},
		InnerOptions: makeInnerOptions(),
		optionMountInput: cli.Options.Functions.MountInput().
			WithKeys(&schema.Genaiz.Function.Test.MountInput).
			BuildStringOption(),
		optionMountOutput: cli.Options.Functions.MountOutput().
			WithKeys(&schema.Genaiz.Function.Test.MountOutput).
			BuildStringOption(),
		optionMountLog: cli.Options.Functions.MountLog().
			WithKeys(&schema.Genaiz.Function.Test.MountLog).
			BuildStringOption(),
		optionMountVar: cli.Options.Functions.MountVar().
			WithKeys(&schema.Genaiz.Function.Test.MountVar).
			BuildStringOption(),
		optionRunImage: cli.Options.Docker.Image().
			WithKeys(&schema.Genaiz.Function.Test.Image).
			BuildStringOption(),
		optionRunPrefix: cli.Options.Docker.ContainerPrefix().
			WithKeys(&schema.Genaiz.Function.Test.Prefix).
			BuildStringOption(),
	}
}
