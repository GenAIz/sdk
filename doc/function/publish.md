## Function Publish

```bash
genaiz sf publish --broker=HOST \
--handle=HANDLE --oem=OEM --version=VERSION --name=NAME --type=TYPE \
--arch=x86 --arch=... --no-update --rebuild
 ```
Publish executes the necessary tasks to provision and publish a new version of a Smart Function, in the current working directory, optionally rebuilding the function and updating its local configurations.

A successful command will print the oem, handle and version that were accepted by the broker. It should also print which configuration file was updated if --no-update wasn't used.

### broker

* if the broker is not specified, publish will read the value from Genaiz.yaml under solution.publish.broker
* if no broker value can be found, the command will attempt publishing on the current active session (see [account login](../account/index.md#login))
* if there are no current session for the specified broker or no active sessions, publish returns `Error: not logged in`

### handle

* if the handle is not specified, publish will read the value from Genaiz.yaml under sf.publish.handle
* if no value can be found, publish will default handle to the name of the working dir
* if the resolved handle string does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.publish.handle] is invalid`

### oem

* if the oem is not specified, publish will read the value from Genaiz.yaml under sf.publish.oem
* if the oem is not specified under the Smart Function Genaiz.yaml file, publish will retrieve the parent's solution oem
* if the resolved oem string does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.publish.oem] is invalid`

### version

* if the version is not specified, publish will read the value from Genaiz.yaml under sf.publish.version
* if the version is not specified under the Smart Function Genaiz.yaml file, publish will retrieve the parent's solution version
* if the version could not be retrieved from the parent solution, 0.1.0 is used as the default string
* If the resolved version string does not match a valid version string (see [version validity](index.md#version)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.publish.version] is invalid`

### name

* if the name is not specified, publish will read the value from Genaiz.yaml under sf.publish.name
* if the name is not specified under the Smart Function Genaiz.yaml file, publish will use the [Handle](#handle) value specified above.
* if the resolved name does not match a valid name string (see [name validity](index.md#name)), the command will return an error with the key of the field and the invalid value, shortened to **32 characters**; `Error: value [...] for option [sf.publish.name] is invalid`

### type

* if the type is not specified publish will read the value from Genaiz.yaml under sf.publish.type
* if the resolved type does not match a valid type option (see [type validity](index#type)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.publish.type] is invalid`

### arch

* architecture values are optional and currently will be ignored by the broker.
* multiple options can be specified on the command line by repeating the option --arch or by passing CSV compatible values: `--arch=x86,"arm",arm64`. Double quotes are optional
* architectures must match one of **"x86", "x86_64", "arm" and "arm64"**, case-sensitive
* if the list of resolved architectures has any invalid values, the command will return an error with the key of the field, the list index and the invalid value; `Error: value [n:...] for option [sf.publish.arch] is invalid`

### no-update

By default, the command will update values in the Genaiz.yaml file of the function; Publish will skip its init task if `--no-update` is used

### rebuild

By default, the command will assume the docker was built and is available for publishing; Publish will call the [build](build.md) task of the function if `--rebuild` is used