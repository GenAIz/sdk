# Genaiz Smart Function Toolkit

## Minimal Build

```shell
cd genaiz
go build
./genaiz sf --help
```

## Commands

### Build

Simply builds the Smart Function image. Build should always be called as part of other commands, except stop, if no image with the Function definition can be found.

```shell
cd genaiz
go build
./genaiz sf build --help
```

### Create

The command creates a new Smart Function folder with a typical layout. By default, the command will ask the user interactively to confirm all initial values assigned to the function. The layout create should have a .genaiz.yaml file populated with the values passed to this command.

```shell
cd genaiz
go build
./genaiz sf create --help
```

> [!CAUTION]
> The Create command is planned, not yet implemented

### Debug

This should start the image with a disposable container in interactive mode.

```shell
cd genaiz
go build
./genaiz sf debug --help
```

### Init

The command initiates a new Smart Function under an existing folder. By default, the command will ask the user interactively to confirm what it found under the folder, creating the .genaiz.yaml file for the Smart Function.

```shell
cd genaiz
go build
./genaiz sf init --help
```

> [!CAUTION]
> The Init command is planned, not yet implemented

### Publish

The command initiates a session with the Genaiz broker retrieving authorization tokens to publish a Smart Function image onto the Genaiz marketplace. This would require the user to be logged in using a **genaiz ac login** preamble to retrieve licensing agreements.

```shell
cd genaiz
go build
./genaiz sf publish --help
```

> [!CAUTION]
> The Publish command is planned, not yet implemented

### Run

This should start the image with a disposable container in detached mode.

```shell
cd genaiz
go build
./genaiz sf run --help
```

### Start

This should start the image with a named container, potentially replacing any existing one, and potentially disposing of it after completion.

```shell
cd genaiz
go build
./genaiz sf start --help
```

### Stop

This should stop a named container, potentially disposing of it after it exits.

```shell
cd genaiz
go build
./genaiz sf stop --help
```

### Test

Similar to run, but starting a disposable container attached to the current shell.

```shell
cd genaiz
go build
./genaiz sf test --help
```

## Integration Testing

### Locally Authenticated Docker Registry

If the need for testing docker login commands with the sdk arises, we can deploy a local registry featuring a basic username/password login preamble:

```shell
mkdir -p .registry && cd .registry
mkdir -p auth
docker run --entrypoint htpasswd httpd:2 -Bbn genaiz_user genaiz_pass > auth/htpasswd
docker run -d -p 5000:5000 --restart=always --name registry_genaiz  \
  -v `pwd`/auth:/auth  \
  -e "REGISTRY_AUTH=htpasswd"  \
  -e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm"  \
  -e "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd" registry:latest
docker login -u genaiz_user -p genaiz_pass
cd -
```