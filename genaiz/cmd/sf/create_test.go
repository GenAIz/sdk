package sf

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

func TestCreatorExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewCreateOptions(NewSfCli(nil, nil, nil))
	var expectedArches = []string{layout.ArchTypeArm64, layout.ArchTypeX86}
	var expectedHandle = "handle"
	var expectedName = "name-create"
	var expectedOem = "oem"
	var expectedRecipe = "recipe"
	var expectedVersion = "0.0.0"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionArches.Key, expectedArches)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionOem.Key, expectedOem)
	testViper.Set(testOptions.optionRecipe.Key, expectedRecipe)
	testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
	testViper.Set(testOptions.optionVersion.Key, expectedVersion)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionArches.Param+`:[\s\t]*\[`+strings.Join(expectedArches, " ")+`\]`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionRecipe.Param+`:[\s\t]*`+expectedRecipe), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionType.Param+`:[\s\t]*`+layout.FunctionTypeFunction), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
}

func TestCreatorExecutor_PretendNoRecipe(t *testing.T) {
	var calledCreate, calledInit, calledRecipe bool
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		CreateOptions: NewCreateOptions(testCli),

		initTaskFactory:   newInitTaskPretendStub(&calledInit),
		createTaskFactory: newCreateTaskPretendStub(&calledCreate),
		recipeTaskFactory: newRecipeTaskPretendStub(&calledRecipe),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testViper.Set(testExecutor.optionType.Key, layout.FunctionTypeFunction)
	testViper.Set(testExecutor.optionHandle.Key, "create-pretend")
	testViper.Set(testExecutor.optionOem.Key, "oem")
	testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
	testExecutor.Pretend()
	assert.True(t, calledCreate)
	assert.False(t, calledRecipe)
	assert.True(t, calledInit)
}

func TestCreatorExecutor_PretendWithRecipe(t *testing.T) {
	var calledCreate, calledInit, calledRecipe bool
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		CreateOptions: NewCreateOptions(testCli),

		initTaskFactory:   newInitTaskPretendStub(&calledInit),
		createTaskFactory: newCreateTaskPretendStub(&calledCreate),
		recipeTaskFactory: newRecipeTaskPretendStub(&calledRecipe),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testViper.Set(testExecutor.optionType.Key, layout.FunctionTypeFunction)
	testViper.Set(testExecutor.optionHandle.Key, "create-pretend")
	testViper.Set(testExecutor.optionOem.Key, "oem")
	testViper.Set(testExecutor.optionRecipe.Key, "test-recipe")
	testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
	testExecutor.Pretend()
	assert.True(t, calledCreate)
	assert.True(t, calledRecipe)
	assert.True(t, calledInit)
}

func TestCreatorExecutor_ProceedNoRecipe(t *testing.T) {
	var calledCreate, calledInit, calledRecipe bool
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		CreateOptions: NewCreateOptions(testCli),

		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
		}),
		createTaskFactory: newCreateTaskCompleteStub(&calledCreate),
		recipeTaskFactory: newRecipeTaskCompleteStub(&calledRecipe),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testViper.Set(testExecutor.optionType.Key, layout.FunctionTypeFunction)
	testViper.Set(testExecutor.optionHandle.Key, "create-proceed")
	testViper.Set(testExecutor.optionOem.Key, "oem")
	testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Proceed()
	assert.True(t, calledCreate)
	assert.False(t, calledRecipe)
	assert.True(t, calledInit)
}

func TestCreatorExecutor_ProceedWithRecipe(t *testing.T) {
	var calledCreate, calledInit, calledRecipe bool
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		CreateOptions: NewCreateOptions(testCli),

		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
		}),
		createTaskFactory: newCreateTaskCompleteStub(&calledCreate),
		recipeTaskFactory: newRecipeTaskCompleteStub(&calledRecipe),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testViper.Set(testExecutor.optionType.Key, layout.FunctionTypeFunction)
	testViper.Set(testExecutor.optionHandle.Key, "create-proceed")
	testViper.Set(testExecutor.optionOem.Key, "oem")
	testViper.Set(testExecutor.optionRecipe.Key, "test-recipe")
	testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Proceed()
	assert.True(t, calledCreate)
	assert.True(t, calledRecipe)
	assert.True(t, calledInit)
}

func TestNewCreate(t *testing.T) {
	var createCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithOutput(io.Writer(testOutput)).WithViper(testViper).Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
		optionDockerContext: newOptionDockerContext(),
		optionDockerFile:    newOptionDockerFile(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testCreate = NewCreate(testLedger, testCli)
	var expectedFolder = "test-folder"

	testCreate.PostRun = func(cmd *cobra.Command, args []string) {
		createCompleted = true
	}
	testCreate.SetArgs([]string{expectedFolder})
	assert.NoError(t, testCreate.Execute())
	assert.True(t, createCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedFolder)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func TestNewCreate_InvalidFolder(t *testing.T) {
	var createCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithOutput(io.Writer(testOutput)).WithViper(testViper).Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
		optionDockerContext: newOptionDockerContext(),
		optionDockerFile:    newOptionDockerFile(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testCreate = NewCreate(testLedger, testCli)
	var expectedFolder = "#invalidtest-folder"

	testCreate.PostRun = func(cmd *cobra.Command, args []string) {
		createCompleted = true
	}
	testCreate.SetArgs([]string{expectedFolder})
	assert.Error(t, testCreate.Execute())
	assert.False(t, createCompleted)
}

func newCreateTaskCompleteStub(flag *bool) CreateTaskFactory {
	return func() *task.Task[layout.CreateParams] {
		return &task.Task[layout.CreateParams]{
			OnPrepare: func(params *layout.CreateParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *layout.CreateParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newRecipeTaskCompleteStub(flag *bool) RecipeTaskFactory {
	return func(paths ...string) *task.Task[layout.RecipeParams] {
		return &task.Task[layout.RecipeParams]{
			OnPrepare: func(params *layout.RecipeParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *layout.RecipeParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newCreateTaskPretendStub(flag *bool) CreateTaskFactory {
	return func() *task.Task[layout.CreateParams] {
		return &task.Task[layout.CreateParams]{
			OnPrepare: func(params *layout.CreateParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *layout.CreateParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newRecipeTaskPretendStub(flag *bool) RecipeTaskFactory {
	return func(paths ...string) *task.Task[layout.RecipeParams] {
		return &task.Task[layout.RecipeParams]{
			OnPrepare: func(params *layout.RecipeParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *layout.RecipeParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}
