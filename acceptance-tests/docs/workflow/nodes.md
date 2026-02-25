## Workflow Nodes

Nodes is a command used to manage workflow nodes on solutions to be published. A solution can be viewed as a set of workflows each dictating a workflow of functions or repositories.

With the current design of the SDK a workflow node can be any Smart Function. Smart Functions which are not contained within a solution can be referenced as nodes.

> [!IMPORTANT]
> The SDK does not validate whether the Smart Function of a node exists or not, this is deferred to [solution publish](../solution/publish.md) and the Orchestrator used.

### nodes add

```
genaiz wf nodes add WORKFLOW_HANDLE [NODE_HANDLE|FUNCTION_PATH] \
  --config-type=TYPE --description=DESC --name=NAME \
  --sf=OEM/HANDLE:VERSION[-|.]SEQ \
  --sf-handle=HANDLE --sf-oem=OEM --sf-seq=SEQ --sf-version=VERSION
```

A workflow node can be added with a Smart Function contained under the parent solution or without any Smart Function specified.

#### WORKFLOW_HANDLE

* if the workflow handle specified can not be found, the command will return an error: `Error: workflow hande [...] not found`

#### NODE_HANDLE

* if the specified handle does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command will return an error: `Error: [...] is not a valid node handle`
* if the specified handle already exists under the specified [WORKFLOW_HANDLE](#workflow_handle), the command will return an error: `Error: the node specified already exists `

#### FUNCTION_PATH

* if the specified handle corresponds to a folder under which a Smart Function configuration can be found. The function's configurations will be used as default values for the command's SF options.
* if a function is found, the [NODE_HANDLE](#node_handle) value will be composed of the function's handle with added suffix: `-node`

#### config-type

* if the config-type is not specified, the current work dir will be read for a file named `Genaiz`, the type will be inferred by the file's extension.
* if the config-type can not be inferred from any `Genaiz` file in the current work dir, the command will return an error: `Error: could not find local config [Genaiz] under [...]`

#### description

* if description is not specified, the empty value will be used under the added node
* if the resolved description is too long, (see [description validity](../index.md#description)), the command will return an error with the key of the field and the invalid shortened value: `Error value [...] for option [workflow.nodes.add.description] is invalid`

#### name

* if name is not specified, it will default to [NODE_HANDLE](#node_handle) or [FUNCTION_PATH](#function_path) with a `Node` suffix
* if the name does not resolve to a valid name string (see [name validity](index.md#name)), the command will return an error with the key of the field and the shortened invalid value: `Error: value [...] for option [workflow.nodes.add.name] is invalid`

#### sf

The option provides a serialized value of a function within the workflow solution. It does not have precedence over [sf-oem](#sf-oem), [sf-handle](#sf-handle), [sf-seq](#sf-seq) and [sf-version](#sf-version), nor over the values retrieved from [FUNCTION_PATH](#function_path).

* sf requires the same validity checks as [sf-oem](#sf-oem), [sf-handle](#sf-handle), [sf-seq](#sf-seq) and [sf-version](#sf-version)
* if sf is specified with a valid [FUNCTION_PATH](#function_path), the command will return an Error with a field conflict if the values are different: `Error: value [...] for option [...] conflicts with [...] under [...]`

#### sf-oem

* the value of sf-oem has precedence over the [sf](#sf) serialized oem
* if sf-oem is specified with a valid [FUNCTION_PATH](#function_path), and the values are different, the command will return an error with a field conflict: `Error: value [...] for option [workflow.nodes.add.oem] conflicts with [...] under [...]`
* if sf-oem is specified without a sf-handle and or sf-version, the command will return an error: `Error: incomplete smart function specification, minimum required: OEM/HANDLE:VERSION`
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [workflow.nodes.add.oem] is invalid`

#### sf-handle

* the value of sf-handle has precedence over the [sf](#sf) serialized handle
* if sf-handle is specified with a valid [FUNCTION_PATH](#function_path), and the values are different, the command will return an error with a field conflict: `Error: value [...] for option [workflow.nodes.add.handle] conflicts with [...] under [...]`
* if sf-handle is specified without a sf-handle and or sf-version, the command will return an error: `Error: incomplete smart function specification, minimum required: OEM/HANDLE:VERSION`
* if the resolved handle does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [workflow.nodes.add.handle] is invalid`

#### sf-seq

* the value of sf-seq has precedence over the [sf](#sf) serialized seq
* if sf-seq is not specified, the Smart Function version will be omitted
* if the resolved seq does not match a valid sequence number (see [sequence validity](index.md#sequence)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [workflow.nodes.add.seq] is invalid`

#### sf-version

* the value of sf-version has precedence over the [sf](#sf) serialized version
* if sf-version with or without sf-seq is specified with a valid [FUNCTION_PATH](#function_path), and the concatenated values are different from the function version, the command will return an error with a field conflict: `Error value [...] for option [workflow.nodes.add.version] conflicts with [...] under [...]`
* if sf-version is specified without a sf-oem and or sf-handle, the command will return an error: `Error: incomplete smart function specification, minimum required: OEM/HANDLE:VERSION`
* if the resolved version does not match a valid version (see [version validity](index.md#version)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [workflow.nodes.add.version] is invalid`

### nodes remove (rm)

```
genaiz wf nodes rm WORKFLOW_HANDLE NODE_HANDLE [NODE_HANDLE...] \
  --config-type=TYPE
```

Workflow nodes can be removed from an existing workflow, if the handle matches a node. Removing a node that does not exist is not an Error.

#### WORKFLOW_HANDLE

* if the workflow handle specified can not be found, the command will return an error: `Error: workflow hande [...] not found`

#### NODE_HANDLE

* multiple values can be specified after the first node handle
* if no node handles could be removed, the command succeeds with a no-op on the selected configuration

#### config-type

* if the config-type is not specified, the current work dir will be read for a file named `Genaiz`, the type will be inferred by the file's extension.
* if the config-type can not be inferred from any `Genaiz` file in the current work dir, the command will return an error: `Error: could not find local config [Genaiz] under [...]`