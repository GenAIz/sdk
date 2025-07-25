# GenAIz Integration Test Toolkit

## Makefile

Building the test toolkit with its associated make file can install the application. To install locally:

```shell
cd genaiz && make all
```

To build a docker image of the test toolkit

```shell
cd genaiz && make docker
```

To execute tests alone or to get coverage metrics

```shell
cd genaiz && make coverage
```

For more information about all provided targets

```shell
cd genaiz && make help
```

## Minimal Build

```shell
cd genaiz-it
go build
./genaiz-it --help
```

## Commands

### feature

The feature command is used to start, stop and report on test bundles. Tests are grouped using Gherkin feature files. Each file can sport several Scenarios and each of them can be executed individually or in group. 

Note that Scenarios from Genaiz IT can trigger dependent scenarios listed only under the same feature file. These often cascade in a series of SDK commands which are supposed to put the feature in a certain state.

When the **report** command runs, the state of a feature stops at the very last scenario it executed providing there were interactions with a Wiremock request.

```shell
genaiz-it feature --help
```
### registry

A standalone way to bootstrap the embedded genaiz_registry container, which deploys a registry with the necessary OAUTH gear to verify token signatures and their issuers according to the [Specs](https://distribution.github.io/distribution/spec/auth/jwt/#verifying-the-token)

```shell
genaiz-it registry --help
```

Note that tokens for the registry can be produced using the [genaiz-oauth](../genaiz-oauth/README.md#tokens) utility

### wiremock

A standalone way to bootstrap the embedded genaiz_wiremock container, which deploys **orchestrator** request mappings for specified features and scenarios. The utility should provide a way to modify mappings according to scenario requirements as well (TODO)

```shell
genaiz-it wiremock --help
```
