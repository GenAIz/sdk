# GenAIz SDK

## Usage

### Single Node Solution Examples

#### Create

The following example creates a solution **solution-1** and a smart function **my-bash-example** using the recipe **bash-example**. It then builds the function with the default version assigned by create, [0.1.0](doc/function/index.md#version), assigns a single node to the **default** workflow, authenticates a user with [broker.genaiz.com](doc/account/index.md#host) and publishes the solution to the broker.

```bash
genaiz sn create mySolutionDir --name="My Solution" --handle="solution-1" \
  --oem="com.genaiz.dev" --description="A Description"
cd mySolutionDir
genaiz sf create mySolutionDir/mySmartFunction --recipe="bash-example" \
  --oem="com.genaiz.dev" --handle="my-bash-example" --name="My Bash Example"
cd mySmartFunction
genaiz sf build
cd ..
genaiz wf nodes add default node1 --name="Single Node" \
  --description="My Single Node" --sf="com.genaiz.dev/my-bash-example:0.1.0"
genaiz ac login broker.genaiz.com --username="myUsername"
genaiz sn publish broker.genaiz.com
```

## Design

Design was modeled with a behavior driven approach focusing on user's usage studies. The design rules are presented under the [doc](doc/index.md) folder with associated Gherkin files accessible to the documents. The following target the [Genaiz Toolkit](#genaiz-smartfunction-toolkit):

* [Account Scenarios](doc/account/index.md)
* [Function Scenarios](doc/function/index.md)

## Modules

### [GenAIz SmartFunction Toolkit](genaiz/README.md)

Handles building, debugging, running, testing and publishing Smart Functions to a GenAIz Broker Platform.

### [GenAIz Integration Test Toolkit](genaiz-it/README.md)

Handles testing of service deployments with the GenAIz SmartFunction Kit. It offers the following functionality:

- CNCF Distribution Registry bootstrapping for tests
- Wiremock deployment to simulate what the SDK expects out of an orchestrator deployment
- Cucumber-like Genaiz runtime implementation with the step definitions used under Gherkin features 

### [GenAIz Library](genaiz-lib/README.md)

The library of common functionality used by other GenAIz go modules. Providing:

- Common language enhancements
- Common support code for unit tests as GO unit tests can not share code across compilation units

### [GenAIz OAuth Utility Toolkit](genaiz-oauth/README.md)

The oauth utility kit provides functionality to:

- Generate Certificate Authority ECDSA key pair files
- Generate Signed Certificate ECDSA key pair files
- Manage JWKS public key stores
- Create and Decode JWT Signed tokens

These facilities should be used to integrate token authentication on  [CNCF Distribution Registry](https://distribution.github.io/distribution/spec/auth/jwt/). With source code available on [GitHub](https://github.com/distribution/distribution)

## Development Guide

### Prerequisites

* [Golang](https://go.dev/doc/install)
* [Gnu Make](https://www.gnu.org/software/make/manual/make.html)
* [Docker](https://docs.docker.com/engine/install/)

### Building All Modules

Individual modules with make files will answer their own targets, but most will have *clean, build* and *test* at a minimum.

> [!TIP]
> make install on the root Makefile should install all of genaiz, genaiz-it & genaiz-oauth under $HOME/go/bin

#### [GenAIz Makefile](genaiz/README.md#makefile)

#### [GenAIz-IT Makefile](genaiz-it/README.md#makefile)

#### [GenAIz-Lib Makefile](genaiz-lib/README.md#makefile)

#### [GenAIz-OAuth Makefile](genaiz-oauth/README.md#makefile)

### Testing GenAIz

TODO