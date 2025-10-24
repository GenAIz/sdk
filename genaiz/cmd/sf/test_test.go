package sf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
)

func TestTestExecutor_Display(t *testing.T) {
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
	var testOptions = &RunOptions{
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
	var testExecutor = &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: testOptions,
	}
	var expectedDockerContext = "TestDockerContext"
	var expectedDockerFile = "TestDockerfile"
	var expectedDockerTag = "TestDockerTag"
	var expectedDockerVersion = "TestDockerVersion"
	var expectedRunImage = "TestRunImage"

	testViper.Set(testCli.optionDockerContext.Key, expectedDockerContext)
	testViper.Set(testCli.optionDockerFile.Key, expectedDockerFile)
	testViper.Set(testCli.optionDockerTag.Key, expectedDockerTag)
	testViper.Set(testCli.optionDockerVersion.Key, expectedDockerVersion)
	testViper.Set(testOptions.optionRunImage.Key, expectedRunImage)
	testExecutor.Display()

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedDockerContext)
		assert.Contains(t, actual, expectedDockerFile)
		assert.Contains(t, actual, expectedDockerTag)
		assert.Contains(t, actual, expectedDockerVersion)
		assert.Contains(t, actual, expectedRunImage)
	} else {
		assert.Fail(t, "output is empty")
	}
}

func TestTestExecutor_Pretend(t *testing.T) {
	var calledBuild bool
	var calledTest int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: &RunOptions{
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
		},

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		testTaskFactory:  newContainerTaskPretendStub(&calledTest),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.EqualValues(t, 1, calledTest)
	} else {
		assert.NoError(t, err)
	}
}

func TestTestExecutor_Pretend_WithBuild(t *testing.T) {
	var calledBuild bool
	var calledTest int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: &RunOptions{
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
		},

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		testTaskFactory:  newContainerTaskPretendStub(&calledTest),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.rebuildImage = true
		testExecutor.Pretend()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledTest)
	} else {
		assert.NoError(t, err)
	}
}

func TestTestExecutor_Proceed(t *testing.T) {
	var calledBuild bool
	var calledTest int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: &RunOptions{
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
			rebuildImage: true,
		},

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		testTaskFactory:  newContainerTaskCompleteStub(&calledTest),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledTest)
	} else {
		assert.NoError(t, err)
	}
}

func TestNewTestOptions(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewTestOptions(testCli)

	assert.NotEmpty(t, testOptions.optionMountInput)
	assert.NotEmpty(t, testOptions.optionMountOutput)
	assert.NotEmpty(t, testOptions.optionMountLog)
	assert.NotEmpty(t, testOptions.optionMountVar)
	assert.NotEmpty(t, testOptions.optionRunImage)
	assert.NotEmpty(t, testOptions.optionRunPrefix)
}

func TestNewTest(t *testing.T) {
	var testCompleted = false
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
	var testTest = NewTest(testLedger, testCli)
	var expectedImage = "dockerImage"

	testTest.PostRun = func(cmd *cobra.Command, args []string) {
		testCompleted = true
	}

	testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
	testViper.Set(schema.Genaiz.Function.Test.Image.Doc, expectedImage)
	assert.NoError(t, testTest.Execute())
	assert.True(t, testCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedImage)
	} else {
		assert.Fail(t, "no --dry content")
	}
}
