package layout

import (
	"maps"
	"os"

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

func (rp RecipeParams) getRecipeFinder(path string) recipe.Finder {
	if path == "" {
		return func(r recipe.Registry) (recipe.Recipe, error) {
			return r.FindRecipe(rp.Name, rp.Type)
		}
	}

	return func(r recipe.Registry) (recipe.Recipe, error) {
		return r.GetRecipe(path, rp.Type)
	}
}

func NewRecipeTask(paths ...string) *task.Task[RecipeParams] {
	var registry = recipe.NewRegistry(paths...)

	return &task.Task[RecipeParams]{
		Name:       "layout-recipe",
		OnPrepare:  lang.Assists(registry, handleLayoutRecipeContext),
		OnComplete: lang.Assists(registry, handleLayoutRecipeCreate),
		OnPretend:  lang.Assists(registry, handleLayoutRecipePretend),
	}
}

func handleLayoutRecipeContext(registry recipe.Registry, params *RecipeParams, state *task.State) error {
	var layoutRecipe recipe.Recipe
	var err error

	state.Logger.Debugf("Finding a recipe for [%s] of type [%s]", params.Name, params.Type)

	if layoutRecipe, err = registry.FindRecipe(params.Name, params.Type); layoutRecipe != nil {
		var fd os.FileInfo

		state.Output = layoutRecipe.GetName()

		for _, artifact := range layoutRecipe.GetArtifacts() {
			fd, err = os.Stat(artifact.Name)

			if fd != nil {
				state.Logger.Warningf("Artifact [%s] already exists and will be overwritten", artifact.Name)
			} else {
				err = nil
			}
		}
	}

	return err
}

func handleLayoutRecipeCreate(registry recipe.Registry, params *RecipeParams, state *task.State) error {
	var findRecipe = params.getRecipeFinder(state.Output)
	var layoutRecipe recipe.Recipe
	var err error

	if layoutRecipe, err = findRecipe(registry); err == nil {
		var allVariables = map[string]string{}
		var recipeParams = map[string]string{}
		var currentFolder, _ = os.Getwd()

		maps.Copy(allVariables, params.Options)
		maps.Copy(allVariables, params.Parameters)
		state.Logger.Debugf("Initiating recipe [%s] on %s type for instance [%s]", params.Name, params.Type, params.InstanceName)

		if recipeParams, err = layoutRecipe.Init(currentFolder, params.InstanceName, allVariables); err == nil {
			state.Logger.Debugf("Writing recipe [%d] files", len(layoutRecipe.GetArtifacts())-1)

			if err = layoutRecipe.WriteFiles(currentFolder, params.InstanceName, allVariables); err == nil {
				var initState = NewInitState(state)

				initState.AddParams(recipeParams)
				initState.InitVars(allVariables)
				state.Logger.Debugf("Completing recipe instance [%s]", params.InstanceName)

				if err = layoutRecipe.Finish(currentFolder, params.InstanceName, allVariables); err == nil {
					state.Output = ""
					state.Reportf("Constructed recipe %s", layoutRecipe.GetName())
				} else {
					state.Logger.Errorf("Could not complete construction of recipe [%s] : %s", params.Name, err)
				}
			} else {
				state.Logger.Errorf("Could not write all files on recipe instance [%s]: %s", params.InstanceName, err)
			}
		}
	}

	return err
}

func handleLayoutRecipePretend(registry recipe.Registry, params *RecipeParams, state *task.State) error {
	var findRecipe = params.getRecipeFinder(state.Output)
	var layoutRecipe recipe.Recipe
	var err error

	if layoutRecipe, err = findRecipe(registry); err == nil {
		var allVariables = map[string]string{}
		var currentFolder, _ = os.Getwd()

		maps.Copy(allVariables, params.Options)
		maps.Copy(allVariables, params.Parameters)
		state.Logger.Debugf("Pretending recipe [%s] on %s type for instance [%s]", params.Name, params.Type, params.InstanceName)
		layoutRecipe.PrintInit(currentFolder, params.InstanceName, allVariables)
		layoutRecipe.PrintFiles(currentFolder, params.InstanceName, allVariables)
		layoutRecipe.PrintFinish(currentFolder, params.InstanceName, allVariables)
		return nil
	}

	return err
}
