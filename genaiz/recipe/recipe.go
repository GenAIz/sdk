// Package recipe provides methods for producing smart function, solution and workflow hierarchies from known examples
package recipe

import (
	"bytes"
	"embed"
	"errors"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/mapz"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/lang/enumz"
)

type parseFunction func(path string, t *template.Template) (*template.Template, error)
type readFileFunction func(string) ([]byte, error)
type readDirFunction func(string) ([]os.DirEntry, error)

type Type = string

type Varies interface {
	GetVariations(name string) []string

	IsaVariety(name string) bool
}

var (
	//go:embed all:embedded_recipes
	embedded     embed.FS
	embeddedPath = "embedded_recipes"

	//go:embed recipe.json
	schema     embed.FS
	recipeFile = "recipe.yaml"

	TypeFunction = "function" // a single function with no children
	TypeModule   = "module"   // a module belongs to a project under an IDE
	TypeProject  = "project"  // a project belongs to a specific IDE (ex: Idea, Xcode, etc..)
	TypeSolution = "solution" // a solution is a publishing container for functions and workflows
	TypeWorkflow = "workflow" // a workflow is a composition of several functions

	Types = enumz.NewEnumType(TypeFunction, TypeModule, TypeProject, TypeSolution, TypeWorkflow)

	Variations = map[Type]Varies{
		TypeFunction: &Variation{
			Fixes: []string{"func", "fn", "sf"},
		},
		TypeSolution: &Variation{
			Fixes: []string{"sol", "sn", "so"},
		},
		TypeWorkflow: &Variation{
			Fixes: []string{"flow", "wf"},
		},
	}
)

// Artifact is a tuple for a File with a Command
type Artifact struct {
	File     []byte   // The content of the File, unprocessed, this will be processed with the Go template engine
	Commands []string // An optional Command to run after the file has been processed
	Name     string   // Name of the artifact for the purpose of displaying
}

// Book represents a list of local paths to search for recipes and a cache of already known recipes
type Book struct {
	Paths   []string           // Paths in which to look for finding recipes
	Recipes map[string]*Recipe // Known embedded Recipe are inscribed into the book, found recipes may be cached as well
}

// AddRecipe adds a recipe with keyed variations to the Book.Recipes map
func (b *Book) AddRecipe(recipe *Recipe) {
	b.Recipes[recipe.GetKey()] = recipe

	for _, key := range recipe.GetKeyVariations() {
		b.Recipes[key] = recipe
	}

	for _, key := range recipe.GetKeyAliases() {
		b.Recipes[key] = recipe
	}
}

// GetRecipe returns the recipe for the provided recipeName looking in the recipeType section of the book
func (b *Book) GetRecipe(recipeName string, recipeType Type) (*Recipe, error) {
	var key = folderKey(recipeType, recipeName)
	var result = b.Recipes[key]

	if result == nil {
		return nil, errors.New("recipe not found")
	}

	return result, nil
}

// FindRecipe tries to find a recipe from its current Book entries
//
// TODO: have the function look through paths using baseName as a filename
func (b *Book) FindRecipe(baseName string, recipeType Type) (*Recipe, error) {
	var recipe = b.Recipes[baseName]
	var err error

	if recipe == nil {
		recipe, err = b.GetRecipe(baseName, recipeType)
	}

	if recipe == nil {
		recipe, err = b.FindVariation(baseName, recipeType)
	}

	return recipe, err
}

// FindVariation tries to find a recipe from its current Book entries using a variation definition for the recipeType
func (b *Book) FindVariation(baseName string, recipeType Type) (*Recipe, error) {
	var varies = Variations[recipeType]

	if varies != nil {
		var variations = varies.GetVariations(baseName)
		var recipe *Recipe

		for _, variation := range variations {
			var key = folderKey(recipeType, variation)

			if recipe = b.Recipes[key]; recipe != nil {
				return recipe, nil
			}
		}
	}

	return nil, errors.New("recipe not found")
}

// Recipe is a set of commands with a set of artifacts to write and optionally to process
type Recipe struct {
	Name         string      // Name of the recipe, usually found in the filename if external with some variations
	Path         string      // Path of the recipe
	Type         Type        // Type of recipe, see recipe.Type
	Aliases      []string    // Aliases represent a list of strings with which one can refer to the recipe, these are not combined with variations
	Artifacts    []*Artifact // Artifacts contains File content and an optional command to execute after writing
	InitCommands []string    `yaml:"initCommands"` // InitCommand should be executed before the Artifacts are processed with their own Command
	PostCommands []string    `yaml:"postCommands"` // PostCommand should be executed after the Artifacts are written locally

	parse parseFunction // parse provides the parseFunction for parsing the Recipe.Path
}

// Finish calls the Recipe.PostCommand. Returns an error if the command fails.
func (r *Recipe) Finish(dest string, instanceName string, variables map[string]string) error {
	var params = instanceParams(dest, instanceName)

	maps.Copy(params, variables)
	return executeCommands(r.PostCommands, "postCommands", &params)
}

// GetKey creates a key for the recipe for classifying in a Book or in a map[string]*Recipe
func (r *Recipe) GetKey() string {
	return folderKey(r.Type, r.Name)
}

// GetKeyAliases returns a slice of all the recipe key aliases supported by this recipe
func (r *Recipe) GetKeyAliases() []string {
	var result []string

	for _, alias := range r.Aliases {
		result = append(result, folderKey(r.Type, alias))
	}

	return result
}

// GetKeyVariations returns a slice of all the recipe key variations supported by this recipe
func (r *Recipe) GetKeyVariations() []string {
	var result []string

	if v, _ := Variations[r.Type]; v != nil {
		if !v.IsaVariety(r.Name) {
			for _, variation := range v.GetVariations(r.Name) {
				result = append(result, folderKey(r.Type, variation))
			}
		}
	}

	return result
}

