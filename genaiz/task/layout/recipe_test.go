package layout

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/recipe"
	"genaiz.com/genaiz/task"
)

type stubRecipe struct {
	artifacts            []recipe.Artifact
	finishDest           string
	finishError          error
	finishName           string
	finishVariables      map[string]string
	initDest             string
	initError            error
	initName             string
	initParams           map[string]string
	initVariables        map[string]string
	printFilesDest       string
	printFilesName       string
	printFilesVariables  map[string]string
	printFinishDest      string
	printFinishName      string
	printFinishVariables map[string]string
	printInitDest        string
	printInitName        string
	printInitVariables   map[string]string
	name                 string
	writeDest            string
	writeError           error
	writeName            string
	writeVariables       map[string]string
}

func (sr *stubRecipe) Finish(dest string, instanceName string, variables map[string]string) error {
	sr.finishDest = dest
	sr.finishName = instanceName
	sr.finishVariables = variables
	return sr.finishError
}

func (sr *stubRecipe) GetArtifacts() []recipe.Artifact {
	return sr.artifacts
}

func (sr *stubRecipe) GetName() string {
	return sr.name
}

func (sr *stubRecipe) Init(dest string, instanceName string, variables map[string]string) (map[string]string, error) {
	sr.initDest = dest
	sr.initName = instanceName
	sr.initVariables = variables
	return sr.initParams, sr.initError
}

func (sr *stubRecipe) PrintFiles(dest string, instanceName string, variables map[string]string) {
	sr.printFilesDest = dest
	sr.printFilesName = instanceName
	sr.printFilesVariables = variables
}

func (sr *stubRecipe) PrintFinish(dest string, instanceName string, variables map[string]string) {
	sr.printFinishDest = dest
	sr.printFinishName = instanceName
	sr.printFinishVariables = variables
}

func (sr *stubRecipe) PrintInit(dest string, instanceName string, variables map[string]string) {
	sr.printInitDest = dest
	sr.printInitName = instanceName
	sr.printInitVariables = variables
}

func (sr *stubRecipe) WriteFiles(dest string, instanceName string, variables map[string]string) error {
	sr.writeDest = dest
	sr.writeName = instanceName
	sr.writeVariables = variables
	return sr.writeError
}

type stubRegistry struct {
	recipePath     string
	recipeName     string
	recipeType     recipe.Type
	registryAdds   []recipe.Recipe
	registryError  error
	registryRecipe recipe.Recipe
	variationBase  string
}

func (sr *stubRegistry) GetRecipe(path string, t recipe.Type) (recipe.Recipe, error) {
	sr.recipePath = path
	sr.recipeType = t
	return sr.registryRecipe, sr.registryError
}

func (sr *stubRegistry) FindRecipe(name string, t recipe.Type) (recipe.Recipe, error) {
	sr.recipeName = name
	sr.recipeType = t
	return sr.registryRecipe, sr.registryError
}

func (sr *stubRegistry) FindVariation(baseName string, t recipe.Type) (recipe.Recipe, error) {
	sr.variationBase = baseName
	sr.recipeType = t
	return sr.registryRecipe, sr.registryError
}

