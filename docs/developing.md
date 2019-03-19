# Developer's Guide

## Getting the code

Cloning this repository and change directory to it:
```console
$ go get -d github.com/deislabs/duffle/...
$ cd $(go env GOPATH)/src/github.com/deislabs/duffle
```

## Building

### Prerequisites
You need:
* A working Go 1.11.4 (or later) environment
* make

Before you start, issue:
```console
$ make bootstrap
```

Then, to build `duffle`, issue:
```console
$ make build
```
The resultant binary is placed in `bin/duffle`.

If you want to install `duffle`, issue:
```console
$ sudo make install
```

## Testing

To run the tests, issue:
```console
$ make test
```

## Debugging

For instructions on using VS Code to debug the Duffle binary, see [the debugging document](debugging.md).