// Init creates the recipe layout using the Recipe.InitCommand for the destination provided and the instanceName specified. Returns an error if the command fails.
func (r *Recipe) Init(dest string, instanceName string, variables map[string]string) error {
	var params = instanceParams(dest, instanceName)

	maps.Copy(params, variables)
	return executeCommands(r.InitCommands, "initCommands", &params)
}

// WriteFiles processes all the Recipe.Artifacts under the destination provided. Returns an error if file processing fails.
func (r *Recipe) WriteFiles(dest string, instanceName string, variables map[string]string) error {
	var params = instanceParams(dest, instanceName)
	var err error

	maps.Copy(params, variables)

	for _, artifact := range r.Artifacts {
		var tpl = template.New(artifact.Name)
		var artifactPath = filepath.Join(r.Path, artifact.Name)
		var output = new(bytes.Buffer)

		if tpl, err = r.parse(artifactPath, tpl); err == nil {
			if err = tpl.Execute(io.Writer(output), params); err == nil {
				if err = os.WriteFile(artifact.Name, output.Bytes(), 0640); err == nil {
					err = executeCommands(artifact.Commands, instanceName, &variables)
				}
			}
		}

		if err != nil {
			break
		}
	}

	return err
}

type Variation struct {
	Fixes []string
}

func (v Variation) GetVariations(name string) []string {
	var result []string

	for _, fx := range v.Fixes {
		result = append(result, name+"-"+fx, name+"_"+fx, name+"."+fx)
		result = append(result, fx+"-"+name, fx+"_"+name, fx+"."+name)
	}

	return result
}

func (v Variation) IsaVariety(name string) bool {
	return slices.IndexFunc(v.Fixes, func(s string) bool {
		return strings.Contains(strings.ToLower(name), s)
	}) > -1
}

// NewBook creates a new Recipe Book based on the provided local paths. The Book will also be initialized with all embedded recipes under the genaiz executable.
func NewBook(paths ...string) *Book {
	var result = Book{
		Paths:   paths,
		Recipes: map[string]*Recipe{},
	}
	var entries []fs.DirEntry
	var err error

	if entries, err = embedded.ReadDir(embeddedPath); entries != nil {
		for _, recipeEntry := range entries {
			var recipe = NewEmbedded(recipeEntry, embeddedPath)

			result.AddRecipe(recipe)
		}
	}

	panicz.PanicIfError(err)
	return &result
}

func NewEmbedded(entry fs.DirEntry, parentPath string) *Recipe {
	var recipe, err = readRecipe(entry, parentPath,
		func(path string, t *template.Template) (*template.Template, error) {
			return t.ParseFS(embedded, path)
		},
		embedded.ReadDir,
		embedded.ReadFile)

	// embedded recipes can not contain errors, panic, it's a bug
	panicz.PanicIfError(err)
	return recipe
}

func executeCommand(cmd string) error {
	var tokens = strings.Split(cmd, " ")
	// GO ism #1 the exec.Command constructor does not like taking just an array, it needs tokens split and re-merged
	// to avoid duplicates, missing path checks and disappearing ARG[0], it doesn't like treating ARG[0] like everyone else
	// and needs a copy in Command.Path
	var command = exec.Command(tokens[0])
	var size = len(tokens)
	var errorBuffer = new(bytes.Buffer)
	var err error

	if size > 1 {
		command.Args = append(command.Args, tokens[1:]...)
	}

	// GO ism #2 the exec.Command does not manufacture an error with the provided error message from the command Stderr,
	// so we have to retrieve it manually by overriding it with a write buffer
	command.Stderr = io.Writer(errorBuffer)

	if _, err = command.Output(); err != nil {
		err = errors.New(errorBuffer.String())
	}

	return err
}

func executeCommands(commands []string, name string, params *map[string]string) error {
	var err error

	panicz.RequiresNotNil("params", params)

	for _, cmd := range commands {
		var cmdBuffer = new(bytes.Buffer)
		var tpl = template.New(name)

		if tpl, err = tpl.Parse(cmd); err == nil {
			if err = tpl.Execute(io.Writer(cmdBuffer), params); err == nil {
				err = executeCommand(cmdBuffer.String())
			}

			if err != nil {
				break
			}
		}
	}

	return err
}

func folderKey(recipeType Type, name string) string {
	return recipeType + "/" + name
}

func instanceParams(dest string, name string) map[string]string {
	return map[string]string{
		"PWD":          dest,
		"InstanceName": name,
	}
}

func readRecipe(entry fs.DirEntry, parentPath string,
	parsingWith parseFunction,
	readingDirWith readDirFunction,
	readingFilesWith readFileFunction) (*Recipe, error) {
	var basePath = filepath.Join(parentPath, entry.Name())
	var entryDescriptor = filepath.Join(basePath, recipeFile)
	var result = &Recipe{
		Path:  basePath,
		parse: parsingWith,
	}
	var desc []byte
	var err error

	if desc, err = readingFilesWith(entryDescriptor); err == nil {
		if err = yaml.Unmarshal(desc, result); err == nil {
			var artifacts = mapz.Mapped(result.Artifacts, func(a *Artifact) string { return a.Name })
			var files, _ = readingDirWith(basePath)

			for _, file := range files {
				if !strings.Contains(file.Name(), recipeFile) {
					var filePath = filepath.Join(basePath, file.Name())
					var artifact = artifacts[file.Name()]

					if artifact != nil {
						artifact.File, err = readingFilesWith(filePath)
					}
				}
			}
		}
	}

	return result, err
}
