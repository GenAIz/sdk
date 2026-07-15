# Workflow List

```
genaiz wf list [FOLDER|FQDN] \
  --account=[[<user>@]host] \
  --json
```

Workflow list is used to produce a list, either tab delimited with workflow id, handle, created timestamp, solution fqdn
and version if multiple solutions are listed, the amount of nodes in the workflow and a local flag, or with all details
in JSON format. The workflows listed are the ones held by an account known to the SDK with a valid session or the
workflows of the solutions found under the specified folder if it exists.

The command does not filter the result set. It the argument provided can neither be interpreted as an existing path, or
a valid FQDN string, it will return an empty list or print nothing on the screen.

### FOLDER

* if the string passed evaluates to an existing folder, the command will attempt listing all workflows within the path
  recursively for all solutions found.
* folder values can not contain the character `:`, if the command finds the character it will interpret it
  as a [FQDN](#fqdn)
* if the string passed does not evaluate to an existing folder, the command will attempt parsing it as a [FQDN](#fqdn)

### FQDN

* if the string passed contains the character `:`, the command will try to resolve a solution on the specified account
  with that fqdn. If it can't resolve a solution it will return an error: `Error: solution [...] could not be resolved`
* if the string passed is not a valid FQDN, the command will return an error:
  `Error: [...] is not a valid path or solution fqdn`
* FQDN values may auto-complete if the shell completion script is sourced. If no solutions can be auto-completed, bash
  will default to matching folders
* if there are no active accounts, the command will return an error: `Error: no login session`

### account

The account for which to list the available solutions.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

### json

The JSON printer switch affects the type of output the command yields. In JSON mode, the command will display the
solutions listed as a `REST` array of resources to `STDOUT`

* by default, the solution list command will display a tab-delimited table with solution id, name, fqdn, creation,
  status indicator and a local flag columns.
