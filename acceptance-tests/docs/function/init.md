## Function Init

```
genaiz sf init --context=PATH --file=FILE --config-type=TYPE \
  --repository=LOCAL --handle=HANDLE --oem=OEM --version=VERSION \
  --type=TYPE --name=NAME --mount-in=/PATH-IN --mount-out=/PATH-OUT \
  --arch=x86 --arch=...
```

### context

* if there is no context specified the init command assumes it is established as the current working dir.
* the context is recorded under the function.build.context configuration and used by the [build](build.md) command as default value.
* if the resolved context does not correspond to an existing folder, the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [function.build.context] is invalid`

### file

Initializes the key `function.build.file` under the function's `Genaiz.yaml` configuration.

* if there is no file specified, the default `Dockerfile` will be used by the [build](build.md) command. No file configurations will be retained under Genaiz.yaml
* if the file specified does not resolve to an existing file, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [function.build.file] is invalid`

### config-type

> [!CAUTION]
> Only the yaml config type is supported by all commands at this time. Support for json and toml are under planning and none is under testing.

* if there is no config type specified and the parent folder contains a valid solution file named `Genaiz`, the config-type is the one associated with this file
* if there is no config type specified and there is no parent folder containing a solution file, the config type is set to yaml by default
* if the config type specified does not resolve to `yaml`, `json`, `toml` or `none`, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [function.init.configtype] is invalid`

### repository

Initializes the key `function.build.repository` under the function's `Genaiz.yaml` configuration.

* if there is no repository specified, the value will be made from `<oem>/<handle>`
* if the repository specified does not match a valid repository string (see [repository validity](index.md#repository)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [function.build.repository] is invalid`

### handle

The handle option of the init command is used to initialize the key `function.publish.handle` and part of the key `function.build.repository` if not specified.

* if there is no handle specified, the name of the context folder will be used by default. If the function is held inside a folder myFunction, the handle will be myFunction
* if the resolved handle does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error value [...] for option [function.init.handle] is invalid`

### oem

The oem option of the init command is used to initialize the key `function.publish.oem` and part of the key `function.build.repository` if not specified.

* if there is no oem specified, and the parent folder has a valid solution configuration, the oem value will default to the solution value.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [function.init.oem] is invalid`

### version

The version option of the init command is used to initialize the key `function.publish.version`. This version is used strictly when using the [publish](publish.md) command. The key `function.build.version` will always default to `latest` when using init.

* if there is no version specified, and the parent folder has a valid solution configuration, the version value will default to the solution value.
* if there is no version specified, and the parent folder does not contain a valid solution configuration, the version value will default to `0.1.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [function.init.version] is invalid`

### type

The type option of the init command is used to initialize the key `function.publish.type`.

* if the type option is not specified `function` is used by default.
* if the type does not resolve to a valid type string (see [type validity](index.md#type)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [function.init.type] is invalid`

### name

The name option of the init command is used to initialize the key `function.publish.name`.

* if there is no name specified, the command will use the value of [handle](#handle) as the default value.
* if the name does not resolve to a valid name string (see [name validity](index.md#name)), the command will return an error with the key of the field and the shortened invalid value: `Error: value [...] for option [function.init.name] is invalid`

### mount-in

The mount-in option of the init command is used to initialize the key `function.run.input`.

* There is no validation applied to the string passed here, since the folder may not have been created and is only used the [run](run.md) command.
* It is expected that the value will eventually map to existing folder

### mount-out

The mount-out option of the init command is used to initialize the keys `function.run.ouput`, `function.run.log` and `function.run.var`.

* There is no validation applied to the string passed here, since the folder may not have been created and is only used the [run](run.md) command.
* It is expected that the value will eventually map to existing folder.
* The values of `function.run.log` and `function.run.var` will be made sub-folders of `function.run.output` on `OUTPUT/log` and `OUTPUT/var` respectively

### arch

The arch option list of the init command is used to initialize the key `function.publish.arches` which list all target architectures for the function.

* multiple options can be specified on the command line by repeating the option --arch or by passing CSV compatible values: `--arch=x86,"arm",arm64`. Double quotes are optional
* architectures must match one of **"x86", "x86_64", "arm" and "arm64"**, case-sensitive
* if the list of resolved architectures has any invalid values, the command will return an error with the key of the field, the list index and the invalid value; `Error: value [n:...] for option [function.init.arch] is invalid`
