// Package recipe provides methods for producing smart function, solution and workflow hierarchies from known examples
package recipe

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
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

type Finder = func(registry Registry) (Recipe, error)
type Type = string

type Recipe interface {
	// Finish calls any post instructions the Recipe may require before it is completed. Returns an error if the wrapping up fails
	Finish(string, string, map[string]string) error

	// GetArtifacts provides artifacts initiated by the recipe
	GetArtifacts() []Artifact

	// GetName returns the name of the recipe
	GetName() string

	// Init creates the recipe layout using any initialization instructions it may have. Returns an error if the initialization fails
	Init(string, string, map[string]string) (map[string]string, error)

	// PrintFiles prints the files that would be created to STDOUT that would be created by WriteFiles
	PrintFiles(string, string, map[string]string)

	// PrintFinish prints the commands to STDOUT that would be executed by Finish, instead of executing them
	PrintFinish(string, string, map[string]string)

	// PrintInit prints the commands to STDOUT that would be executed by Init, instead of executing them
	PrintInit(string, string, map[string]string)

	// WriteFiles processes the collection of Artifact under the destination provided. Returns an error if file processing fails
	WriteFiles(string, string, map[string]string) error
}

type Registry interface {
	GetRecipe(string, Type) (Recipe, error)

	FindRecipe(string, Type) (Recipe, error)

	FindVariation(string, Type) (Recipe, error)
}

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
	Dest     string   // Dest is checked to see if we should omit this Artifact, left empty will cause the recipe to overwrite what was there previously
	Name     string   // Name of the artifact for the purpose of displaying
}

// registry represents a list of local paths to search for recipes and a cache of already known recipes
type registry struct {
	Paths   []string          // Paths in which to look for finding recipes
	Recipes map[string]Recipe // Known embedded Recipe are inscribed into the book, found recipes may be cached as well
}

// GetRecipe returns the recipe for the provided recipeName looking in the recipeType section of the book
func (r *registry) GetRecipe(recipeName string, recipeType Type) (Recipe, error) {
	var key = folderKey(recipeType, recipeName)
	var result = r.Recipes[key]

	if result == nil {
		return nil, errors.New("recipe not found")
	}

	return result, nil
}

// FindRecipe tries to find a recipe from its current Book entries
//
// TODO: have the function look through paths using baseName as a filename
func (r *registry) FindRecipe(baseName string, recipeType Type) (Recipe, error) {
	var result = r.Recipes[baseName]
	var err error

	if result == nil {
		result, err = r.GetRecipe(baseName, recipeType)
	}

	if result == nil {
		result, err = r.FindVariation(baseName, recipeType)
	}

	return result, err
}

// FindVariation tries to find a recipe from its current Book entries using a variation definition for the recipeType
func (r *registry) FindVariation(baseName string, recipeType Type) (Recipe, error) {
	var varies = Variations[recipeType]

	if varies != nil {
		var variations = varies.GetVariations(baseName)

		for _, variation := range variations {
			var key = folderKey(recipeType, variation)

			if result, ok := r.Recipes[key]; ok {
				return result, nil
			}
		}
	}

	return nil, errors.New("recipe not found")
}

// addRecipe adds a recipe with keyed variations to the Book.Recipes map
func (r *registry) addRecipe(rp *recipe) {
	r.Recipes[rp.getKey()] = rp

	for _, key := range rp.getKeyVariations() {
		r.Recipes[key] = rp
	}

	for _, key := range rp.getKeyAliases() {
		r.Recipes[key] = rp
	}
}

// recipe is a set of commands with a set of artifacts to write and optionally to process
type recipe struct {
	Name         string            // Name of the recipe, usually found in the filename if external with some variations
	Path         string            // Path of the recipe
	Type         Type              // Type of recipe, see recipe.Type
	Aliases      []string          // Aliases represent a list of strings with which one can refer to the recipe, these are not combined with variations
	Artifacts    []Artifact        // Artifacts contains File content and an optional command to execute after writing
	InitParams   map[string]string `yaml:"initParams"`   // InitParams provide initial configurations from the recipe
	InitCommands []string          `yaml:"initCommands"` // InitCommands should be executed before the Artifacts are processed with their own Command
	PostCommands []string          `yaml:"postCommands"` // PostCommands should be executed after the Artifacts are written locally

	parse parseFunction // parse provides the parseFunction for parsing the Recipe.Path
}

func (r *recipe) Finish(dest string, instanceName string, variables map[string]string) error {
	var params = instanceParams(dest, instanceName)

	maps.Copy(params, variables)
	return executeCommands(r.PostCommands, "postCommands", params)
}

func (r *recipe) GetArtifacts() []Artifact {
	return r.Artifacts
}

func (r *recipe) GetName() string {
	return r.Name
}

func (r *recipe) Init(dest string, instanceName string, variables map[string]string) (map[string]string, error) {
	var params = instanceParams(dest, instanceName)

	maps.Copy(params, variables)

	if err := executeCommands(r.InitCommands, "initCommands", params); err != nil {
		return nil, err
	}

	return r.InitParams, nil
}

