## Workspace Listing

```
genaiz ws list --account=[[<user>@]host] --owner-only \
    --rc-enabled[=true|false] --today|--weekly|--monthly --json
```

Workspace list is used to produce a list, either tab delimited with workspace id, name, creation and type/status indicators, or with all details in JSON format. The workspaces listed are the ones held by an account known to the SDK with a valid session.

By default, the command does not apply any filtering and lists everything.

### account

The account for which to list the available workspaces.

* if the value of the account does not evaluate to a Host address, the command returns an error: `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

### owner-only

Owner only will restrict the result set to workspaces that were created by the same user as the one owning the account session used.

* by default, all available workspaces are listed, including `ORGANIZATION` listings created by different users

### rc-enabled

The rc-enabled flag instructs the command to include or exclude workspaces that can use release candidate workflows.

### today

Today restricts the result set to all workspaces which were created on the same day after Midnight.

### weekly

Weekly restricts the result set to all workspaces which were created in the same week after Midnight on the last Sunday.

### monthly

Monthly restricts the result set to all workspaces which were created during the same current month.

### json

The JSON printer switch affects the type of output the command yields. In JSON mode, the command will display the workspaces listed as a `REST` array of resources to `STDOUT`

* by default, the workspace list command will display a tab-delimited table with workspace id, name, creation date, status and type column.
