# Developer's Guide

## Getting the code

Cloning this repository and change directory to it:
```console
$ go get -d github.com/deislabs/duffle/...
$ cd $(go env GOPATH)/src/github.com/deislabs/duffle
```

### Prerequisites

You need:

* make
* Docker

## Containerized Development Environment

To ensure a consistent development environment for all contributors, Duffle
relies heavily on Docker containers as sandboxes for all development activities
including dependency resolution and executing tests.

`make` targets seamlessly handle the container orchestration.

If, for whatever reason, you must opt-out of executing development tasks within
containers, set the `SKIP_DOCKER` environment variable to `true`, but be aware
that by doing so, the success or failure of development-related tasks, tests,
etc. will be dependent on the state of your system, with no guarantee of the
same results in CI.

## Developing on Windows

All development-related tasks should "just work" on Linux and Mac OS systems.
When developing on Windows, the maintainers strongly recommend utilizing the
Windows Subsystem for Linux.

[This blog post](https://nickjanetakis.com/blog/setting-up-docker-for-windows-and-wsl-to-work-flawlessly)
provides excellent guidance on making the Windows Subsystem for Linux work
seamlessly with Docker Desktop (Docker for Windows).

## Building

To build everything (binaries for Linux, Mac, and Windows on amd64 architecture)
as well as a linux/amd64 Docker image containing the corresponding binary:

```console
$ make build
```

To build binaries for Linux, Mac, and Windows, but no Docker image:

```console
$ make build-all-bins
```

To build for one specific OS / architecture:

```console
$ OS=<desired OS> ARCH=<desired architecture> make build-bin
```

To build only the Docker image and nothing else:

```console
$ make build-image
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

## Dependency Resolution

If, at any time, you need to (re-)resolve the project's dependencies, perhaps
because a new one is needed or an existing one is no longer needed, issue:

```console
$ make dep
```

## Debugging

For instructions on using VS Code to debug the Duffle binary, see [the debugging document](debugging.md).
