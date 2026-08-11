## Workspace Flow

Flow is a command group which targets flow instances created under workspaces for the purpose of running Solution
Workflows. A flow is similar to a container definition which gets re-used by the Orchestration and GenAIz clients.

### flow create

```
genaiz ws flow create [WORKSPACE_NAME]|WORKSPACE_ID [SOLUTION_FQDN_VERSION WORKFLOW_HANDLE]|WORKFLOW_ID \
  --name=NAME --description=DESC \
  --account=[[<user>@]host] \
  --json
```

The create command can be used to create a new flow container under a workspace, it requires the ID of an existing
workflow under an optionally specified solution. The command can be used with both the internal ids and reference
strings. For the workspace, the command will autocomplete the workspace name if it is unique. Solutions can only be
referred by FQDN and version. Workflows can be referred by handle or Ids.

All arguments, workspace, solution and workflow should support autocompletion provided there's an active session to use.

The command can be invoked with only 2 arguments, the workspace Id and the workflow Id for which the flow needs to be
created. When a third argument is specified, the 2 argument becomes a solution string or id, which is optional for when
a workflow id is already known.

#### WORKSPACE_NAME

* The command expects a workspace name that exists, if it can not find the specified workspace by name, it will return
  an error: `Error: workspace [...] could not be found`
* If the workspace string passed is matched to multiple workspaces the command will return an error:
  `Error: workspace [...] can be several possibilities`

#### WORKSPACE_ID

* If the command finds the input is a valid integer, then a workspace by id is assumed to be the target. If it can not
  find a workspace with the provided id, it will return an error: `Error: workspace [...] can not be accessed`

#### SOLUTION_FQDN_VERSION

* If the command can not find any solution with the specified FQDN:VERSION string, it will return an error:
  `Error: solution [...] could not be found`

#### WORKFLOW_HANDLE

* If the command receives a string, it will list all workflows for the specified and attempt matching a single workflow
  by handle. If no workflow correspond to the string, it will return an error: `Error: workflow [...] can not be found`

#### WORKFLOW_ID

* If the command receives an integer, it will attempt retrieving a workflow with the corresponding Id from the specified
  solution. If no workflow exists with the id, it will return an error: `Error workflow [...] does not exist`

#### name

The name option of the create command is used to initialize the Workspace Flow name.

* if the name does not resolve to a valid name string (see [name validity](../index.md#name)), the command will return
  an error with the key of the field and the shortened invalid value:
  `Error: value [...] for option [workspace.flow.create.name] is invalid`

#### desc

The description option of the create command is used to initialize Workspace Flow description.

* if the name does not resolve to a valid description string (see [description validity](../index.md#description)), the
  command will return an error with the key of the field and the shortened invalid value:
  `Error: value [...] for option [workspace.flow.create.description] is invalid`

#### account

The account for which to list the available solutions.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

#### json

The JSON flag indicates the command should return the Workspace Flow created as JSON output.