func (r *recipe) PrintFiles(dest string, instanceName string, variables map[string]string) {
	var params = instanceParams(dest, instanceName)

	maps.Copy(params, variables)

	for _, artifact := range r.Artifacts {
		var artifactPath = filepath.Join(r.Path, artifact.Name)

		fmt.Printf("touch %s\n", artifactPath)

		if err := printCommands(artifact.Commands, instanceName, params); err != nil {
			fmt.Printf("%s\n", err.Error())
			break
		}
	}
}

func (r *recipe) PrintFinish(dest string, instanceName string, variables map[string]string) {
	var params = instanceParams(dest, instanceName)

	maps.Copy(params, variables)

	if err := printCommands(r.PostCommands, "postCommands", params); err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}

func (r *recipe) PrintInit(dest string, instanceName string, variables map[string]string) {
	var params = instanceParams(dest, instanceName)

	maps.Copy(params, variables)

	if err := printCommands(r.InitCommands, "initCommands", params); err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}

func (r *recipe) WriteFiles(dest string, instanceName string, variables map[string]string) error {
	var params = instanceParams(dest, instanceName)
	var err error

	maps.Copy(params, variables)

	for _, artifact := range r.Artifacts {
		var processArtifact = true

		if artifact.Dest != "" {
			if _, err = os.Stat(filepath.Join(r.Path, artifact.Dest)); err == nil {
				processArtifact = false
			}
		}

		if processArtifact {
			if err = r.writeFile(&artifact, instanceName, params, variables); err != nil {
				break
			}
		}
	}

	return err
}

func (r *recipe) getKey() string {
	return folderKey(r.Type, r.Name)
}

func (r *recipe) getKeyAliases() []string {
	var result []string

	for _, alias := range r.Aliases {
		result = append(result, folderKey(r.Type, alias))
	}

	return result
}

func (r *recipe) getKeyVariations() []string {
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

func (r *recipe) writeFile(artifact *Artifact, instanceName string, params, variables map[string]string) error {
	var tpl = template.New(artifact.Name)
	var artifactPath = filepath.Join(r.Path, artifact.Name)
	var output = new(bytes.Buffer)
	var err error

	if tpl, err = r.parse(artifactPath, tpl); err == nil {
		if err = tpl.Execute(io.Writer(output), params); err == nil {
			if err = os.WriteFile(artifact.Name, output.Bytes(), 0640); err == nil {
				err = executeCommands(artifact.Commands, instanceName, variables)
			}
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

// NewRegistry creates a new Recipe Registry based on the provided local paths. The Book will also be initialized with all embedded recipes under the genaiz executable.
func NewRegistry(paths ...string) Registry {
	return newRegistry(paths...)
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

func executeCommands(commands []string, name string, params map[string]string) error {
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

func newEmbedded(entry fs.DirEntry, parentPath string) *recipe {
	var result, err = readRecipe(entry, parentPath,
		func(path string, t *template.Template) (*template.Template, error) {
			return t.ParseFS(embedded, path)
		},
		embedded.ReadDir,
		embedded.ReadFile)

	// embedded recipes can not contain errors, panic, it's a bug
	panicz.PanicIfError(err)
	return result
}

func newRegistry(paths ...string) *registry {
	var result = &registry{
		Paths:   paths,
		Recipes: map[string]Recipe{},
	}
	var entries []fs.DirEntry
	var err error

	if entries, err = embedded.ReadDir(embeddedPath); entries != nil {
		for _, recipeEntry := range entries {
			var embRecipe = newEmbedded(recipeEntry, embeddedPath)

			result.addRecipe(embRecipe)
		}
	}

	panicz.PanicIfError(err)
	return result
}

func printCommands(commands []string, name string, params map[string]string) error {
	var err error

	for _, cmd := range commands {
		var cmdBuffer = new(bytes.Buffer)
		var tpl = template.New(name)

		if tpl, err = tpl.Parse(cmd); err == nil {
			if err = tpl.Execute(io.Writer(cmdBuffer), params); err == nil {
				fmt.Printf("%s\n", cmdBuffer.String())
			}

			if err != nil {
				break
			}
		}
	}

	return err
}

func readRecipe(entry fs.DirEntry, parentPath string,
	parsingWith parseFunction,
	readingDirWith readDirFunction,
	readingFilesWith readFileFunction) (*recipe, error) {
	var basePath = filepath.Join(parentPath, entry.Name())
	var entryDescriptor = filepath.Join(basePath, recipeFile)
	var result = &recipe{
		Path:  basePath,
		parse: parsingWith,
	}
	var desc []byte
	var err error

	if desc, err = readingFilesWith(entryDescriptor); err == nil {
		if err = yaml.Unmarshal(desc, result); err == nil {
			var artifacts = mapz.Mapped(result.Artifacts, func(a Artifact) string { return a.Name })
			var files, _ = readingDirWith(basePath)

			for _, file := range files {
				if !strings.Contains(file.Name(), recipeFile) {
					var filePath = filepath.Join(basePath, file.Name())

					if artifact, ok := artifacts[file.Name()]; ok {
						artifact.File, err = readingFilesWith(filePath)
					}
				}
			}
		}
	}

	return result, err
}
