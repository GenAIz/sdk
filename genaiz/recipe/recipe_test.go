package recipe

import (
	"bytes"
	"embed"
	"errors"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed all:test_recipes
	testRecipes     embed.FS
	testRecipesPath = "test_recipes"
)

func TestBook_FindRecipe(t *testing.T) {
	var expectedRecipe = "bash-example"
	var testBook = NewBook()
	var testRecipe, err = testBook.FindRecipe(expectedRecipe, TypeFunction)

	assert.NoError(t, err)
	assert.EqualValues(t, expectedRecipe, testRecipe.Name)
}

func TestBook_FindRecipeRedundant(t *testing.T) {
	var expectedRecipe = "bash-example"
	var testBook = NewBook()
	var testRecipe, err = testBook.FindRecipe(folderKey(TypeFunction, expectedRecipe), TypeFunction)

	assert.NoError(t, err)
	assert.EqualValues(t, expectedRecipe, testRecipe.Name)
}

func TestBook_FindRecipeVariation(t *testing.T) {
	var testRecipeName = "varied-example"
	var expectedRecipeName = testRecipeName + "-func"
	var testBook = NewBook()

	testBook.AddRecipe(&Recipe{
		Name: expectedRecipeName,
		Type: TypeFunction,
	})

	if testRecipe, err := testBook.FindRecipe(testRecipeName, TypeFunction); err == nil {
		assert.EqualValues(t, expectedRecipeName, testRecipe.Name)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestBook_FindVariationNotExisting(t *testing.T) {
	var testBook = NewBook()
	var _, err = testBook.FindVariation("project-x", TypeProject)

	assert.Error(t, err)
}

func TestBook_GetRecipe(t *testing.T) {
	var expectedRecipe = "bash-example"
	var testBook = NewBook()
	var testRecipe, err = testBook.GetRecipe(expectedRecipe, TypeFunction)

	assert.NoError(t, err)
	assert.EqualValues(t, expectedRecipe, testRecipe.Name)
}

func TestBook_GetRecipeNoFound(t *testing.T) {
	var testBook = NewBook()

	_, err := testBook.GetRecipe("_will_never_exist", TypeFunction)
	assert.Error(t, err)
}

func TestRecipe_Finish(t *testing.T) {
	var testDir = t.TempDir()
	var expectedInstance = "_will_never_exist"
	var testRecipe = &Recipe{
		Name:         "testRecipe_Finish",
		Type:         TypeFunction,
		PostCommands: []string{"ls {{.InstanceName}}"},
	}

	if err := testRecipe.Finish(testDir, expectedInstance, map[string]string{}); err != nil {
		assert.Contains(t, err.Error(), expectedInstance)
	} else {
		assert.Fail(t, "expected error")
	}
}

func TestRecipe_Init(t *testing.T) {
	var testDir = t.TempDir()
	var expectedInstance = "_will_never_exist"
	var testRecipe = &Recipe{
		Name:         "testRecipe_Init",
		Type:         TypeFunction,
		InitCommands: []string{"ls {{.InstanceName}}"},
	}

	if err := testRecipe.Init(testDir, expectedInstance, map[string]string{}); err != nil {
		assert.Contains(t, err.Error(), expectedInstance)
	} else {
		assert.Fail(t, "expected error")
	}
}

func TestRecipe_WriteFiles(t *testing.T) {
	var testDir = t.TempDir()

	if fd, err := os.CreateTemp(testDir, "genaiz-test-write-riles"); err == nil {
		var testArtifact = &Artifact{
			Name: fd.Name(),
		}
		var testRecipe = &Recipe{
			Name:      "testRecipe_WriteFilesParseError",
			Type:      TypeFunction,
			Artifacts: []*Artifact{testArtifact},
			parse: func(path string, t *template.Template) (*template.Template, error) {
				return t.Parse("{{.test}}")
			},
		}

		assert.NoError(t, testRecipe.WriteFiles(testDir, "instanceName", map[string]string{}))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestRecipe_WriteFilesExecuteError(t *testing.T) {
	var testDir = t.TempDir()
	var testArtifact = &Artifact{
		Name: "testArtifact",
	}
	var testRecipe = &Recipe{
		Name:      "testRecipe_WriteFilesParseError",
		Type:      TypeFunction,
		Artifacts: []*Artifact{testArtifact},
		parse: func(path string, t *template.Template) (*template.Template, error) {
			return t, nil
		},
	}

	assert.Error(t, testRecipe.WriteFiles(testDir, "instanceName", map[string]string{}))
}

func TestRecipe_WriteFilesParseError(t *testing.T) {
	var testDir = t.TempDir()
	var expectedError = errors.New("expected")
	var testArtifact = &Artifact{
		Name: "testArtifact",
	}
	var testRecipe = &Recipe{
		Name:      "testRecipe_WriteFilesParseError",
		Type:      TypeFunction,
		Artifacts: []*Artifact{testArtifact},
		parse: func(path string, t *template.Template) (*template.Template, error) {
			return nil, expectedError
		},
	}
	var err = testRecipe.WriteFiles(testDir, "instanceName", map[string]string{})

	assert.ErrorIs(t, expectedError, err)
}

func TestRecipe_WriteFilesPathError(t *testing.T) {
	var testDir = t.TempDir()
	var testArtifact = &Artifact{
		Name: "/_will_never_exist/testArtifact",
	}
	var testRecipe = &Recipe{
		Name:      "testRecipe_WriteFilesParseError",
		Type:      TypeFunction,
		Artifacts: []*Artifact{testArtifact},
		parse: func(path string, t *template.Template) (*template.Template, error) {
			return t.Parse("{{.test}}")
		},
	}
	var err = testRecipe.WriteFiles(testDir, "instanceName", map[string]string{})

	assert.ErrorContains(t, err, testArtifact.Name)
}

func TestNewBook(t *testing.T) {
	var expectedPath = t.TempDir()
	var testBook = NewBook(expectedPath)
	var variations = Variations[TypeFunction].GetVariations("bash-example")
	var allRecipeKeys = slices.Collect(maps.Keys(testBook.Recipes))

	for _, key := range variations {
		assert.Contains(t, allRecipeKeys, folderKey(TypeFunction, key))
	}

	assert.Contains(t, testBook.Paths, expectedPath)
}

func TestNewEmbedded(t *testing.T) {
	var err error
	var entries []fs.DirEntry

	if entries, err = embedded.ReadDir(embeddedPath); err == nil {
		for _, entry := range entries {
			if strings.EqualFold("bash_example", entry.Name()) {
				var testRecipe = NewEmbedded(entry, embeddedPath)
				var tpl = template.New("Dockerfile.tmpl")
				var dockerIndex = slices.IndexFunc(testRecipe.Artifacts, func(a *Artifact) bool {
					return strings.EqualFold(a.Name, "Dockerfile.tmpl")
				})
				var dockerPath = filepath.Join(embeddedPath, entry.Name(), "Dockerfile.tmpl")

				assert.EqualValues(t, "bash-example", testRecipe.Name)
				assert.EqualValues(t, len(testRecipe.Artifacts), 2)
				assert.True(t, slices.ContainsFunc(testRecipe.Artifacts, func(a *Artifact) bool {
					return strings.EqualFold(a.Name, "app.sh.tmpl")
				}))
				assert.True(t, dockerIndex > 0)
				_, err = testRecipe.parse(dockerPath, tpl)
				assert.NoError(t, err)
				return
			}
		}

		assert.Fail(t, "could not find bash_example")
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_embedded_recipes_bash_example_APP_SH(t *testing.T) {
	var testTemplate = template.New("app.sh.tmpl")
	var err error

	if testTemplate, err = testTemplate.ParseFS(embedded, "embedded_recipes/bash_example/app.sh.tmpl"); err == nil {
		var buffer = new(bytes.Buffer)
		var expectedValues = map[string]string{
			"InstanceName": "testInstance",
			"version":      "testVersion",
		}

		if err = testTemplate.Execute(io.Writer(buffer), expectedValues); err == nil {
			var content = buffer.String()

			assert.Contains(t, content, "Smart Function "+expectedValues["InstanceName"]+" version "+expectedValues["version"])
		}
	}

	assert.NoError(t, err)
}

func Test_embedded_recipes_bash_example_DOCKERFILE(t *testing.T) {
	var expectedBinPath = t.TempDir()
	var testTemplate = template.New("Dockerfile.tmpl")
	var err error

	if testTemplate, err = testTemplate.ParseFS(embedded, "embedded_recipes/bash_example/Dockerfile.tmpl"); err == nil {
		var buffer = new(bytes.Buffer)

		if err = testTemplate.Execute(io.Writer(buffer), map[string]string{"SF_BIN_PATH": expectedBinPath}); err == nil {
			var content = buffer.String()

			assert.Contains(t, content, "ADD app.sh "+expectedBinPath+"/")
			assert.Contains(t, content, "RUN chmod +x "+expectedBinPath+"/app.sh")
			assert.Contains(t, content, "CMD [\""+expectedBinPath+"/app.sh\"]")
		}
	}

	assert.NoError(t, err)
}

func Test_embedded_recipes_bash_example_RECIPE_YAML(t *testing.T) {
	var testPath = filepath.Join(embeddedPath, "bash_example", recipeFile)

	assert.NoError(t, testAllCommands(testPath))
}

func Test_executeCommand(t *testing.T) {
	var testCommand = "ls ."

	assert.NoError(t, executeCommand(testCommand))
}

func Test_executeCommandsEmptyCommand(t *testing.T) {
	var testMap = map[string]string{}

	assert.Error(t, executeCommands([]string{""}, "test", &testMap))
}

func Test_executeCommandsInvalidInstruction(t *testing.T) {
	var testInvalidTemplate = "{{template \"doesNotExist\"}}"
	var testMap = map[string]string{}

	assert.Error(t, executeCommands([]string{testInvalidTemplate}, "test", &testMap))
}

func Test_executeCommandsInvalidTemplate(t *testing.T) {
	var testInvalidTemplate = "{{it's so easy to make something invalid}}"
	var testMap = map[string]string{}

	assert.Error(t, executeCommands([]string{testInvalidTemplate}, "test", &testMap))
}

func Test_executeCommandsNilParams(t *testing.T) {
	var testInvalidTemplate = "{{it's so easy to make something invalid}}"

	assert.Panics(t, func() {
		_ = executeCommands([]string{testInvalidTemplate}, "test", nil)
	})
}

func Test_folderKey(t *testing.T) {
	var expectedKey = "key"
	var testKey = folderKey(TypeFunction, expectedKey)

	assert.Contains(t, testKey, TypeFunction)
	assert.Contains(t, testKey, expectedKey)
}

func Test_instanceParams(t *testing.T) {
	var expectedDest = "dest"
	var expectedName = "name"
	var testMap = instanceParams(expectedDest, expectedName)

	assert.EqualValues(t, expectedDest, testMap["PWD"])
	assert.EqualValues(t, expectedName, testMap["InstanceName"])
}

func Test_readRecipeDescriptorError(t *testing.T) {
	var entries []fs.DirEntry
	var err error

	if entries, err = testRecipes.ReadDir(testRecipesPath); err == nil {
		assert.True(t, len(entries) > 0)

		_, err = readRecipe(entries[0], testRecipesPath,
			func(path string, t *template.Template) (*template.Template, error) {
				return t, nil
			},
			func(s string) ([]os.DirEntry, error) {
				return nil, nil
			},
			func(s string) ([]byte, error) {
				if strings.HasSuffix(s, recipeFile) {
					return nil, errors.New("test error")
				}

				return nil, nil
			})
		assert.Error(t, err)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_readRecipeYamlParsingError(t *testing.T) {
	var entries []fs.DirEntry
	var err error

	if entries, err = testRecipes.ReadDir(testRecipesPath); err == nil {
		assert.True(t, len(entries) > 0)
		_, err = readRecipe(entries[0], testRecipesPath,
			func(path string, t *template.Template) (*template.Template, error) {
				return t, nil
			},
			testRecipes.ReadDir,
			testRecipes.ReadFile)
		assert.Error(t, err)
	} else {
		assert.Fail(t, err.Error())
	}
}

func testAllCommands(recipePath string) error {
	var recipeBytes []byte
	var err error

	if recipeBytes, err = embedded.ReadFile(recipePath); err == nil {
		var recipe *Recipe

		if err = yaml.Unmarshal(recipeBytes, &recipe); err == nil {
			for _, artifact := range recipe.Artifacts {
				for _, command := range artifact.Commands {
					if _, err = template.New("test").Parse(command); err != nil {
						return err
					}
				}
			}

			for _, command := range recipe.InitCommands {
				if _, err = template.New("test").Parse(command); err != nil {
					return err
				}
			}

			for _, command := range recipe.PostCommands {
				if _, err = template.New("test").Parse(command); err != nil {
					return err
				}
			}
		}
	}

	return err
}
