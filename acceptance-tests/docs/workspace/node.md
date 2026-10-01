## Workspace Node

Node is a command group of workspace which targets node instances created under a [Workspace Flow](flow.md) after its
inception. A node id is always a requirement for being able to configure a Node's Smart Function parameters before
executing the Flow.

### node data src add

```
genaiz ws node data src add [WORKSPACE_NAME]|WORKSPACE_ID \
  [WORKFLOW_HANDLE]|FLOW_ID \ 
  [SF_HANDLE]|NODE_ID \
  [SOURCE_NAME]|SOURCE_ID \
  --account=[[<user>@]host] \
  --json
```

The data src add command is used to assign a data source published from a [Locker](../locker/index.md), to an existing
workspace flow node. The command does not require unlocking a locker file.

The command can be invoked multiple times, and it will preserve data sources already assigned. If the source added is
already assigned, the command will succeed and be a no-op.

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
* Workflow handles will autocomplete to values with limitations on duplicates. Multiple values will need to be resolved
  by Flow ID

#### FLOW_ID

When the command finds an integer id as the second argument, it first tries to resolve it as a flow id on the specified
workspace.

* If the id can not be found in the workspace flows, the command will return an error:
  `Error: Unknown workspace flow id [...]`
* Flow ids will autocomplete to workspace flow values

#### SF_HANDLE

If the command finds a smart function handle string as the third argument, it will attempt listing all nodes under the
specified
workspace flow and attempt matching the value with the Handle of the Smart Function in the Solution Workflow.

* If the command can not find a flow node with the provided node handle, it will return an error:
  `Error: node handle [...] could not be found`.
* If the value of the Smart Function handle evaluates to a duplicate value in the workspace flow, the command will
  return an error: `Error: smart function [...] under workspace flow [...] is not unique`
* Smart Function handles should autocomplete to values present in the associated Workflow for the Solution that was
  deployed on the Workspace flow.

#### NODE_ID

When the command finds an integer id as the third argument, it will try to resolve it as a flow node id on the specified
workspace flow.

* If the id can not be found in the workspace flow nodes, the command will return an error:
  `Error: Unknown flow node id [...]`
* Node ids will autocomplete to workspace flow node values

#### SOURCE_NAME

The command will attempt to resolve the fourth argument as a DataSource name when it does not correspond to a sole
integer.

* If the command can not find a data source with the specified string name, it will return an error:
  `Error: data source [...] could not be found`
* Data Source names should autocomplete to values accessible to the user on a list request.

> [!IMPORTANT]
> The list request on brokers, currently only return the data sources for which the user is an owner. Although the
> command could work with the ID of an Organization level source, it won't list the value.
>
> Second limitation is with names which can contain spaces. The command makes a best effort list, but spaces typically
> terminate the completed string.

#### SOURCE_ID

If the command finds an integer as the last argument, it tries to resolve the argument as the ID of a Data Source. This
is done with the Read endpoint, which is more permissive than the list.

* If the id can not be found or is inaccessible to the user, the command will return an error:
  `Error: Unknown data source id [...]`
* Data source ids will autocomplete to values accessible to the user on a list request.

#### account

The account for which to list the available solutions.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

#### json

The JSON flag indicates the command should return the Workspace Flow created as JSON output.

The command should output a graph of the Node with associated resources.

### node data src rm

```
genaiz ws node data src rm [WORKSPACE_NAME]|WORKSPACE_ID \
  [WORKFLOW_HANDLE]|FLOW_ID \ 
  [SF_HANDLE]|NODE_ID \
  [SOURCE_NAME]|SOURCE_ID \
  --account=[[<user>@]host] \
  --json
```

The data src remove command is used to remove a data source assignment from an existing
workspace flow node. The command can be invoked multiple times, removing multiple data sources if needed.

If the source added conflicts with another source for the same datalink, the command should return an error:
`Error: data source [...] conflicts on link instance [...]`

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
* Workflow handles will autocomplete to values with limitations on duplicates. Multiple values will need to be resolved
  by Flow ID

#### FLOW_ID

When the command finds an integer id as the second argument, it first tries to resolve it as a flow id on the specified
workspace.

* If the id can not be found in the workspace flows, the command will return an error:
  `Error: Unknown workspace flow id [...]`
* Flow ids will autocomplete to workspace flow values

#### SF_HANDLE

If the command finds a smart function handle string as the third argument, it will attempt listing all nodes under the
specified
workspace flow and attempt matching the value with the Handle of the Smart Function in the Solution Workflow.

* If the command can not find a flow node with the provided node handle, it will return an error:
  `Error: node handle [...] could not be found`.
* If the value of the Smart Function handle evaluates to a duplicate value in the workspace flow, the command will
  return an error: `Error: smart function [...] under workspace flow [...] is not unique`
* Smart Function handles should autocomplete to values present in the associated Workflow for the Solution that was
  deployed on the Workspace flow.

#### NODE_ID

When the command finds an integer id as the third argument, it will try to resolve it as a flow node id on the specified
workspace flow.

* If the id can not be found in the workspace flow nodes, the command will return an error:
  `Error: Unknown flow node id [...]`
* Node ids will autocomplete to workspace flow node values

#### SOURCE_NAME

The command will attempt to resolve the fourth argument as a DataSource name when it does not correspond to a sole
integer.

* If the command can not find a data source with the specified string name, it will return an error:
  `Error: data source [...] could not be found`
* Data Source names should autocomplete to values accessible to the user on a list request.

> [!IMPORTANT]
> The list request on brokers, currently only return the data sources for which the user is an owner. Although the
> command could work with the ID of an Organization level source, it won't list the value.
>
> Second limitation is with names which can contain spaces. The command makes a best effort list, but spaces typically
> terminate the completed string.

#### SOURCE_ID

If the command finds an integer as the last argument, it tries to resolve the argument as the ID of a Data Source. This
is done with the Read endpoint, which is more permissive than the list.

* If the id can not be found or is inaccessible to the user, the command will return an error:
  `Error: Unknown data source id [...]`
* Data source ids will autocomplete to values accessible to the user on a list request.

#### account

The account for which to list the available solutions.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

#### json

The JSON flag indicates the command should return the Workspace Flow created as JSON output.

The command should output a graph of the Node with associated resources.

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
