## Function Create

```bash
genaiz sf create FOLDER --context=PATH --file=FILE --config-type=TYPE \
  --tag=LOCAL --handle=HANDLE --oem=OEM --version=VERSION \
  --recipe=RECIPE --type=TYPE --name=NAME \
  --mount-in=/PATH-IN --mount-out=/PATH-OUT --arch=x86 --arch=...
```

### context & folder

The context option on the create command is used to change the working dir before creating the folder that will hold the smart function. This matters, if the solution file is in a different location than the folder where we want to create the function.

* if a folder with the specified name already exists under the specified context, the command will return an error: `Error: smart function [...] already exist`
* if there is no context specified the create command assumes it is creating the smart function folder from the current working dir.
* if the resolved context does not correspond to an existing folder, the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.build.context] is invalid`

### file

> [!CAUTION]
> The file option is not currently implemented and should not do anything for the create command. 

The file option on the create command is used to change the name of any recipe Dockerfile created under the Smart Function folder.

* if the resolved file name does not match a valid file name, without spacing, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [sf.build.file] is invalid`

### config-type

> [!CAUTION]
> Only the yaml config type is supported by all commands at this time. Support for json and toml are under planning and none is under testing.

* if there is no config type specified and the context folder contains a valid solution file named `Genaiz`, the config-type is the one associated with this file
* if there is no config type specified and there is no context folder containing a solution file, the config type is set to yaml by default
* if the config type specified does not resolve to `yaml`, `json`, `toml` or `none`, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [sf.create.configtype] is invalid`

### tag

Initializes the key `sf.build.tag` under the function's `Genaiz.yaml` configuration.

* if there is no tag specified, the value will be made from `<oem>/<handle>`
* if the tag specified does not match a valid repository string (see [repository validity](index.md#repository)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.build.tag] is invalid`

### handle

The handle option of the create command is used to initialize the key `sf.publish.handle` and part of the key `sf.build.tag` if not specified.

* if there is no handle specified, the name of the Smart Function folder will be used by default. If the function is held inside a folder myFunction, the handle will be myFunction
* if the resolved handle does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error value [...] for option [sf.create.handle] is invalid`

### oem

The oem option of the create command is used to initialize the key `sf.publish.oem` and part of the key `sf.build.tag` if not specified.

* if there is no oem specified, and the context folder has a valid solution configuration, the oem value will default to the solution value.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [sf.create.oem] is invalid`

### version

The version option of the create command is used to initialize the key `sf.publish.version`. This version is used strictly when using the [publish](publish.md) command. The key `sf.build.version` will always default to `latest` when using init.

* if there is no version specified, and the context folder has a valid solution configuration, the version value will default to the solution value.
* if there is no version specified, and the context folder does not contain a valid solution configuration, the version value will default to `0.1.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [sf.create.version] is invalid`

### recipe

> [!CAUTION]
> Recipes are not fully implemented, only bash-example is available at this time, and the genaiz recipe (rp) command is still under design.

The recipe option of the create command is used to initialize the content of the created Smart Function folder.

* if the recipe option is not specified, the command will create an empty Smart Function folder with a `Genaiz` configuration file
* if the recipe specified does not exist, the command will return the error `Error: recipe not found`

### type

The type option of the create command is used to initialize the key `sf.publish.type`.

* if the type option is not specified `function` is used by default.
* if the type does not resolve to a valid type string (see [type validity](index.md#type)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.create.type] is invalid`

### name

The name option of the create command is used to initialize the key `sf.publish.name`.

* if there is no name specified, the command will use the value of [handle](#handle) as the default value.
* if the name does not resolve to a valid name string (see [name validity](index.md#name)), the command will return an error with the key of the field and the shortened invalid value: `Error: value [...] for option [sf.create.name] is invalid`

### mount-in

> [!WARNING]
> The value of mount-in will be copied under the Genaiz.yaml file, but any recipe selected may or may not respect its creation. 

The mount-in option of the create command is used to initialize the key `sf.run.input`.

* There is no validation applied to the string passed here, since the folder may not have been created and is only used by the [run](run.md) command.
* It is expected that the value will eventually map to existing folder

### mount-out

> [!WARNING]
> The value of mount-out will be copied under the Genaiz.yaml file, but any recipe selected may or may not respect its creation.

The mount-out option of the create command is used to initialize the keys `sf.run.ouput`, `sf.run.log` and `sf.run.var`.

* There is no validation applied to the string passed here, since the folder may not have been created and is only used by the [run](run.md) command.
* It is expected that the value will eventually map to existing folder.
* The values of `sf.run.log` and `sf.run.var` will be made sub-folders of `sf.run.output` on `OUTPUT/log` and `OUTPUT/var` respectively

### arch

The arch option list of the create command is used to initialize the key `sf.publish.arches` which list all target architectures for the function.

* multiple options can be specified on the command line by repeating the option --arch or by passing CSV compatible values: `--arch=x86,"arm",arm64`. Double quotes are optional
* architectures must match one of **"x86", "x86_64", "arm" and "arm64"**, case-sensitive
* if the list of resolved architectures has any invalid values, the command will return an error with the key of the field, the list index and the invalid value; `Error: value [n:...] for option [sf.create.arch] is invalid`