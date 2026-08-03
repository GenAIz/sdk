## Solutions List

```
genaiz sn list [FOLDER|OEM[/HANDLE][:VERSION]] \
  --account=[[<user>@]host] --account-only[=true] \
  --json
```

Solution list is used to produce a list, either tab delimited with solution id, name, fqdn, creation, status
indicator and a local flag, or with all details in JSON format. The solutions listed are the ones held by an account
known to the SDK with a valid session or the solutions found under the specified folder if it exists.

By default, the command will filter solutions only if they are listed by `OEM/HANDLE:VERSION` string. To get all the
solutions for a specified `OEM` only provide the oem as prefix on the argument line.

### FOLDER

* if the string passed evaluates to an existing folder, the command will attempt listing all solutions within the path
  recursively.
* if the string passed does not evaluate to an existing folder, and an account is active, the command will list all
  solutions available to the account matching the string has a prefix.
* when the string passed evaluates to an existing folder, it may be matched to account solutions as well and produce a
  listing of both locations with the remote solution favored on duplicates on released versions.

### OEM

* if oem is provided, the command will attempt matching remote solutions with a FQDN string prefixed with it.
* if oem is both a local folder and a remote classification, the list produced is a merged list with remote solutions
  favored on duplicates.

### HANDLE

* if handle is provided, the command will attempt matching remote solutions with a FQDN string prefixed with it.
* if oem/handle both corresponds to a local folder and a remote classification, the list produced is a merged list with
  remote solution favored on duplicates involving released versions.

### VERSION

* if version is provided, the command will attempt matching remote solutions with a FQDN string prefixed with it.
  Sequence values may be considered not duplicates.
* the command will not attempt to match a local folder if `:VERSION` is provided as argument, with the colon not
  considered a valid handle character.

### account

The account for which to list the available solutions.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.
* the account switch implies [account-only](#account-only) is set and only solutions from the account will be listed.

### account-only

Account only will restrict the result set to solutions that are available on the broker available to the account
specified or the active session one.

* by default the command does not restrict the set of solutions and may produce a merged result set.

### json

The JSON printer switch affects the type of output the command yields. In JSON mode, the command will display the
solutions listed as a `REST` array of resources to `STDOUT`

* by default, the solution list command will display a tab-delimited table with solution id, name, fqdn, creation,
  status indicator and a local flag columns.
