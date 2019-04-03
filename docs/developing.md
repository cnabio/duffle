# Developer's Guide

## Getting the code

Cloning this repository and change directory to it:
```console
$ go get -d github.com/deislabs/duffle/...
$ cd $(go env GOPATH)/src/github.com/deislabs/duffle
```

### Prerequisites
You need:
* A working Go 1.11.4 (or later) environment
* make

Before you start working with the code, issue:
```console
$ make bootstrap
```

## Building

To build `bin/duffle`, issue:
```console
$ make build
```

If you want to install `duffle`, issue:
```console
$ sudo make install
```

## Testing

To run the tests, issue:
```console
$ make test
```

## Linting

To lint the code, issue:
```console
$ make lint
```

If this detects that some imports need re-organising (errors like "File is not `goimports`-ed"), issue:
```console
$ make goimports
```

## Debugging

For instructions on using VS Code to debug the Duffle binary, see [the debugging document](debugging.md).
