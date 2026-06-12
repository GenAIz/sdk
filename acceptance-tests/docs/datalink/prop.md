## Datalink Prop

The prop commands of datalink are used to define property specifications on locally synchronized or created data links.
This would be done before publishing the datalink, as an admin, to an Orchestrator.

This is specifically critical for testing Smart Function that are of the `Connector` type. Most of these connectors
require credentials to retrieve data and bring it to the volume of `Workflow`.

* if the prop commands can not find the specified datalink locally, they will return an error with the `FQDN` not found:
  `Error: data link [FQDN] not found`
    * either [create](create.md) or [sync](sync.md) will need to be called to bring the definition to edit locally
    * note that [publish](publish.md) will not be able to push an existing datalink definition if the version is not
      incremented when publishing

### prop add

```
genaiz dk prop add [OEM/]HANDLE[:VERSION] KEY \
    --oem=OEM \ 
    --version=VERSION \
    --type=bool|double|int|string|enum \
    --secret \  
    --name=NAME \
    --description=DESCRIPTION \
    --default-value=VALUE \
    --enum-value=value1 --enum-value=value2 \
    --config-type=yaml|toml|json \
    --user-defined
```

#### handle
``
* the command expects its first argument to at least be a handle, but it may be a FQDN string composed of the
  OEM/HANDLE:VERSION fields, which are all required
* if the handle value does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command
  will return an error with the key of the field and the invalid value:
  `Error value [...] for option [datalink.propspecadd.handle] is invalid```

#### oem

* if the oem value is specified as part of the first argument, it will override the option regardless of validity.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will
  return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.propspecadd.oem] is invalid`

#### version

* if the version value is specified as part of the first argument, it will override the option regardless of validity.
* if no version is specified the value will default to `1.0.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the
  command will return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.propspecadd.version] is invalid`

#### key

* if the key does not match a valid Key string (see [KEY validity](../index.md#property-key)), the command will return
  an error with the offending key: `Error: [...] is not a valid environment key`
* if the key is already used by another specification, an error will be returned with the offending key:
  `Error: the key [...] already exists`

#### type

* if the type is not specified the default type is "string"
* if the type does not match a valid type string (see [type validity](../index.md#property-types)), an error will be
  returned with the value: `Error: value [...] for option [datalink.propspecadd.type] is invalid`

#### secret

* if secret is not specified, the property will be added to the regular PropSpec set of the selected datalink.
* if secret is specified, the property will be added to the SecretSpecs set of the selected datalink.
* secret properties can not have default values. If secret and [default-value](#default-value) are specified, the
  command returns an error with the value:
  `Error: value [...] for option [datalink.propspecadd.defaultvalue] is invalid`

#### name

* if the name does not match a valid name string (see [name validity](../index.md#property-name)), the command will
  return an Error: `Error: value [...] for option [datalink.propspecadd.name] is invalid`
* if the resolved name is empty, name will default to the value [KEY](#key)

#### description

* if the description does not match a valid description string (
  see [description validity](../index.md#property-description)), the command will return an Error:
  `Error: value [...] for option [datalink.propspecadd.description] is invalid`
* a description is optional and will be left empty if not specified

#### default-value

* a default value must always be valid according to the specified [type](#type)
* if default value is left empty, it will be interpreted as no default value and the property must be specified.
* if the default value specified is not valid for the property type, the command returns an error:
  `Error: illegal default value for [...] type`

#### enum-value

* can be specified multiple times or be a comma separated string of values
* if an enum value does not match a valid enum string (see [enum validity](../index.md#property-enum-values)), the
  command will return an error with the key of the field and the invalid value;
  `Error: value [n:...] for option [datalink.propspecadd.enumvalue] is invalid`

#### config-type

* if config-type is not specified, the default is YAML
* if the config-type specified is not a recognized value, the command will return an error with the field and the value:
  `value [...] for option [datalink.propspecadd.configtype] is invalid`

> [!IMPORTANT]
> Only YAML is supported on the current early version of the SDK

#### user-defined

* user-defined is set to `TRUE` by default, which will direct the command to read and write configuration files under
  the user's .config folder
* if user-defined is explicitly set to `FALSE`, the command will use the current working directory to persist datalink
  modifications

### prop edit

```
genaiz dk prop edit [OEM/]HANDLE[:VERSION] KEY \
    --oem=OEM \
    --version=VERSION \
    --secret \
    --name=NAME \
    --description=DESCRIPTION \
    --default-value=VALUE \
    --enum-value=value1 --enum-value=value2 \
    --add-enum-value=value1 --add-enum-value=value2 \
    --rm-enum-value=value1 --rm-enum-value=value2 \
    --config-type=yaml|toml|json \
    --user-defined
