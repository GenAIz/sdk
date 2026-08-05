## Datalink Create

```
genaiz dk create [OEM/]HANDLE[:VERSION] [CONFIG_FOLDER] \
  --oem=OEM \
  --version=VERSION \
  --name=NAME \
  --description=DESCRIPTION \
  --config-type yaml|toml|json \
  --user-defined
```

### handle

* the command expects its first argument to at least be a handle, but it may be a FQDN string composed of the
  OEM/HANDLE:VERSION fields, which are all required
* if the handle value does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command
  will return an error with the key of the field and the invalid value:
  `Error value [...] for option [datalink.create.handle] is invalid`

### oem

* if the oem value is specified as part of the first argument, it will override the option regardless of validity.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will
  return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.create.oem] is invalid`

### version

* if the version value is specified as part of the first argument, it will override the option regardless of validity.
* if no version is specified the value will default to `1.0.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the
  command will return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.create.version] is invalid`

### config_folder

* if config_folder is specified, the command will look for a `Genaiz.yaml` under the specified folder to update or to
  create it.
* if config_folder is not specified and [user-defined](#user-defined) is false, the current working directory will be
  used to update or create the configuration.
* if [user-defined](#user-defined) is specified, with no config_folder argument, the file written will be
  `$HOME/.config/genaiz/Genaiz.yaml`

### name

* if name is not specified, the value of handle will be used.
* if name does not match a valid name string, (see [name validity](index.md#name)), the command will return an error
  with the field and the shortened value: `value [...] for option [datalink.create.name] is invalid`

### description

* if description does not match a valid description, (see [description validity](index.md#description)), the command
  will return an error with the field and the shortened value:
  `value [...] for option [datalink.create.description] is invalid`

### config-type

* if config-type is not specified, the default is YAML
* if the config-type specified is not a recognized value, the command will return an error with the field and the value:
  `value [...] for option [datalink.create.configtype] is invalid`

> [!IMPORTANT]
> Only YAML is supported on the current early version of the SDK

### user-defined

* user-defined is set to `TRUE` by default, which will direct the command to write under the user's .config folder
* if user-defined is set to `FALSE` explicitly, the command will use either the provided [config-folder](#config_folder)
  argument or the default working directory.
