## Solution Create

```bash
genaiz sn create [FOLDER] --config-type=TYPE --handle=HANDLE --oem=OEM \
  --description=DESC --name=NAME --version=VERSION \
  --workflow-desc=DESC --workflow-handle=HANDLE --workflow-name=NAME
```

Solution create is used to create a solution folder or transform an existing folder without a solution into a GenAIz Solution. Solutions are the containers of Smart Functions and Workflows.

### folder

* The folder path will be created if it doesn't exist and if write permissions allow. Otherwise, the command returns an error: `Error: mkdir ...: permission denied`
* If the folder path is not provided, the command creates a solution in the current working dir.

### config-type

> [!CAUTION]
> Only the yaml config type is supported by all commands at this time. Support for json and toml are under planning and none is under testing.

* if there is no config type specified, the default type will be `yaml`
* if the config type specified does not resolve to `yaml`, `json`, `toml` or `none`, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [solution.create.configtype] is invalid`

### handle

The handle option of the create command is used to initialize the key `solution.handle`.

* if there is no handle specified, the name of the current working dir will be used. 
* if the resolved handle does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error value [...] for option [solution.create.handle] is invalid`

### oem

The oem option of the create command is used to initialize the key `solution.oem`.

* if there is no oem specified, and the context folder has a valid solution configuration, the **solution will be initialized without a valid OEM value**.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [solution.create.oem] is invalid`

> [!TIP]
> With an empty OEM the [publish](publish.md) command will fail, but the solution file is not the only way to specify OEM. The key `solution.ceate.oem`, for all create commands, can be specified globally under ~/.config/genaiz/Genaiz.yaml

> [!TIP]
> The key `solution.oem` specified under the same file should apply to all solutions as well.

### description

The description option of the create command is used to initialize the key `solution.handle`.

* if there is no description specified, the value of [handle](#handle) will be used.
* if the resolved description is too long, (see [description validity](../index.md#description)), the command will return an error with the key of the field and the invalid shortened value: `Error value [...] for option [solution.create.description] is invalid`

### name

The name option of the create command is used to initialize the key `solution.name`.

* if there is no name specified the value of [handle](#handle) will be used.
* if the name does not resolve to a valid name string (see [name validity](index.md#name)), the command will return an error with the key of the field and the shortened invalid value: `Error: value [...] for option [solution.create.name] is invalid`

### version

The version option of the create command is used to initialize the key `solution.version`. A solution version is always required, even on create.

* if there is no version specified, the version value will default to `0.1.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [solution.create.version] is invalid`

### workflow-desc

The workflow-desc option used to populate the description field of the first workflow defined for the created solution. All solutions must have at least 1 workflow. The option will initialize the node `solution.workflows[0].description`.

* if there is no workflow description provided, `default workflow` will be used.
* if the resolved description is too long, (see [description validity](../index.md#description)), the command will return an error with the key of the field and the invalid shortened value: `Error value [...] for option [solution.create.workflow.description] is invalid`

### workflow-handle

The workflow-handle option used to populate the handle field of the first workflow defined for the created solution. All solutions must have at least 1 workflow. The option will initialize the node `solution.workflows[0].handle`.

* if there is no workflow handle provided, `default` will be used.
* if the resolved handle does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error value [...] for option [solution.create.workflow.handle] is invalid`

### workflow-name

The workflow-name option used to populate the name field of the first workflow defined for the created solution. All solutions must have at least 1 workflow. The option will initialize the node `solution.workflows[0].name`.

* if there is no workflow name provided, `Default Workflow` will be used.
* if the name does not resolve to a valid name string (see [name validity](index.md#name)), the command will return an error with the key of the field and the shortened invalid value: `Error: value [...] for option [solution.create.workflow.name] is invalid`
