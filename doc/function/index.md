# Smart Function Command Specs

## Test Cases

* [Build bash example](build_bash_example.feature)
* [Publish bash example](publish_bash_example.feature)

## Commands

* [build](build.md)
* [create](create.md)
* [debug](debug.md)
* [init](init.md)
* [list](list.md)
* [publish](publish.md)
* [run](run.md)
* [start](start.md)
* [stop](stop.md)
* [test](test.md)

## Validation

### Handle and OEM

See [Global Validation](../index.md#handle-and-oem)

### Name

See [Global Validation](../index.md#name)

### Repository

* A repository is a combination of [Oem](#handle-and-oem), namespace components and a [Handle](#handle-and-oem).
* Only lower cases letters are accepted by registries, but the SDK will lower all upper case characters by default.
* Valid namespace components are in the same format as a handle, separated by a `/` character

### Type

* Must be either **FUNCTION**, **TRIGGER** or **CONNECTOR** in lower or upper case characters.

### Version

See [Global Validation](../index.md#version)
