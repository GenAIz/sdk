# GenAIz SDK
<sub>Genaiz Version 0.3.3</sub>


The GenAIz SDK is a tool for creating, building and publishing Smart Functions to the GenAiz Orchestrator platform.

* [Installation](#installation)
* [Running the GenAIz CLI](#running-the-genaiz-cli)
  * [Quickstart Examples](#quickstart-examples)
  * [Result Values](#result-values)
* [Development Guide](#development-guide)
  * [Prerequisites](#prerequisites)
  * [Building from source](#building-from-source)
  * [Modules](#modules)
  * [Unit testing](#unit-testing)
* [Acceptance Testing](#acceptance-testing)
* [Troubleshooting](#troubleshooting)
  * [GoLang](#golang)
  * [Docker](#docker)
* [Contributing](#contributing)
* [License](#license)

## Installation

> [!IMPORTANT]
> TODO: We need a GitHub package which can be installed with curl (#32)

## Running the GenAIz CLI

Working with the GenAIz CLI is a simple series of commands. You may run `genaiz --help` at any time to get a list of available commands and their options.

### Quickstart Examples

#### Creating a simple solution

The following example creates a solution **solution-1** and a smart function **my-bash-example** using the recipe **bash-example**. It then builds the function with the default version assigned by create, [1.0.0](acceptance-tests/docs/function/index.md#version), assigns a single node to the **default** workflow, authenticates a user with [broker.genaiz.com](acceptance-tests/docs/account/index.md#host) and publishes the solution to the broker.

```bash
genaiz sn create mySolutionDir --name="My Solution" --handle="solution-1" \
  --oem="com.genaiz" --description="A Description"
cd mySolutionDir
genaiz sf create mySmartFunction --recipe="bash-example" \
  --oem="com.genaiz" --handle="my-bash-example" --name="My Bash Example"
cd mySmartFunction
genaiz sf build
cd ..
genaiz wf nodes add default node1 --name="Single Node" \
  --description="My Single Node" --sf="com.genaiz/my-bash-example:1.0.0"
genaiz ac login broker.genaiz.com --username="myUsername"
genaiz sn publish
```

#### Running a simple function

The following example creates a smart function **function-1**, without a parent solution. It then proceeds to build it out of context. We can then list and run the smart function on a local **Docker** installation within the folder created or out of it.

```bash
genaiz sf create function-1 --oem="com.genaiz" \
  --recipe="bash-example"
genaiz sf build --context=function-1/
cd function-1
genaiz sf list
genaiz sf run
```

> [!NOTE]
> Run disposes of any container created for the operation

#### Testing a simple function with Properties

The following example creates a smart function **function-2**, without a parent solution. We then proceed to add property specifications to the function. The example completes with building the function in context, and testing it with the console attached using `test`.

```bash
genaiz sf create function-2 --oem="com.genaiz" \
  --recipe="bash-example"
cd function-2
genaiz sf prop add MY_KEY --name="Key Example" \
  --default-value=10 --type=int
genaiz sf prop env MY_KEY 12
genaiz sf build
genaiz sf test
```

#### Testing a simple function with Data Link Properties

The following example creates a smart function **function-3**, without a parent solution. We then proceed to create and publish a data link **datalink-1** and add it to the smart function for provisioning on a broker.

```bash
genaiz sf create function-3 --oem="com.genaiz" --version="0.2.0" \
  --recipe="bash-example"
cd function-3
genaiz dk create datalink-1 --oem="com.genaiz" --version="1.0.0" \
  --description="My DataLink"
cd function-3
genaiz data source add com.genaiz/datalink-1:1.0.0
genaiz sf build
genaiz sf publish
```

### Result Values

Support for result values is provided on `genaiz sf publish` only. From manually edited settings, for example:

```yaml
function:
    build:
        repository: com.genaiz/result-values
        version: latest
    publish:
        handle: result-values
        name: result-values
        oem: com.genaiz
        resultvalues:
            - excel
            - valid-set
        type: function
        version: 1.0.0
    run:
        input: run/in
        log: run/{timestamp}/log
        output: run/{timestamp}/out
        var: run/{timestamp}/var
```

## Development Guide

### Prerequisites

To be able to build the project you will need to install the following tools for your platform. You can find more information at the links provided here:

* [Golang](https://go.dev/doc/install) - All tools and the SDK are written in Go
* [GNU Make](https://www.gnu.org/software/make/manual/make.html) - Not strictly required, but all build scripts are driven through Makefiles
* [Docker](https://docs.docker.com/engine/install/) - The local execution and build runtime requires a local Docker installation

### Building from Source

```bash
cd sdk
make genaiz/install
```

> [!TIP]
> It is possible $HOME/go/bin is not included within a users' path. That can be fixed with `export PATH=$PATH:$HOME/go/bin` added to the users' .bashrc, .bash_profile or .profile file, depending on the platform.

Building the individual modules can be achieved in the same manner or simply by running the install target on the root project: 

```bash
make install
```

Print a summary of all make targets by entering

```bash
make help
```

### Modules

Individual modules with make files will answer their own targets, but most will have *clean, build* and *test* at a minimum.

#### [GenAIz SmartFunction Toolkit](genaiz/README.md)

Handles building, running, testing and publishing Smart Functions to a GenAIz Broker Platform.

#### [GenAIz Integration Test Toolkit](genaiz-it/README.md)

Handles testing of service deployments with the GenAIz SmartFunction Kit. It offers the following functionality:

- CNCF Distribution Registry bootstrapping for tests
- Wiremock deployment to simulate what the SDK expects out of an orchestrator deployment
- Cucumber-like Genaiz runtime implementation with the step definitions used under Gherkin features

#### [GenAIz Library](genaiz-lib/README.md)

The library of common functionality used by other GenAIz go modules. Providing:

- Common language enhancements
- Common support code for unit tests as GO unit tests can not share code across compilation units

#### [GenAIz OAuth Utility Toolkit](genaiz-oauth/README.md)

The oauth utility kit provides functionality to:

- Generate Certificate Authority ECDSA key pair files
- Generate Signed Certificate ECDSA key pair files
- Manage JWKS public key stores
- Create and Decode JWT Signed tokens

These facilities should be used to integrate token authentication on  [CNCF Distribution Registry](https://distribution.github.io/distribution/spec/auth/jwt/). With source code available on [GitHub](https://github.com/distribution/distribution)

### Unit testing

Genaiz unit testing facilities only cover the production module of [genaiz](#genaiz-smartfunction-toolkit). The unit test files cover the entire functionality with ideally, individual units tested in isolation to provide a `White Box` map of the implementation.

The goal here is to provide a minimal harness for catching regression and also provide future functionality for associating units to acceptance test cases.

Currently, a coverage report can be obtained with:

```bash
cd genaiz && make coverage
```
This opens a visual HTML report of the implementation tested by the unit tests.

## Acceptance Testing

Not to be confused with unit testing: Acceptance testing is to unit testing what acceleration is to speed. An acceptance test case is composed of several units working to achieve feature requirements. Acceptance typically provides a `Black Box` map of a feature set.

[Acceptance testing is made of a repository](acceptance-tests/docs/index.md) of [Gherkin](https://cucumber.io/docs/gherkin/) features describing what the [genaiz](#genaiz-smartfunction-toolkit) needs to provide to a CLI user, but also to other types of integrations relying on CLI commands.

>[!NOTE]
>Currently, the features can only be read and ran manually. The runtime to automatically execute them would be hosted under the [genaiz-it](#genaiz-integration-test-toolkit) module.

## Troubleshooting

### Golang

#### Go tools tag

We use 2 different environment builds. The `prod` build is the build configured by default under the `Makefile`s of the project. It builds with an HTTP request gate which denies connections on `http` addresses. When this is not practical, for testing reasons, the `dev` tag can be enabled to exempt `localhost` from this restriction.

When using an IDE to run go tests, you will have to configure it manually with `-tags dev`. Under **Intellij**, this is configured under

* `Run/Debug Configurations`
    * `Edit configuration templates`
        * `Go Test`
            * `Go tool arguments`

#### Tests are failing with `permission denied`

Golang's testing tools rely on executing code from the `/tmp` folder. On many modern system /tmp is now mounted to [tmpfs](https://www.kernel.org/doc/html/latest/filesystems/tmpfs.html) with `noexec`. It is a recommendation from [CIS](https://www.cisecurity.org/). You may have to fix GO's toolchain environment adding:

```bash
export GOTMPDIR="/var/tmp"
```

to your `$HOME/.bashrc` or `$HOME/.bash_profile` file. Note that `/var/tmp` is just another one of those "locations" that is usually the target for temporary files while compiling projects. You could very well just use `$HOME/go/tmp`, depending on how your partition layout was created, if `/var/tmp` is also "secured".

#### Tests are failing will nil panic traces

We use a flavour of **Monk Patching** to capture calls to various core methods like os.Exit and fmt.Printf. These require the source to be compiled without function inlining enabled:

`-gcflags=-l`

The option is set in the `Makefile` requiring it for the `test` and `report` targets, but if you use an IDE Run configuration you will have to configure it manually. Under **Intellij**, this is configured under 

* `Run/Debug Configurations`
  * `Edit configuration templates`
    * `Go Test`
      * `Go tool arguments`

### Docker

#### Connection to /var/run/docker.sock failing

Check that docker is running:

```bash
systemctl status docker
```

Check that the permissions on /var/run/docker.sock allows your user to connect to it:

```bash
ls -l /var/run/docker.sock
groups
```

Your user should be in the same group owning the docker.sock file. 

If all conditions are met, and you are not using Windows, write to `sdk 'at' genaiz.com` with the following details:
* Docker version
* OS and version
* If you can build with docker build
* the output of the command with --log-level=debug

## Contributing

To contribute to this project, please follow these steps:

1. Fork this repository to your personal account.
2. Create a new branch for your feature or bug fix.
3. Commit and push changes to your forked repository.
4. Open a Pull Request (PR) to allow merging into the upstream repository.

## License

This project is licensed under a commercial license. Unauthorized duplication is prohibited.

&copy; 2018 - 2026 GenAIz. All rights reserved.