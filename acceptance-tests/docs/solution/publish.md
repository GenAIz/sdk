## Solution Publish

```
genaiz sn publish --broker=BROKER --config-type=TYPE \ 
 --handle=HANDLE --oem=OEM --version=VERSION \ 
 --description=DESC --name=NAME 
```

Publish executes the necessary tasks to publish a new version of a Solution, in the current working directory,
provisioning and publishing all underlying Smart Functions used as Workflow Nodes.

### broker

* if the broker is not specified, publish will read the value from Genaiz.yaml under solution.publish.broker
* if no broker value can be found, the command will attempt publishing on the current active session (
  see [account login](../account/index.md#login))
* if there are no current session for the specified broker or no active sessions, publish returns `Error: not logged in`

### config-type

> [!CAUTION]
> The config type specified with publish will restrict smart functions read to the file type specified. This is a
> limitation of this preliminary version.

* if there is no config type specified, the default type will be `yaml`
* if the config type specified does not resolve to `yaml`, `json`, `toml` or `none`, the command will return an error
  with the key of the field and the invalid value:
  `Error: value [...] for option [solution.publish.configtype] is invalid`

### handle

The handle option of the create command is used to initialize the key `solution.handle`.

* if there is no handle specified, the value of `solution.handle` will be used.
* if the resolved handle does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the
  command will return an error with the key of the field and the invalid value:
  `Error value [...] for option [solution.publish.handle] is invalid`

### oem

* if the oem is not specified, the value of `solution.oem` will be used.
* if the resolved oem string does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the
  command will return an error with the key of the field and the invalid value;
  `Error: value [...] for option [solution.publish.oem] is invalid`

### description

* if the description is not specified, the value of `solution.description` will be used.
* if the resolved description evaluates to empty string, the value of [name](#name) will be used.
* if the resolved description string is too long (see [description validity](index.md#description)), the command will
  return an error with the key of the field and the invalid shortened value:
  `Error: value [...] for option [solution.publish.description] is invalid`

### name

* if the name is not specified, the value of `solution.name` will be used.
* if the resolved name evaluates to empty string, the value of [handle](#handle) will be used.
* if the resolved name string is too long (see [name validity](index.md#name)), the command will return an error with
  the key of the field and the invalid shortened value:
  `Error: value [...] for option [solution.publish.name] is invalid`

### version

* if the version is not specified, the value of `solution.version` will be used.
* if the resolved version evaluates to empty string, the default `1.0.0` is used.
* If the resolved version string does not match a valid version string (see [version validity](index.md#version)), the
  command will return an error with the key of the field and the invalid value;
  `Error: value [...] for option [solution.publish.version] is invalid`
