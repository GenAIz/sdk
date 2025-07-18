package layout

import (
	"maps"
	"os"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/recipe"
	"genaiz.com/genaiz/task"
)

type RecipeParams struct {
	Name         string            // Name of the recipe or a string contained in its name, the recipe.Book will try its best to find a match
	InstanceName string            // InstanceName refers to the name of the entity created with the recipe
	Type         recipe.Type       // Type dictates in which recipe.Book paths and variations the layout creator will look for the Name
	Destination  string            // Destination is where the recipe is created, usually matching CreateParams.FolderPath
	Options      map[string]string // Options is a map of the keys used in configuration files, the keys may be used in templates
	Parameters   map[string]string // Parameters is a map of the parameters used on the command line, the keys may be used in templates
}

func NewRecipeTask(paths ...string) *task.Task[RecipeParams] {
	var book = recipe.NewBook(paths...)

	return &task.Task[RecipeParams]{
		Name:       "layout-recipe",
		OnPrepare:  lang.Assists(book, handleLayoutRecipeContext),
		OnComplete: lang.Assists(book, handleLayoutRecipeCreate),
		OnPretend:  lang.Assists(book, handleLayoutRecipePretend),
	}
}

func getRecipeFinder(params *RecipeParams, state *task.State) func(b *recipe.Book) (*recipe.Recipe, error) {
	if state.Output != "" {
		return func(b *recipe.Book) (*recipe.Recipe, error) {
			return b.GetRecipe(params.Name, params.Type)
		}
	}

	return func(b *recipe.Book) (*recipe.Recipe, error) {
		return b.FindRecipe(params.Name, params.Type)
	}
}

func handleLayoutRecipeContext(book *recipe.Book, params *RecipeParams, state *task.State) error {
	var layoutRecipe *recipe.Recipe
	var err error

	state.Logger.Debugf("Finding a recipe for [%s] of type [%s]", params.Name, params.Type)

	if layoutRecipe, err = book.FindRecipe(params.Name, params.Type); layoutRecipe != nil {
		state.Output = layoutRecipe.Name

		for _, artifact := range layoutRecipe.Artifacts {
			var fd *os.File

			if fd, err = os.OpenFile(artifact.Name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0640); fd != nil {
				filez.CloseSilently(fd)
			} else {
				state.Logger.Errorf("Can not write file [%s] to destination [%s]: %s", artifact.Name, params.Destination, err)
				break
			}
		}
	}

	return err
}

func handleLayoutRecipeCreate(book *recipe.Book, params *RecipeParams, state *task.State) error {
	var findRecipe = getRecipeFinder(params, state)
	var layoutRecipe *recipe.Recipe
	var err error

	if layoutRecipe, err = findRecipe(book); layoutRecipe != nil {
		var allVariables = map[string]string{}
		var currentFolder, _ = os.Getwd()

		maps.Copy(allVariables, params.Options)
		maps.Copy(allVariables, params.Parameters)
		state.Logger.Debugf("Initiating recipe [%s] on %s type for instance [%s]", params.Name, params.Type, params.InstanceName)

		if err = layoutRecipe.Init(currentFolder, params.InstanceName, allVariables); err == nil {
			state.Logger.Debugf("Writing recipe [%d] files", len(layoutRecipe.Artifacts)-1)

			if err = layoutRecipe.WriteFiles(currentFolder, params.InstanceName, allVariables); err == nil {
				state.Logger.Debugf("Completing recipe instance [%s]", params.InstanceName)
				err = layoutRecipe.Finish(currentFolder, params.InstanceName, allVariables)

				if err != nil {
					state.Logger.Errorf("Could not complete construction of recipe [%s] : %s", params.Name, err)
				}
			} else {
				state.Logger.Errorf("Could not write all files on recipe instance [%s]: %s", params.InstanceName, err)
			}
		}
	}

	return err
}

func handleLayoutRecipePretend(book *recipe.Book, params *RecipeParams, state *task.State) error {
	return nil
}
