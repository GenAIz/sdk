# Data Source

Source is a command group of workspace which targets data sources available to a workspace flow nodes or that are used
by flow nodes. Managing node sources is split between source and [node](../workspace/node.md). The source command is
strictly for listing source instances.

## source list

```
genaiz data source list [OEM[/HANDLE][:VERSION][-rc-N]] \
    --account=[[<user>@]host] \
    --json
```

Using the list command should not typically produce any errors unless the account specified can not be used for some
reason. The filter provided will be parsed to DataLink coordinates and may also provoke syntax errors.

### OEM

If a single value is provided as argument to the list command, it will be interpreted as an OEM.

* If no data sources match the provided OEM an empty list will be rendered.
* If the value of OEM parses to an empty value, ex: `/handle`, the command will return an error:
  `Error: oem is required for filtering`

### HANDLE

If the value provided contains a `/`, the command will assume that the suffix of the argument is a handle value that
should be matched

* If no data sources match the argument with a handle, the list command will render an empty list.
* If the value of HANDLE parses to an empty value, ex: `oem/`, the command assumes no handle should be filtered.
* If the value of HANDLE parses to an empty value, ex: `oem/:1.0.0`, but VERSION or -RC-N evaluate to a value, the
  command will return an error: `Error: handle is invalid for version`

### VERSION

If the value provided contains a `:`, the command will assume that the suffix of the argument is a version string.

* If no data sources match the argument with a version, the list command will render an empty list.
* If the value of VERSION parses to an empty value, ex: `oem/version:`, the command assumes no version should be
  filtered.
* If the value of VERSION parses to an empty value, ex: `oem/handle:-rc-2`, but SEQUENCE is provided, the command
  returns an error: `Error: version is invalid for sequence`

### RC-N

A sequence number may be passed with the filter by appending the string `-rc-N` where N is the integer corresponding to
the sequence number targeted on the data source version.

* If no data sources match the argument with a sequence suffix, the list command will render an empty list.
* If the value of N parses to an invalid integer, the command will return an error:
  `Error: sequence number is invalid`.

### account

Accounts do not need to be activated for the command to query the Orchestration session, but the default account will
always be the one currently activated.

* if the account does not evaluate to an active session, the command will return an error:
  `Error: account session is unknown`
* if the account session is expired, the command will return an error: `Error: broker session is expired`
* account values will auto-complete if the shell completion script is sourced.

### json

The JSON printer switch affects the type of output the command yields. In JSON mode, the command will display the
datalinks listed as a `REST` array of resources to `STDOUT`

* by default, the datalink list command will display a tab-delimited table with datalink id, name, fqdn, creation,
  status indicator and a local flag columns.
