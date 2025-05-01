# Genaiz SDK Toolkits

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
### Debug

This should start the image with a disposable container in interactive mode.

```shell
cd genaiz
go build
./genaiz sf debug --help
```

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