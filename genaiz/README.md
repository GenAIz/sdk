# Genaiz SmartFunction Toolkit

## Makefile

Building the project with its associated make file can install the application, its manual pages and associated resources. To install locally:

```shell
cd genaiz && make all
```

To build a docker image of the sdk

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
cd genaiz
go build
./genaiz --help
```

## Commands

### function (sf)

The function module is used to manage smart functions and publish them as Docker images to an Orchestrating Broker.

```shell
genaiz sf --help
```

#### build

Simply builds the Smart Function image. Build should always be called as part of other commands, except stop, if no image with the Function definition can be found.

```shell
genaiz sf build --help
```

#### create

The command creates a new Smart Function folder with a typical layout. By default, the command will ask the user interactively to confirm all initial values assigned to the function. The layout created should have a genaiz.yaml file populated with the values passed to this command.

```shell
genaiz sf create --help
```

#### debug

This should start the image with a disposable container in interactive mode.

```shell
genaiz sf debug --help
```

> [!CAUTION]
> The debug command is planned, not yet implemented

#### init

The command initiates a new Smart Function under an existing folder. By default, the command will ask the user interactively to confirm what it found under the folder, creating the genaiz.yaml file for the Smart Function.

```shell
genaiz sf init --help
```

#### list

The command lists all images with their versions belonging to Smart Function folder. In addition, it will list any local containers configured with any of the listed images.

```shell
genaiz sf list --help
```

#### publish

The command initiates a session with the GenAIz broker retrieving authorization tokens to publish a Smart Function image onto the GenAIz marketplace. This would require the user to be logged in using a **genaiz ac login** preamble to retrieve licensing agreements.

```shell
genaiz sf publish --help
```

#### run

This should start the image with a disposable container in detached mode.

```shell
genaiz sf run --help
```

#### start

This should start the image with a named container, potentially replacing any existing one, and potentially disposing of it after completion.

```shell
genaiz sf start --help
```

#### stop

This should stop a named container, potentially disposing of it after it exits.

```shell
genaiz sf stop --help
```

#### test

Similar to run, but starting a disposable container attached to the current shell.

```shell
genaiz sf test --help
```

### account (ac)

The account module is used to manage account credentials and configuration policies with an Orchestrating Broker.

#### login

The login command obtains an identity token from the specified Orchestrating Broker and registers the current active account for a specified amount of time by the broker.

```shell
genaiz ac login www.genaiz.com
```

#### logout

The logout command is invoked to explicitly remove a known session id from the local sdk configuration. The file is found under $HOME/.cache/genaiz/.auth. The command will log out the active session by default is no --host parameter is specified.

```shell
genaiz ac logout
```