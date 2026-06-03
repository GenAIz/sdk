## Workspace Create

```
genaiz ws create NAME --account=[<user>@][host] \
    --description=DESC --disallow-rc --visibility=[PRIVATE|ORG] \
    --json 
```

Workspace create is used to create a workspace on an account from which a user can eventually deploy Solution Workflows. 

### NAME

The name of the workspace is required by the create command, but is not mandatory for a workspace to exist.

* if the name of the workspace is missing, the command will fail printing usage definition
* if the value of name is not a valid [name](../index.md#name), the command will fail with error: `Error: workspace name must not exceed 255 characters`

### account

The account under which to create the workspace.

* if the value of the account does not evaluate to a Host address, the command returns an error: `Error could not elect a session`
*account values will auto-complete if the shell completion script is sourced:

```bash
genaiz completion bash > ~/.config/genaiz/completion.sh
source ~/.config/genaiz/completion.sh
```

### description

The description option of the command is used to populate the `workspace.description` field.

* if the value of description is empty, it will be left empty
* if the name does not resolve to a valid description string (see [description validity](../index.md#description)), the command will return an error with the key of the field and the shortened invalid value: `Error: value [...] for option [workspace.create.description] is invalid`

### disallow-rc

Disallowing release candidates, means the workspace will not be able to use Solutions with -rc-N version suffix. Only allowing released stable versions.

* by default, all workspace creation with the create command allow Release Candidates

### visibility

Visibility affects which users can access and manage the Workspace. A `PRIVATE` workspace can only be managed by its owner, while an `ORGANIZATION` workspace should be visible to users of the same organization as the owner.

* by default, all workspace creation with the create command are flagged as `PRIVATE`
* if the value of visibility is not `PRIVATE` or `ORGANIZATION`, case-insensitive, the command will return an error: `Error: value [...] for option [workspace.create.visibility] is invalid`

### json

The JSON printer switch affects the type of output the command yields. In JSON mode, the command will report the workspace created as a `REST` resource to `STDOUT`

* by default, the workspace create command has a single operation report with the workspace id created: `Created workspace id [...]`
