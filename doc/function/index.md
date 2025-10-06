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

* Can only have letters (upper or lower case), digits, dots, dashes and underscores.
* Can only start with a letter (upper or lower case) or a digit.
* Can not have 2 consecutive characters of type dots, dashes and underscores.
* Can not end with a character of type dot, dash or underscore.

### Name

* A name can have any characters.
* The length must not exceed 255 characters.

### Repository

* A repository is a combination of [Oem](#handle-and-oem), namespace components and a [Handle](#handle-and-oem).
* Only lower cases letters are accepted by registries, but the SDK will lower all upper case characters by default.
* Valid namespace components are in the same format as a handle, separated by a `/` character

### Type

* Must be either **FUNCTION**, **TRIGGER** or **CONNECTOR** in lower or upper case characters.

### Version

* Must be a valid [Semantic Version](https://semver.org/) string.
* Can not contain pre-release version identifiers as those are reserved by the broker.
