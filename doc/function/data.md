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

* if the argument translates into an existing path, it will be compared with the value of `sf.run.input` and added only if it shares a common ancestor with the path and the last folder in the path is a valid handle.
* if the folder does not exist, the command will create the folder
* if the resolved port handle does not match a valid handle string (see [handle validity](../index.md#handle-and-oem)), the command will return an error: `Error: [...] is not a valid handle value`

##### name

* if the name is not specified it will default to the value of the port [handle](#portpath).
* if the resolved name does not match a valid name string (see [name validity](../index.md#name)), the command will return as error: `Error: value [...] for option [sf.publish.dataportadd.input.name] is invalid`

##### description

* description is optional and will be left empty if not specified
* if the resolved name does not match a valid name string (see [description validity](../index.md#description)), the command will return as error: `Error: value [...] for option [sf.publish.dataportadd.input.desc] is invalid`

#### data input rm

```
genaiz sf data input rm PORT|PATH
```

* if the argument translates into an existing path, it will be compared with the value of `sf.run.input` and removed only if it shares a common ancestor with the path and the last folder as the handle.
* the handle value is not validated. Removing something invalid produces a state that is valid, so no errors are raised.

### data output

Data output is used to add subfolders under the [run](run.md)'s output folder and identify the output ports for the Orchestration when [publishing](publish.md) a function, or a [solution](../solution/publish.md).

#### data output add

```
genaiz sf data output add PORT|PATH --name=NAME --description=DESC
```

##### PORT|PATH

* if the argument translates into an existing path, it will be compared with the value of `sf.run.output` and added only if it shares a common ancestor with the path and the last folder in the path is a valid handle.
* if the folder does not exist, the command will create the folder
* if the resolved port handle does not match a valid handle string (see [handle validity](../index.md#handle-and-oem)), the command will return an error: `Error: [...] is not a valid handle value`

##### name

* if the name is not specified it will default to the value of the port [handle](#portpath).
* if the resolved name does not match a valid name string (see [name validity](../index.md#name)), the command will return as error: `Error: value [...] for option [sf.publish.dataportadd.output.name] is invalid`

##### description

* description is optional and will be left empty if not specified
* if the resolved name does not match a valid name string (see [description validity](../index.md#description)), the command will return as error: `Error: value [...] for option [sf.publish.dataportadd.output.desc] is invalid`

#### data output rm

```
genaiz sf data output rm PORT|PATH
```

##### PORT|PATH

* if the argument translates into an existing path, it will be compared with the value of `sf.run.output` and removed only if it shares a common ancestor with the path and the last folder as the handle.
* the handle value is not validated. Removing something invalid produces a state that is valid, so no errors are raised.