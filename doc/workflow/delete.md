## Workflow Delete

```
genaiz wf delete HANDLE --config-type=TYPE
```

Workflow delete is a simple command which can be used to remove a workflow from a solution. It can only remove workflows from the current working directory's solution if any is present.

### handle

* if the config file can not be read, the command returns an error: `Error: workflow config file not found`
* if the handle does not exist under the current configuration, the command does not fail, returning normally with the handle reported removed.
  * a non-existing workflow will always be considered removed if it wasn't existing to begin.

### config-type

> [!CAUTION]
> Only the yaml config type is supported by all commands at this time. Support for json and toml are under planning and none is under testing.

* if there is no config type specified, the default type will be `yaml`
* if the config type specified does not resolve to `yaml`, `json`, `toml` or `none`, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [workflow.delete.configtype] is invalid`