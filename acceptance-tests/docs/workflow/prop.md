## Workflow Properties

Prop is the command used to manage property sets for workflow nodes. The property values are used to override the default values specified in the property specifications of a Smart Function under the workflow node, promoting re-usability without having to redefine a Smart Function to change default values.

>[!IMPORTANT]
>Valid property keys can take a serious amount of time to figure out, since keys are taken from multiple sources. For this reason the prop commands should be able to auto-complete the KEY fields below

### prop add

```
genaiz wf prop add WORKFLOW_HANDLE NODE_HANDLE|FUNCTION_PATH KEY VALUE \
  --no-prop-sync --no-validation
```

The command adds a property to the specified workflow on the specified node, by workflow handle, and either the node handle or the function path.

#### WORKFLOW_HANDLE

* if the workflow handle specified can not be found, the command will return an error: `Error: workflow hande [...] not found`

#### NODE_HANDLE

* if the specified handle does not match a valid handle string (see [handle validity](../index.md#handle-and-oem)), the command will return an error: `Error: [...] is not a valid node handle`
* if the specified handle does not exist under the specified [WORKFLOW_HANDLE](#workflow_handle), the command will return an error: `Error: node [...] is not a member of workflow [...]`

#### FUNCTION_PATH

* if the specified handle corresponds to a folder under which a Smart Function configuration can be found. The function's configurations will be used as default values for the command's SF options.
* if the specified handle corresponds to a folder under which a Smart Function configuration can be found, but is not a member node of the workflow, the command will return an error: `Error: smart function [...] is not a member of workflow [...]`
* if a function is found, the [NODE_HANDLE](#node_handle) value will be composed of the function's handle with added suffix: `-node`

#### KEY

* if the key specified is not a valid property for the smart function, or the node specified does not have a Smart Function, the command will run an error: `Error: the key [...] is invalid for node [...]`
* if the key specified already exists under the specified workflow and node, the command will return an error: `Error: the key [...] is already defined for node [...]`

#### VALUE

* if the value specified is not valid for the type associated with the [KEY](#key), the command will return an error: `Error: value [...] is not valid for key [...]`
* if the value passed is empty string, the command will return an error: `Error: empty string is not a valid value`

#### no-sync

* by default, the command will trigger a synchronization of all property specifications found on datalinks belonging to the [NODE_HANDLE](#node_handle) Smart Function, if any is specified.
* if no-sync is specified, the command will only use the prop specs found on locally known datalinks to validate the [KEY](#key).

#### no-validation

* if no-validation is specified, the command will not validate that the [KEY](#key) added is a valid property specification with the right type.
* this option does not disable duplicate checks.

### prop edit

```
genaiz wf prop edit WORKFLOW_HANDLE NODE_HANDLE|FUNCTION_PATH KEY VALUE \
  --no-prop-sync --no-validation
```

The command edits an existing property key specified for the provided workflow on the specified node, by workflow handle, and either the node handle or the function path. The command exists to avoid user errors such as repeatedly adding the same key with different values when the key should change.

#### WORKFLOW_HANDLE

* if the workflow handle specified can not be found, the command will return an error: `Error: workflow hande [...] not found`

#### NODE_HANDLE

* if the specified handle does not match a valid handle string (see [handle validity](../index.md#handle-and-oem)), the command will return an error: `Error: [...] is not a valid node handle`
* if the specified handle does not exist under the specified [WORKFLOW_HANDLE](#workflow_handle), the command will return an error: `Error: node [...] is not a member of workflow [...]`

#### FUNCTION_PATH

* if the specified handle corresponds to a folder under which a Smart Function configuration can be found. The function's configurations will be used as default values for the command's SF options.
* if the specified handle corresponds to a folder under which a Smart Function configuration can be found, but is not a member node of the workflow, the command will return an error: `Error: smart function [...] is not a member of workflow [...]`
* if a function is found, the [NODE_HANDLE](#node_handle) value will be composed of the function's handle with added suffix: `-node`

#### KEY

* if the key specified is not a valid property for the smart function, or the node specified does not have a Smart Function, the command will run an error: `Error: the key [...] is invalid for node [...]`
* if the key specified does not exist under the specified workflow and node, the command will return an error: `Error: the key [...] could not be found under node [...]`

#### VALUE

* if the value specified is not valid for the type associated with the [KEY](#key), the command will return an error: `Error: value [...] is not valid for key [...]`
* if the value passed is empty string, the command will return an error: `Error: empty string is not a valid value`

#### no-prop-sync

* by default, the command will trigger a synchronization of all property specifications found on datalinks belonging to the [NODE_HANDLE](#node_handle) Smart Function, if any is specified.
* if no-sync is specified, the command will only use the prop specs found on locally known datalinks to validate the [KEY](#key).

#### no-validation

* if no-validation is specified, the command will not validate that the [VALUE](#value) edited is valid for the type of its property specification.

### prop list

```
genaiz wf prop list WORKFLOW_HANDLE NODE_HANDLE|FUNCTION_PATH [KEY_STRING] \
  --no-prop-sync
```

The command executes the same functionality that tab should do when entering the KEY field under [add](#prop-add), [edit](#prop-edit) and [remove](#prop-rm)

#### WORKFLOW_HANDLE

* if the workflow handle specified can not be found, the command will return an error: `Error: workflow hande [...] not found`

#### NODE_HANDLE

* if the specified handle does not match a valid handle string (see [handle validity](../index.md#handle-and-oem)), the command will return an error: `Error: [...] is not a valid node handle`
* if the specified handle does not exist under the specified [WORKFLOW_HANDLE](#workflow_handle), the command will return an error: `Error: node [...] is not a member of workflow [...]`

#### FUNCTION_PATH

* if the specified handle corresponds to a folder under which a Smart Function configuration can be found. The function's configurations will be used as default values for the command's SF options.
* if the specified handle corresponds to a folder under which a Smart Function configuration can be found, but is not a member node of the workflow, the command will return an error: `Error: smart function [...] is not a member of workflow [...]`
* if a function is found, the [NODE_HANDLE](#node_handle) value will be composed of the function's handle with added suffix: `-node`

#### KEY_STRING

* if a string is given as a key, the command will attempt listing all the keys containing this string.

#### no-prop-sync

* by default, the command will trigger a synchronization of all property specifications found on datalinks belonging to the [NODE_HANDLE](#node_handle) Smart Function, if any is specified.
* if no-sync is specified, the command will only use the prop specs found on locally known datalinks to validate the [KEY](#key).

### prop rm

```
genaiz wf prop rm WORKFLOW_HANDLE NODE_HANDLE|FUNCTION_PATH KEY
```

The command removes an existing property key and its value for the provided workflow on the specified node, by workflow handle, and either the node handle or the function path.

#### WORKFLOW_HANDLE

* if the workflow handle specified can not be found, the command will return an error: `Error: workflow hande [...] not found`

#### NODE_HANDLE

* if the specified handle does not match a valid handle string (see [handle validity](../index.md#handle-and-oem)), the command will return an error: `Error: [...] is not a valid node handle`
* if the specified handle does not exist under the specified [WORKFLOW_HANDLE](#workflow_handle), the command will return an error: `Error: node [...] is not a member of workflow [...]`

#### FUNCTION_PATH

* if the specified handle corresponds to a folder under which a Smart Function configuration can be found. The function's configurations will be used as default values for the command's SF options.
* if the specified handle corresponds to a folder under which a Smart Function configuration can be found, but is not a member node of the workflow, the command will return an error: `Error: smart function [...] is not a member of workflow [...]`
* if a function is found, the [NODE_HANDLE](#node_handle) value will be composed of the function's handle with added suffix: `-node`

#### KEY

* if the key specified is not a valid property for the smart function, or the node specified does not have a Smart Function, the command will run an error: `Error: the key [...] is invalid for node [...]`
* if the key specified does not exist under the specified workflow and node, the command will succeed as the state of the node is still valid.
