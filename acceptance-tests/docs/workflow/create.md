## Workflow Create

```
genaiz wf create HANDLE [SOLUTION_PATH] --config-type=TYPE \
    --name=NAME --description=DESC 
```

Workflow create is used to create a workflow under a specified solution path or the current working directory.

### handle

The handle of the workflow created will be written as the key `solution.workflows[n].handle`

* if the resolved handle does not match a valid handle string (see [handle validity](../index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error: value [...] is not a valid handle`

### config-type

> [!CAUTION]
> Only the yaml config type is supported by all commands at this time. Support for json and toml are under planning and none is under testing.

* if there is no config type specified, the default type will be `yaml`
* if the config type specified does not resolve to `yaml`, `json`, `toml` or `none`, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [workflow.create.configtype] is invalid`

### solution_path

The path of the solution under which to add the workflow.

* if the path specified does not exist, it will be created with an empty solution and a workflow will be sole element of this solution.
* if the path is not writeable, the command will return an Error: `Error open ...: permission denied`

### name

The name option of the create command is used to initialize the key `solution.workflows[n].name` of the workflow to create.

* if there is no name specified the value of [handle](#handle) will be used.
* if the name does not resolve to a valid name string (see [name validity](../index.md#name)), the command will return an error with the key of the field and the shortened invalid value: `Error: value [...] for option [workflow.create.name] is invalid`

### description

The name option of the create command is used to initialize the key `solution.workflows[n].name` of the workflow to create.

* if there is no name specified the value of [handle](#handle) will be used.
* if the name does not resolve to a valid name string (see [description validity](../index.md#description)), the command will return an error with the key of the field and the shortened invalid value: `Error: value [...] for option [workflow.create.description] is invalid`
