## Datalink List

```
genaiz dk list FOLDER|OEM[/HANDLE[:VERSION]] \
  --account=[[<user>@]host] --account-only[=true] \
  --json
```

Datalink list is used to produce a list, either tab delimited with datalink id, name, fqdn, creation, status
indicator and a local flag, or with all details in JSON format. The datalinks listed are the ones held by an account
known to the SDK with a valid session or the datalinks found under the specified folder if it exists.

By default, the command will filter datalinks only if they are listed by `OEM/HANDLE[:VERSION]` string. If the
`OEM/HANDLE` string corresponds to an existing local folder, it will try merging the results into one set. If the string
contains a version it is assumed that --account-only is true.

### FOLDER

* if the string passed evaluates to an existing folder, the command will attempt listing all datalinks within the path
  recursively.
* if the string passed does not evaluate to an existing folder, and an account is active, the command will list all
  datalinks available to the account matching the string has a prefix.
* when the string passed evaluates to an existing folder, it may be matched to account datalinks as well and produce a
  listing of both locations with the remote datalinks favored on duplicates on released versions.

### OEM

* if oem is provided, the command will attempt matching remote datalinks with a FQDN string prefixed with it.
* if oem is both a local folder and a remote classification, the list produced is a merged list with remote datalinks
  favored on duplicates.

### HANDLE

* if handle is provided, the command will attempt matching remote datalinks with a FQDN string prefixed with it.
* if oem/handle both corresponds to a local folder and a remote classification, the list produced is a merged list with
  remote datalinks favored on duplicates involving released versions.

### VERSION

* if version is provided, the command will attempt matching remote datalinks with a FQDN string prefixed with it.
  Sequence values may be considered not duplicates.
* the command will not attempt to match a local folder if `:VERSION` is provided as argument, with the colon not
  considered a valid handle character.

### account

The account for which to list the available datalinks.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error could not elect a session`
* account values will auto-complete if the shell completion script is sourced.
* the account switch implies [account-only](#account-only) is set and only datalinks from the account will be listed.

### account-only

Account only will restrict the result set to datalinks that are available on the broker available to the account
specified or the active session one.

* by default the command does not restrict the set of datalinks and may produce a merged result set.

### json

The JSON printer switch affects the type of output the command yields. In JSON mode, the command will display the
datalinks listed as a `REST` array of resources to `STDOUT`

* by default, the datalink list command will display a tab-delimited table with datalink id, name, fqdn, creation,
  status indicator and a local flag columns.
