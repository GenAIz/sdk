## Function Data

Data is a command for managing the Data Ports of a Smart Function. These configurations are used to identify data location when a Function runs inside a workflow. Ports are used by [Workflow Links](../workflow/links.md).

There are 2 type of Data Ports supported by the Orchestration: Input and Output Ports. Each corresponding to data taken in and data written out, when a function runs.

### data input

Data input is used to add subfolders under the [run](run.md)'s input folder and identify the input ports for the Orchestration when [publishing](publish.md) a function, or a [solution](../solution/publish.md).

#### data input add

```
genaiz sf data input add PORT|PATH --name=NAME --description=DESC
```

##### PORT|PATH

* if the argument translates into an existing path, it will be compared with the value of `function.run.input` and added only if it shares a common ancestor with the path and the last folder in the path is a valid handle.
* if the folder does not exist, the command will create the folder
* if the resolved port handle does not match a valid handle string (see [handle validity](../index.md#handle-and-oem)), the command will return an error: `Error: [...] is not a valid handle value`

##### name

* if the name is not specified it will default to the value of the port [handle](#portpath).
* if the resolved name does not match a valid name string (see [name validity](../index.md#name)), the command will return as error: `Error: value [...] for option [function.publish.dataportadd.input.name] is invalid`

##### description

* description is optional and will be left empty if not specified
* if the resolved name does not match a valid name string (see [description validity](../index.md#description)), the command will return as error: `Error: value [...] for option [function.publish.dataportadd.input.desc] is invalid`

#### data input rm

```
genaiz sf data input rm PORT|PATH
```

* if the argument translates into an existing path, it will be compared with the value of `function.run.input` and removed only if it shares a common ancestor with the path and the last folder as the handle.
* the handle value is not validated. Removing something invalid produces a state that is valid, so no errors are raised.

### data output

Data output is used to add subfolders under the [run](run.md)'s output folder and identify the output ports for the Orchestration when [publishing](publish.md) a function, or a [solution](../solution/publish.md).

#### data output add

```
genaiz sf data output add PORT|PATH --name=NAME --description=DESC
```

##### PORT|PATH

* if the argument translates into an existing path, it will be compared with the value of `function.run.output` and added only if it shares a common ancestor with the path and the last folder in the path is a valid handle.
* if the folder does not exist, the command will create the folder
* if the resolved port handle does not match a valid handle string (see [handle validity](../index.md#handle-and-oem)), the command will return an error: `Error: [...] is not a valid handle value`

##### name

* if the name is not specified it will default to the value of the port [handle](#portpath).
* if the resolved name does not match a valid name string (see [name validity](../index.md#name)), the command will return as error: `Error: value [...] for option [function.publish.dataportadd.output.name] is invalid`

##### description

* description is optional and will be left empty if not specified
* if the resolved name does not match a valid name string (see [description validity](../index.md#description)), the command will return as error: `Error: value [...] for option [function.publish.dataportadd.output.desc] is invalid`

#### data output rm

```
genaiz sf data output rm PORT|PATH
```

##### PORT|PATH

* if the argument translates into an existing path, it will be compared with the value of `function.run.output` and removed only if it shares a common ancestor with the path and the last folder as the handle.
* the handle value is not validated. Removing something invalid produces a state that is valid, so no errors are raised.

### data source

Data source is used to add data link addresses, as [FQDNV](index.md#fqdnv), to the list of data sources. A data source listed here a Smart Function may be used to establish a **Read-Only** connection while running the Smart Function.

Data sources are only configurable on Connectors, the commands listed under will return an error if used on a Smart Function of another type: `Error: data links can only be configured for connector functions`

#### data source add

```
genaiz sf data source add [OEM/][HANDLE][:VERSION] \
  --oem=OEM --handle=HANDLE --version=VERSION --no-validation
```

When invoking data source add, unless the [no-validation](#no-validation) option is used, the command will attempt verifying that the user has access to the specified data link, by querying the broker by oem and handle for a list of corresponding data links.

If the version specified is not available, the command will return an error with a list of available data links if any were returned.

##### Oem

* needs to be specified with the FQDNV argument or with the oem option.
* if the value of oem does not match a valid oem string, (see [oem validity](../index.md#handle-and-oem)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datasourceadd.oem] is invalid`

##### Handle

* needs to be specified with the FQDNV argument or with the handle option.
* if the value of handle does not match a valid handle string, (see [handle validity](../index.md#handle-and-oem)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datasourceadd.handle] is invalid`

##### Version

* needs to be specified with the FQDNV argument or with the version option.
* if the value of version does not match a valid version string, (see [version validity](../index.md#version)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datasourceadd.version] is invalid`

##### no-validation

* if specified, the command will not attempt verifying the validity of the FQDNV address added to the data source list.

#### data source rm

```
genaiz sf data source rm [OEM/][HANDLE][:VERSION] \
  --oem=OEM --handle=HANDLE --version=VERSION
```

##### Oem

* needs to be specified with the FQDNV argument or with the oem option.
* if the value of oem does not match a valid oem string, (see [oem validity](../index.md#handle-and-oem)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datasourceremove.oem] is invalid`

##### Handle

* needs to be specified with the FQDNV argument or with the handle option.
* if the value of handle does not match a valid handle string, (see [handle validity](../index.md#handle-and-oem)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datasourceremove.handle] is invalid`

##### Version

* needs to be specified with the FQDNV argument or with the version option.
* if the value of version does not match a valid version string, (see [version validity](../index.md#version)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datasourceremove.version] is invalid`

### data store

Data store is used to add data link addresses, as [FQDNV](index.md#fqdnv), to the list of data stores. A data store listed under a Smart Function may be used to establish a **Read/Write** connection while running the Smart Function.

Data stores are only configurable on Connectors, the commands listed here will return an error if used on a Smart Function of another type: `Error: data links can only be configured for connector functions`

#### data store add

```
genaiz sf data store add [OEM/][HANDLE][:VERSION] \
  --oem=OEM --handle=HANDLE --version=VERSION --no-validation
```

When invoking data store add, unless the [no-validation](#no-validation) option is used, the command will attempt verifying that the user has access to the specified data link, by querying the broker by oem and handle for a list of corresponding data links.

If the version specified is not available, the command will return an error with a list of available data links if any were returned.

##### Oem

* needs to be specified with the FQDNV argument or with the oem option.
* if the value of oem does not match a valid oem string, (see [oem validity](../index.md#handle-and-oem)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datastoreadd.oem] is invalid`

##### Handle

* needs to be specified with the FQDNV argument or with the handle option.
* if the value of handle does not match a valid handle string, (see [handle validity](../index.md#handle-and-oem)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datastoreadd.handle] is invalid`

##### Version

* needs to be specified with the FQDNV argument or with the version option.
* if the value of version does not match a valid version string, (see [version validity](../index.md#version)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datastoreadd.version] is invalid`

##### no-validation

* if specified, the command will not attempt verifying the validity of the FQDNV address added to the data store list.

#### data store rm

```
genaiz sf data store rm [OEM/][HANDLE][:VERSION] \
  --oem=OEM --handle=HANDLE --version=VERSION
```

##### Oem

* needs to be specified with the FQDNV argument or with the oem option.
* if the value of oem does not match a valid oem string, (see [oem validity](../index.md#handle-and-oem)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datastoreremove.oem] is invalid`

##### Handle

* needs to be specified with the FQDNV argument or with the handle option.
* if the value of handle does not match a valid handle string, (see [handle validity](../index.md#handle-and-oem)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datastoreremove.handle] is invalid`

##### Version

* needs to be specified with the FQDNV argument or with the version option.
* if the value of version does not match a valid version string, (see [version validity](../index.md#version)), the command will return an error with the field and the invalid value: `value [...] for option [function.publish.datastoreremove.version] is invalid`
