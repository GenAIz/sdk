## Datalink Publish

```
genaiz dk publish [OEM/]HANDLE[:VERSION] [CONFIG_FOLDER] \
    --oem=OEM \ 
    --version=VERSION \
    --published-version=VERSION \
    --config-type=YAML \
    --user-defined \
```

Publish will publish the structure of a locally defined datalink onto an Orchestrator. If the version is not yet
present, it will be accepted, otherwise it needs to be incremented.

### Handle

* the command expects its first argument to at least be a handle, but it may be a FQDN string composed of the
  OEM/HANDLE:VERSION fields, which are all required
* if the handle value does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command
  will return an error with the key of the field and the invalid value:
  `Error value [...] for option [datalink.publish.handle] is invalid`

### Oem

* if the oem value is specified as part of the first argument, it will override the option regardless of validity.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will
  return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.publish.oem] is invalid`

### Version

* if the version value is specified as part of the first argument, it will override the option regardless of validity.
* if no version is specified the value will default to `1.0.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the
  command will return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.publish.version] is invalid`

### published-version

* if specified, the datalink at [version](#version) will be published with a new version.
* if the published version does not match a valid SemVer version string (see [version validity](index.md#version)), the
  command will return as error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.publish.publishedversion] is invalid`

### config_folder

* if config_folder is specified, the command will look for a `Genaiz.yaml` configuration under the specified folder for
  a matching datalink to publish.
* if config_folder is not specified and [user-defined](#user-defined) is false, the current working directory will be
  searched for a matching datalink to publish.
* if [user-defined](#user-defined) is specified, with no config_folder argument, the file written will be
  `$HOME/.config/genaiz/Genaiz.yaml`

### config-type

* if config-type is not specified, the default is YAML
* if the config-type specified is not a recognized value, the command will return an error with the field and the value:
  `value [...] for option [datalink.publish.configtype] is invalid`

> [!IMPORTANT]
> Only YAML is supported on the current early version of the SDK

### user-defined

* user-defined is set to `TRUE` by default, which will direct the command to read and write configuration files under
  the user's .config folder
* if user-defined is explicitly set to `FALSE`, the command will use [config-folder](#config_folder) or the current
  working directory to persist datalink modifications
* if a [config-folder](#config_folder) is provided, user-defined will be overridden to `FALSE`, regardless of whether it
  is specified or not
