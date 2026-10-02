## Workspace Node

Node is a command group of workspace which targets node instances created under a [Workspace Flow](flow.md) after its
inception. A node id is always a requirement for being able to configure a Node's Smart Function parameters before
executing the Flow.

  * [Workspace Node](#workspace-node)
    * [node data src add](#node-data-src-add)
    * [node data src rm](#node-data-src-rm)
    * [node data str add](#node-data-str-add)
    * [node data str rm](#node-data-str-rm)
    * [node list](#node-list)

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

The command will attempt to resolve the fourth argument as a Data Source name when it does not correspond to a sole
integer.

* If the command can not find a data source with the specified string name, it will return an error:
  `Error: data source [...] could not be found`
* Data Source names should autocomplete to values accessible to the user on a list request.

> [!IMPORTANT]
> There is a limitation with names which can contain spaces. The command makes a best effort list, but spaces typically
> terminate the completed string.

#### SOURCE_ID

If the command finds an integer as the last argument, it tries to resolve the argument as the ID of a Data Source.

* If the id can not be found or is inaccessible to the user, the command will return an error:
  `Error: Unknown data source id [...]`
* Data source ids will autocomplete to values accessible to the user on a list request.

#### account

The account for which to list the available nodes.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

#### json

The JSON flag indicates the command should return the Workspace Flow Node updated as JSON output.

### node data src rm

```
genaiz ws node data src rm [WORKSPACE_NAME]|WORKSPACE_ID \
  [WORKFLOW_HANDLE]|FLOW_ID \ 
  [SF_HANDLE]|NODE_ID \
  [SOURCE_NAME]|SOURCE_ID \
  --account=[[<user>@]host] \
  --json
```

The data source remove command is used to remove a data source assignment from an existing
workspace flow node. 

The command can be invoked multiple times, removing multiple data sources if needed. If the source removed is not assigned
to the node targeted, the command succeeds and is a no-op.

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

The command will attempt to resolve the fourth argument as a Data Source name when it does not correspond to a sole
integer.

* If the command can not find a data source with the specified string name, it will return an error:
  `Error: data source [...] could not be found`
* Data Source names should autocomplete to values accessible to the user on a list request.

> [!IMPORTANT]
> There is a limitation with names which can contain spaces. The command makes a best effort list, but spaces typically
> terminate the completed string.

#### SOURCE_ID

If the command finds an integer as the last argument, it tries to resolve the argument as the ID of a Data Source.

* If the id can not be found or is inaccessible to the user, the command will return an error:
  `Error: Unknown data source id [...]`
* Data source ids will autocomplete to values accessible to the user on a list request.

#### account

The account for which to list the available nodes.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

#### json

The JSON flag indicates the command should return the Workspace Flow Node updated as JSON output.

### node data str add

```
genaiz ws node data str add [WORKSPACE_NAME]|WORKSPACE_ID \
  [WORKFLOW_HANDLE]|FLOW_ID \ 
  [SF_HANDLE]|NODE_ID \
  [STORE_NAME]|STORE_ID \
  --account=[[<user>@]host] \
  --json
```

The data store add command is used to assign a data store published from a [Locker](../locker/index.md), to an existing
workspace flow node. The command does not require unlocking a locker file.

The command can be invoked multiple times, and it will preserve data stores already assigned. If the store added is
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

#### STORE_NAME

The command will attempt to resolve the fourth argument as a Data Store name when it does not correspond to a sole
integer.

* If the command can not find a data store with the specified string name, it will return an error:
  `Error: data store [...] could not be found`
* Data Store names should autocomplete to values accessible to the user on a list request.

> [!IMPORTANT]
> There is a limitation with names which can contain spaces. The command makes a best effort list, but spaces typically
> terminate the completed string.

#### STORE_ID

If the command finds an integer as the last argument, it tries to resolve the argument as the ID of a Data Store.

* If the id can not be found or is inaccessible to the user, the command will return an error:
  `Error: Unknown data store id [...]`
* Data store ids will autocomplete to values accessible to the user on a list request.

#### account

The account for which to list the available nodes.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

#### json

The JSON flag indicates the command should return the Workspace Flow Node updated as JSON output.

### node data str rm

```
genaiz ws node data str rm [WORKSPACE_NAME]|WORKSPACE_ID \
  [WORKFLOW_HANDLE]|FLOW_ID \ 
  [SF_HANDLE]|NODE_ID \
  [STORE_NAME]|STORE_ID \
  --account=[[<user>@]host] \
  --json
```

The data store remove command is used to remove a data store assignment from an existing
workspace flow node.

The command can be invoked multiple times, removing multiple data stores if needed. If the store removed is not assigned
to the node targeted, the command succeeds and is a no-op.

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
specified workspace flow and attempt matching the value with the Handle of the Smart Function in the Solution Workflow.

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

#### STORE_NAME

The command will attempt to resolve the fourth argument as a Data Store name when it does not correspond to a sole
integer.

* If the command can not find a data source with the specified string name, it will return an error:
  `Error: data store [...] could not be found`
* Data Store names should autocomplete to values accessible to the user on a list request.

> [!IMPORTANT]
> There is a limitation with names which can contain spaces. The command makes a best effort list, but spaces typically
> terminate the completed string.

#### STORE_ID

If the command finds an integer as the last argument, it tries to resolve the argument as the ID of a Data Store.

* If the id can not be found or is inaccessible to the user, the command will return an error:
  `Error: Unknown data store id [...]`
* Data store ids will autocomplete to values accessible to the user on a list request.

#### account

The account for which to list the available nodes.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

#### json

The JSON flag indicates the command should return the Workspace Flow Node updated as JSON output.

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
