# Developer's Guide

## Duffle Development

### Build

```shell
make build
```

### Install

This target will simply add the binary built above to your path.

```shell
make install
```

### Test

```shell
make lint
make test
```

### Docker Build

```shell
make build-docker-bin docker-build
```

### Docker Push

```shell
make docker-push
```

## Functional Tests

### Local (uses duffle binary in path)

#### Single bundle

Supply the `<name>:<version>` of a remote bundle to run the `duffle` binary against.
This command defaults to running locally using the `duffle` binary in your path.

```shell
BUNDLE=fun:0.2.0 make test-functional
```

#### All remote bundles

```shell
make test-functional
```

### Docker-based (runs in duffle image)

#### Single bundle

Supply the `<name>:<version>` of a remote bundle to run the `duffle` binary against.
This command defaults to running inside of the docker image built above.

```shell
BUNDLE=fun:0.2.0 make test-functional-docker
```

#### All remote bundles

```shell
make test-functional-docker
```