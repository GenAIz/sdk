## Datalink Sync

```
genaiz dk sync [OEM/]HANDLE[:VERSION] [CONFIG_FOLDER] \
    --oem=OEM \ 
    --version=VERSION \
    --sequence=SEQ \
    --config-type=YAML \
    --user-defined
```
Sync will export the structure of a remotely defined datalink into a locally defined folder or under the local user's path. If sequence is specified, the export will create a local datalink of the specified version at the sequence point, but it will treat it as the latest for all other operations.

If the sync operation is denied or if the datalink referenced is not managed by the broker, the command will return an error: `Error: datalink is unknown to the broker`

### Handle

* the command expects its first argument to at least be a handle, but it may be a FQDN string composed of the OEM/HANDLE:VERSION fields, which are all required
* if the handle value does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error value [...] for option [datalink.sync.handle] is invalid`

### Oem

* if the oem value is specified as part of the first argument, it will override the option regardless of validity.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [datalink.sync.oem] is invalid`

### Version

* if the version value is specified as part of the first argument, it will override the option regardless of validity.
* if no version is specified the value will default to `1.0.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [datalink.sync.version] is invalid`

### Sequence

* sequence is entirely optional and will not be provided if is equivalent or equal to 0
* if the resolved version does not match a valid SemVer version string (see [sequence validity](index.md#sequence)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [datalink.sync.sequence] is invalid`

### config_folder

* if config_folder is specified, the command will look for a `Genaiz.yaml` under the specified path for synchronizing the exported link
* if config_folder is not specified and [user-defined](#user-defined) is false, the current working directory will be used for synchronizing the exported link
* if [user-defined](#user-defined) is specified, with no config_folder argument, the file written will be `$HOME/.config/genaiz/Genaiz.yaml`

### config-type

* if config-type is not specified, the default is yaml
* if the config-type specified is not a recognized value, the command will return an error with the field and the value: `value [...] for option [datalink.publish.configtype] is invalid`

> [!IMPORTANT]
> Only yaml is supported on the current early version of the SDK

### user-defined

* user-defined is set to `TRUE` by default, which will direct the command to read and write configuration files under the user's .config folder
* if user-defined is explicitly set to `FALSE`, the command will use [config-folder](#config_folder) or the current working directory to persist datalink modifications
* if a [config-folder](#config_folder) is provided, user-defined will be overridden to `FALSE`, regardless of whether it is specified or not