```

#### handle

* the command expects its first argument to at least be a handle, but it may be a FQDN string composed of the
  OEM/HANDLE:VERSION fields, which are all required
* if the handle value does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command
  will return an error with the key of the field and the invalid value:
  `Error value [...] for option [datalink.propspecedit.handle] is invalid`

#### oem

* if the oem value is specified as part of the first argument, it will override the option regardless of validity.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will
  return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.propspecedit.oem] is invalid`

#### version

* if the version value is specified as part of the first argument, it will override the option regardless of validity.
* if no version is specified the value will default to `1.0.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the
  command will return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.propspecedit.version] is invalid`

#### key

* if the key is not valid or not an actual property specification under PropSpecs or SecretSpecs, the command returns an
  error with the value: `Error: the key [...] could not be found`

#### name

* if the name does not match a valid name string (see [name validity](../index.md#property-name)), the command will
  return an Error: `Error: value [...] for option [datalink.propspecedit.name] is invalid`
* if the resolved name is empty, name will default to the value [KEY](#key)

#### description

* if the description does not match a valid description string (
  see [description validity](../index.md#property-description)), the command will return an Error:
  `Error: value [...] for option [datalink.propspecedit.description] is invalid`
* a description is optional and will be left empty if not specified

#### default-value

* a default value must always be valid according to the specified [type](#type)
* if default value is left empty, it will be interpreted as no default value and the property must be specified.
* if the default value specified is not valid for the property type, the command returns an error:
  `Error: illegal default value for [...] type`

#### enum-value

* can be specified multiple times or be a comma separated string of values
* if an enum value does not match a valid enum string (see [enum validity](../index.md#property-enum-values)), the
  command will return an error with the key of the field and the invalid value;
  `Error: value [n:...] for option [datalink.propspecedit.enumvalue] is invalid`
* if enum-value is specified, it replaces the existing value set, and then [add-enum-value](#add-enum-value)
  and [rm-enum-value](#rm-enum-value) are processed.

#### add-enum-value

* can be specified multiple times or be a comma separated string of values
* if omitted, no new value will be added to the current set of enum values
* if an enum value does not match a valid enum string (see [enum validity](../index.md#property-enum-values)), the
  command will return an error with the key of the field and the invalid value;
  `Error: value [n:...] for option [datalink.propspecedit.enumvalue] is invalid`
* adding an enum value to a non-enum property specification will return an error:
  `Error: the property spec type does not allow enum values`

#### rm-enum-value

* can be specified multiple times or be a comma separated string of values
* if omitted, no value will be removed from the current set of enum values
* removing an enum value which is not a current enum value, **will not yield an error**. This is because the state of
  the specification is still valid and will not contain what the user wanted to remove
* removing an enum value from a non-enum property specification will return an error:
  `Error: the property spec type does not allow enum values`

#### config-type

* if config-type is not specified, the default is YAML
* if the config-type specified is not a recognized value, the command will return an error with the field and the value:
  `value [...] for option [datalink.propspecadd.configtype] is invalid`

> [!IMPORTANT]
> Only YAML is supported on the current early version of the SDK

#### user-defined

* user-defined is set to `TRUE` by default, which will direct the command to read and write configuration files under
  the user's .config folder
* if user-defined is explicitly set to `FALSE`, the command will use the current working directory to persist datalink
  modifications

### prop rm

```
genaiz dk prop rm [OEM/]HANDLE[:VERSION] KEY \
    --oem=OEM \
    --handle=HANDLE \
    --version=VERSION \
    --config-type=yaml|toml|json \
    --user-defined
```

#### handle

* the command expects its first argument to at least be a handle, but it may be a FQDN string composed of the
  OEM/HANDLE:VERSION fields, which are all required
* if the handle value does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command
  will return an error with the key of the field and the invalid value:
  `Error value [...] for option [datalink.propspecremove.handle] is invalid`

#### oem

* if the oem value is specified as part of the first argument, it will override the option regardless of validity.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will
  return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.propspecremove.oem] is invalid`

#### version

* if the version value is specified as part of the first argument, it will override the option regardless of validity.
* if no version is specified the value will default to `1.0.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the
  command will return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.propspecremove.version] is invalid`

#### key

* if the key is not valid or not an actual property specification under PropSpecs or SecretSpecs, the command returns an
  error with the value: `Error: the key [...] could not be found`

#### config-type

* if config-type is not specified, the default is YAML
* if the config-type specified is not a recognized value, the command will return an error with the field and the value:
  `value [...] for option [datalink.propspecadd.configtype] is invalid`

> [!IMPORTANT]
> Only YAML is supported on the current early version of the SDK

#### user-defined

* user-defined is set to `TRUE` by default, which will direct the command to read and write configuration files under
  the user's .config folder
* if user-defined is explicitly set to `FALSE`, the command will use the current working directory to persist datalink
  modifications
