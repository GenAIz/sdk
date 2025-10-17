# SDK Usage Design Doc

The concepts and design related to the Orchestrator and GenAIz platform should be available through their own respective projects. This document serves the purpose of exposing how the main entry point of the SDK, its CLI, is behaving.

## [Accounts](account/index.md)
* [genaiz account login](account/index.md#login)
* [genaiz account logout](account/index.md#logout)

## [Smart Functions](function/index.md)
* [genaiz function build](function/build.md)
* [genaiz function create](function/create.md)
* [genaiz function debug](function/debug.md)
* [genaiz function init](function/init.md)
* [genaiz function list](function/list.md)
* [genaiz function publish](function/publish.md)
* [genaiz function run](function/run.md)
* [genaiz function start](function/start.md)
* [genaiz function stop](function/stop.md)
* [genaiz function test](function/test.md)

## [Solutions](solution/index.md)
* [genaiz solution create](solution/create.md)
* [genaiz solution publish](solution/publish.md)

## Global Validation

### Description

* A valid description can not be longer than 4096 characters.
* Any characters may be accepted.

> [!CAUTION]
> Descriptions currently do not support any kind of official templating engine. If changes do occur, then the validity will depend on the Templating engine selected.

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