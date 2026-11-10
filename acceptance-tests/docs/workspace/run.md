## Workspace Run

Run is a command group of workspace which targets runs created within a workspace. A workspace run is always created
from the perspective of a workspace and a [workspace flow](flow.md).

### run list

```
genaiz ws run list [WORKSPACE_NAME]|WORKSPACE_ID \
  --monthly --weekly --today -- \
  --account=[[<user>@]host] \
  --json
```

The run list command is used to display runs under workspace within a certain time frame. Lists must always be time
delimited. By default, list will only use [today](#today) as the default time frame.

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

#### monthly

The monthly option

#### weekly

#### today

#### account

#### json