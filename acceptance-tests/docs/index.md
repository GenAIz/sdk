# SDK Feature Docs

The concepts and design related to the Orchestrator and GenAIz platform should be available through their own respective projects. This document serves the purpose of exposing how the main entry point of the SDK, its CLI, is behaving.

The usage documentation is established with set of rules assigned to actors implicated on a particular feature. Features are described plainly in [Gherkin](https://cucumber.io/docs/gherkin/) format.

> [!NOTE]
> Historically the features are established like a series of use cases describing the behavior of a user. 

> [!NOTE]
> Since the project is not handled by Cucumber, step definition should be/is defined and compiled in the [genaiz-it](../../genaiz-it/README.md) modules

## [Accounts](account/index.md)
* [genaiz account login](account/index.md#login)
* [genaiz account logout](account/index.md#logout)

## [Smart Functions](function/index.md)
* [genaiz function build](function/build.md)
* [genaiz function create](function/create.md)
* [genaiz function data](function/data.md)
* [genaiz function init](function/init.md)
* [genaiz function list](function/list.md)
* [genaiz function prop](function/prop.md)
* [genaiz function publish](function/publish.md)
* [genaiz function run](function/run.md)
* [genaiz function start](function/start.md)
* [genaiz function stop](function/stop.md)
* [genaiz function test](function/test.md)

## [Solutions](solution/index.md)
* [genaiz solution create](solution/create.md)
* [genaiz solution publish](solution/publish.md)

## [Workflows](workflow/index.md)
* [genaiz workflow create](workflow/create.md)
* [genaiz workflow delete](workflow/delete.md)
* [genaiz workflow links](workflow/links.md)
* [genaiz workflow nodes](workflow/nodes.md)
* [genaiz workflow prop](workflow/prop.md)

## Global Validation

### Description

* A valid description can not be longer than 4096 characters.
* Any characters may be accepted.

> [!CAUTION]
> Descriptions currently do not support any kind of official templating engine. If changes do occur, then the validity will depend on the Templating engine selected.

### FQDNV

For Fully Qualified Domain Name and Version.

* Is simply a composition of [OEM](#handle-and-oem)/[HANDLE](#handle-and-oem):[VERSION](#version)
* Each component needs to be valid or needs to be accounted for the command requiring the field. (Through options if necessary)

### Handle and OEM

* Can only have letters (upper or lower case), digits, dots, dashes and underscores.
* Can only start with a letter (upper or lower case) or a digit.
* Can not have 2 consecutive characters of type dots, dashes and underscores.
* Can not end with a character of type dot, dash or underscore.

### Name

* A name can have any characters.
* The length must not exceed 255 characters.

### Version

* Must be a valid [Semantic Version](https://semver.org/) string.
* Can not contain pre-release version identifiers as those are reserved by the broker.

## Property Specification Validation

### Property Key

* The key of a property specification mirrors the string used to define conventional Environment Variables.
* It must be composed only of capitalized alphanumeric characters and underscores; `[A-Z_][A-Z0-9_]*`
* Keys can not be expanded into other keys. That is you can not define a key using the value of another one. For example, MY_KEY_$KEY_INDEX is not a valid key.

### Property Name

* A name can contain any kind of characters for as long as it does not extend to more than 255 characters.

### Property Description

* A description can extend to up to 4096 characters.

### Property Types

* Only "STRING", "INT", "BOOL", "DOUBLE" and "ENUM" are allowed.
* Lower case strings will be capitalized on publishing and writing.

### Property Enum Values

* Enum values must have between 1 and 512 characters
