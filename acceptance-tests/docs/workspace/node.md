## Workspace Node

Node is a command group of workspace which targets node instances created under a [Workspace Flow](flow.md) after its
inception. A node id is always a requirement for being able to configure a Node's Smart Function parameters before
executing the Flow.

### node list

```
genaiz ws node list [WORKSPACE_NAME]|WORKSPACE_ID \ 
  [WORKFLOW_HANDLE]|FLOW_ID \
  --account=[[<user>@]host] \
  --json
```

The list command can be used to list the Workspace Flow nodes that were created for the Workflow of Solution. The
command can work with workspace names or their internal ids. It also requires a way to identify the Flow under which the
nodes can be found.

If a workflow handle or id are used, and there are multiple flows available for the specified workflow, the command will
return an error: `Error: workflow has multiple instances under workspace [...]`

#### WORKSPACE_NAME

* The command expects a workspace name that exists, if it can not find the specified workspace by name, it will return
  an error: `Error: workspace [...] could not be found`
* If the workspace string passed is matched to multiple workspaces the command will return an error:
  `Error: workspace [...] can be several possibilities`

> [!NOTE]
> Workspace name values should autocomplete, but with some limitations on processing white spaces. To avoid the issue,
> don't use white spaces in names.

#### WORKSPACE_ID

* If the command finds the input is a valid integer, then a workspace by id is assumed to be the target. If it can not
  find a workspace with the provided id, it will return an error: `Error: workspace [...] can not be accessed`
* workspace ids should autocomplete with the workspaces available to a logged in account.

#### WORKFLOW_HANDLE

If the command finds a workflow handle string as the second argument, it will attempt listing all solutions found in
the workspace, trying to retrieve a flow by workflow handle.

* If the command resolves multiple flows for the provided workflow handle it will return an error:
  `Error: Worklow [...] is used by multiple workspace flows`
* If the handle can not be found in the workspace flows, the command will return an error:
  `Error: Workflow [...] is not a configured under workspace [...]`
* workflow handles will autocomplete to values with limitations on duplicates. Multiple values will need to be resolved
  by Flow ID

#### FLOW_ID

When the command finds an integer id as the second argument, it first tries to resolve it as a flow id on the specified
workspace.

* If the id can not be found in the workspace flows, the command will return an error:
  `Error: Unknown workspace flow id [...]`
* Flow ids will autocomplete to workspace flow values

#### account

The account for which to list the available solutions.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

#### json

The JSON flag indicates the command should return the Workspace Flow created as JSON output.