func TestNewRecipeTask(t *testing.T) {
	var testTask = NewRecipeTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestRecipeParams_getRecipeFinder(t *testing.T) {
	var testParams = &RecipeParams{
		Name: "paramName",
		Type: "paramType",
	}
	var testFinder = testParams.getRecipeFinder("")
	var testRegistry = &stubRegistry{
		registryRecipe: &stubRecipe{},
	}
	var expectedPath = "path"

	actual, err := testFinder(testRegistry)
	assert.NoError(t, err)
	assert.Same(t, testRegistry.registryRecipe, actual)
	assert.Empty(t, testRegistry.recipePath)
	assert.Equal(t, testParams.Name, testRegistry.recipeName)
	testFinder = testParams.getRecipeFinder(expectedPath)
	testRegistry.recipeName = ""
	actual, err = testFinder(testRegistry)
	assert.Same(t, testRegistry.registryRecipe, actual)
	assert.Equal(t, expectedPath, testRegistry.recipePath)
	assert.Empty(t, testRegistry.recipeName)
}

func Test_handleLayoutRecipeContext(t *testing.T) {
	var testDir = t.TempDir()
	var testParams = &RecipeParams{}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testRegistry = &stubRegistry{
		registryRecipe: &stubRecipe{
			name: "expectedRecipe",
			artifacts: []recipe.Artifact{
				{
					Name: filepath.Join(testDir, "artifact"),
				},
			},
		},
	}

	assert.NoError(t, handleLayoutRecipeContext(testRegistry, testParams, testState))
	assert.Equal(t, testRegistry.registryRecipe.GetName(), testState.Output)
}

func Test_handleLayoutRecipeContext_ArtifactWarning(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "artifact.txt")

	if fd, err := os.Create(testFile); err == nil {
		var testParams = &RecipeParams{}
		var testState = &task.State{Logger: logrus.New()}
		var testHook = test.NewLocal(testState.Logger)
		var testRegistry = &stubRegistry{
			registryRecipe: &stubRecipe{
				artifacts: []recipe.Artifact{
					{
						Name: testFile,
					},
				},
			},
		}

		filez.CloseSilently(fd)
		assert.NoError(t, handleLayoutRecipeContext(testRegistry, testParams, testState))
		assert.NotEmpty(t, testHook.Entries)
		index := slices.IndexFunc(testHook.Entries, func(entry logrus.Entry) bool {
			return entry.Level == logrus.WarnLevel
		})
		assert.True(t, index >= 0)
		assert.Contains(t, testHook.Entries[index].Message, testFile)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLayoutRecipeContext_FinderError(t *testing.T) {
	var testParams = &RecipeParams{}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testRegistry = &stubRegistry{
		registryError: errors.New("expected"),
	}

	assert.ErrorIs(t, handleLayoutRecipeContext(testRegistry, testParams, testState), testRegistry.registryError)
}

func Test_handleLayoutRecipeCreate(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &RecipeParams{}
	var testRecipe = &stubRecipe{
		name: "expectedName",
	}
	var testRegistry = &stubRegistry{
		registryRecipe: testRecipe,
	}

	assert.NoError(t, handleLayoutRecipeCreate(testRegistry, testParams, testState))
	assert.Equal(t, 1, len(testState.Reports))
	assert.Contains(t, testState.Reports[0], testRecipe.name)
}

func Test_handleLayoutRecipeCreate_FinderError(t *testing.T) {
	var testParams = &RecipeParams{}
	var testState = &task.State{}
	var testRegistry = &stubRegistry{
		registryError: errors.New("expected"),
	}

	assert.ErrorIs(t, handleLayoutRecipeCreate(testRegistry, testParams, testState), testRegistry.registryError)
}

func Test_handleLayoutRecipeCreate_FinishError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &RecipeParams{}
	var testRegistry = &stubRegistry{
		registryRecipe: &stubRecipe{
			finishError: expectedError,
		},
	}

	assert.ErrorIs(t, handleLayoutRecipeCreate(testRegistry, testParams, testState), expectedError)
}

func Test_handleLayoutRecipeCreate_InitError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &RecipeParams{}
	var testRegistry = &stubRegistry{
		registryRecipe: &stubRecipe{
			initError: expectedError,
		},
	}

	assert.ErrorIs(t, handleLayoutRecipeCreate(testRegistry, testParams, testState), expectedError)
}

func Test_handleLayoutRecipeCreate_WriteError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &RecipeParams{}
	var testRegistry = &stubRegistry{
		registryRecipe: &stubRecipe{
			writeError: expectedError,
		},
	}

	assert.ErrorIs(t, handleLayoutRecipeCreate(testRegistry, testParams, testState), expectedError)
}

func Test_handleLayoutRecipePretend(t *testing.T) {
	var testDir = t.TempDir()
	var expectedOptionKey = "optionKey"
	var expectedOptionValue = "optionValue"
	var expectedParamKey = "paramKey"
	var expectedParamValue = "paramValue"
	var testParams = &RecipeParams{
		InstanceName: "expectedName",
		Options: map[string]string{
			expectedOptionKey: expectedOptionValue,
		},
		Parameters: map[string]string{
			expectedParamKey: expectedParamValue,
		},
	}
	var testState = &task.State{Logger: logrus.New()}
	var testRecipe = &stubRecipe{}
	var testRegistry = &stubRegistry{
		registryRecipe: testRecipe,
	}

	t.Chdir(testDir)
	assert.NoError(t, handleLayoutRecipePretend(testRegistry, testParams, testState))
	assert.Equal(t, testDir, testRecipe.printFilesDest)
	assert.Equal(t, testDir, testRecipe.printFinishDest)
	assert.Equal(t, testDir, testRecipe.printInitDest)
	assert.Equal(t, testParams.InstanceName, testRecipe.printFilesName)
	assert.Equal(t, testParams.InstanceName, testRecipe.printFinishName)
	assert.Equal(t, testParams.InstanceName, testRecipe.printInitName)
	assert.Equal(t, expectedOptionValue, testRecipe.printFilesVariables[expectedOptionKey])
	assert.Equal(t, expectedOptionValue, testRecipe.printFinishVariables[expectedOptionKey])
	assert.Equal(t, expectedOptionValue, testRecipe.printInitVariables[expectedOptionKey])
	assert.Equal(t, expectedParamValue, testRecipe.printFilesVariables[expectedParamKey])
	assert.Equal(t, expectedParamValue, testRecipe.printFinishVariables[expectedParamKey])
	assert.Equal(t, expectedParamValue, testRecipe.printInitVariables[expectedParamKey])
}

func Test_handleLayoutRecipePretend_FinderError(t *testing.T) {
	var testParams = &RecipeParams{}
	var testState = &task.State{}
	var testRegistry = &stubRegistry{
		registryError: errors.New("expected"),
	}

	assert.ErrorIs(t, handleLayoutRecipePretend(testRegistry, testParams, testState), testRegistry.registryError)
}
