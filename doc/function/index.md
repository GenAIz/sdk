# Smart Function Command Specs

## Test Cases

* [Build bash example](build_bash_example.feature)
* [Create bash example](create_bash_example.feature)
* [Data bash example](data_bash_example.feature)
* [Data source example](data_source_example.feature)
* [Data store example](data_store_example.feature)
* [Init empty example](init_empty_example.feature)
* [List bash example](list_bash_example.feature)
* [Prop bash example](prop_bash_example.feature)
* [Publish bash example](publish_bash_example.feature)
* [Run bash example](run_bash_example.feature)
* [Start bash example](start_bash_example.feature)
* [Stop bash example](stop_bash_example.feature)
* [Test bash example](test_bash_example.feature)

## Commands

* [build](build.md)
* [create](create.md)
* [data](data.md)
* [init](init.md)
* [list](list.md)
* [prop](prop.md)
* [publish](publish.md)
* [run](run.md)
* [start](start.md)
* [stop](stop.md)
* [test](test.md)

## Environment:

The [run](run.md), [start](start.md) and [test](test.md) commands should normally honor the following environment variables, forwarding them to the containers they create:

There should also be a mechanism for specifying environment variables using a .env file and be able to specify the path of such file if it can not be found under the resolved Docker context.

### SF_INPUT_PATH

- default is /mnt/in
- this is the input path -- will be read-only -- each port defined in the SF has its own data-set

### SF_OUTPUT_PATH

- default is /mnt/out
- this is the output path -- read/write -- each port defined in the SF should target a directory here

### SF_LOG_PATH

- default is /mnt/log
- you can have separate log files for different purposes -- totally optional

### SF_VAR_PATH

- default is /mnt/var
- just like the unix directory /var, it's meant to be variables -- read/write

### SF_PROGRESS_FILE

- default is SF_VAR_PATH + "/progress"
- output a number between 0 and 100 and that progress will show up on the orchestrator (yes, validation exists) -- totally optional

### SF_RESULT_FILE

- default is SF_VAR_PATH + "/result"
- simple result string -- it allows the workflow to use simple branch conditions -- totally optional

### SF_STATUS_FILE

- default is SF_VAR_PATH + "/status"
- status of the SF -- mandatory -- you need to write 'SUCCESS' in there before exiting, or else we consider that the function failed

### SF_TYPE

- the type of smart function (CONNECTOR or FUNCTION or TRIGGER)
- this should be provided by default by the SDK relying on the type field specified under the function publish configurations

## Validation

### FQDNV

See [Global Validation](../index.md#fqdnv)

### Handle and OEM

See [Global Validation](../index.md#handle-and-oem)

### Name

See [Global Validation](../index.md#name)

### Component

* A component is a string without white spaces which accepts letters, digits, dashes, dots and underscores.
* Components can not have 2 consecutive non-alphanumeric characters.
* Components must start with an alphanumeric character and end with an alphanumeric.

### Repository

* A repository is a combination of [Oem](#handle-and-oem), namespace components and a [Handle](#handle-and-oem).
* Only lower cases letters are accepted by registries, but the SDK will lower all upper case characters by default.
* Valid namespace components are in the same format as a handle, separated by a `/` character

### Type

* Must be either **FUNCTION**, **TRIGGER** or **CONNECTOR** in lower or upper case characters.

### Version

See [Global Validation](../index.md#version)
