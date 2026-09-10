# Locker Source

The source sub-command of `locker` is used to add and update data source properties to a locker file. The properties are
used subsequently from the [publish](#source-publish) command or from [run](../function/run.md),
[start](../function/start.md) and [test](../function/test.md).

If the `locker source` commands can not find a locker file to manipulate, they return an error:
`Error: no accessible locker found, run init first?`

The locker file used by default is always `$HOME/.config/genaiz/locker.bin`, all commands support the `--locker`
option.

## source add

```
genaiz lk source add HANDLE FQDN:VERSION[-rc-N] \
  --account=[[<user>@]host] \
  --locker=LOCKER_PATH
```

Using the source add command requires a valid [Account](../account/index.md) session with an Orchestration. If no
account is found or specified the command will return an error: `Error: client not authenticated`

When adding a data source, the [handle](#handle) argument is always used within the context of the user's local
development environment. The handles are not shared across locker files and are not made unique by any Orchestration
Account.

Adding data sources for links which have no properties defined will fail with an error:
`Error: datalink property set is empty, is it incomplete?`

### HANDLE

The value of handle may be used by the [Data Source Create]() command to populate the
required name field of a Data Source.

* if the value of handle is duplicated within the locker's account sources, the command will return an error:
  `Error: data source [...] for account [...] is already defined`
* if the value of handle is not a valid [handle](../index.md#handle-and-oem), the command will fail with error:
  `Error: value [...] is not a valid handle`

### FQDN:VERSION

* if the value of FQDN:VERSION can not be matched with the latest release candidate or a released version for the
  datalink targeted, the command will return an error: `Error: datalink [...] is not accessible`

### -rc-N

Optionally, a specific release candidate can be passed using the format `-rc-N`.

* if the value of the RC tag does not follow the format expected, the command will return an error:
  `Error: invalid version tag, must follow -rc-N`
* if the value of the sequence `N` is not an integer, the command will return an error: `Error: invalid sequence [...]`

### account

Accounts do not need to be activated for the command to query the Orchestration session, but the default account will
always be the one currently activated.

* if the account does not evaluate to an active session, the command will return an error:
  `Error: account session is unknown`
* if the account session is expired, the command will return an error: `Error: broker session is expired`
* account values will auto-complete if the shell completion script is sourced.

### locker

When the user wishes to use a different locker file, this option expects a valid, readable path.

* if the value points to a locker that can not be read, the command will return an error:
  `Error: locker [...] can not be read`

## source update

```
genaiz lk source update HANDLE MyKey [MyValue] \
  --account=[[<user>@]host]
  --locker=LOCKER_PATH
```

> [!IMPORTANT]
> Limitations: We can not have a policy on empty secrets as those can be used in conjunctions with Outbound Proxying. So
> the update command will not fail on finding empty values, but a warning will be produce nonetheless.

### HANDLE

The value of handle is matched under the locker's account data sources for updating the properties contained within

* if the value of handle is not found within the locker's account sources, the command will return an error:
  `Error: data source [...] for account [...] does not exist`

### MyKey

All keys specified must be members of the [Prop Specs](../datalink/prop.md) of the associated Datalink with the source.
Secret keys can only be updated through STDIN.

* if the key specified is not a valid Prop Spec key, the command will return an error:
  `Error: property [...] is not a valid key for datalink [...]`
* if the key specified evaluates to a Secret Prop Spec and there is a second argument, the command returns an error:
  `Error: property [...] is a secret key for datalink [...]`

### MyValue

* if the [key](#mykey) specified corresponds to a Prop Spec key, the command should allow empty values to be set, but it
  should be reported to the user: `Property key [...] set to the empty string`
* if the value specified does not evaluate to the same type as the Prop Spec, the command will return an error:
  `Error: property value type for key [...] is invalid`

### locker

When the user wishes to use a different locker file, this option expects a valid, readable path.

* if the value points to a locker that can not be read, the command will return an error:
  `Error: locker [...] can not be read`

## source publish

```
genaiz lk source publish HANDLE \
  --name=NAME --description=DESCRIPTION --visibility=[PRIVATE|ORG]
  --account=[[<user>@]host]
  --locker=LOCKER_PATH
```

The publish command will create a data source with [name](#name) set to [HANDLE](#handle-2) if it isn't provided. The
data source will be created on the specified [account](#account-1) or the currently active account.

Publish will ask for a passphrase to open the specified [locker](#locker-2), if the passphrase can not be used to
decrypt the locker, the command returns an error: `Error: the passphrase failed decryption`

Publish does not necessarily error out if a data source with the same name already exists. It will check if the datalink
coordinates are the same, before deciding if there is a conflict. it should also be able to update the data link id on
the source, if the version has a newer release candidate.

### HANDLE

* if the value provided does not correspond to an existing data source in the specified [locker](#locker-2), the command
  returns and error: `Error: the data source [...] is unknown`
* if the handle or the name already exists for the user creating it, but with a different data link, the command will
  fail with an error: `Error: incompatible datalinks found for data source [...], locker [...] has [...]`

### name

* if the value of name is not specified, the handle string will be used instead.
* if name does not match a valid name string, (see [name validity](../index.md#name)), the command will return an error
  with the field and the shortened value: `value [...] for option [locker.datasource.publish.name] is invalid`

### description

* if description does not match a valid description, (see [description validity](../index.md#description)), the command
  will return an error with the field and the shortened value:
  `value [...] for option [locker.datasource.publish.description] is invalid`

### visibility

Visibility affects which users can access and manage the Workspace. A `PRIVATE` workspace can only be managed by its
owner, while an `ORGANIZATION` workspace should be visible to users of the same organization as the owner.

* by default, all workspace creation with the create command are flagged as `PRIVATE`
* if the value of visibility is not `PRIVATE` or `ORGANIZATION`, case-insensitive, the command will return an error:
  `Error: value [...] for option [workspace.create.visibility] is invalid`

### account

The account under which to create the workspace.

* if the value of the account does not evaluate to a Host address, the command returns an error:
  `Error: could not elect a session`
* account values will auto-complete if the shell completion script is sourced.

### locker

When the user wishes to use a different locker file, this option expects a valid, readable path.

* if the value points to a locker that can not be read, the command will return an error:
  `Error: locker [...] can not be read`
