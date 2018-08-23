# Duffle Credential Sets

This document covers how credentials are passed into Duffle from the environment.

> This functionality was not part of the initial draft specification.

## The Credential Problem

Consider the case where a CNAB bundle named `example/myapp:1.0.0` connects to both ARM (Azure IaaS API service) and Kubernetes. Each has its own API surface which is secured by a separate set of credentials. ARM requires a periodically expiring token managed via `az`. Kubernetes stores credentialing information in a `KUBECONFIG` YAML file.

When the user runs `duffle run example/myapp:1.0.0`, the operations in the invocation image need to be executed with a specific set of credentials from the user.

Those values must be injected from `duffle` into the invocation image.

But Duffle must know which credentials to send.

The situation is complicated by the following factors:

- There is no predefined set of services (and thus predefined set of credentialing) specified by CNAB.
- It is possible, and even likely, that a user may have more than one set of credentials for a service or backend (e.g. credentials for two different Kubernetes clusters)
- Some credentials require commands to be executed on a host, such as unlocking a vault or regenerating a token
- There is no standard file format for storing credentials
- The consuming applications may require the credentials be submitted via different methods, including as environment variables, files, or STDIN.

Subsequently, any satisfactory solution must be able to accommodate a wide variety of configurational permutations, ideally without dictating that credentialing tools change in any way.

## Credential Sets

A *credential set* is a named set of credentials (or credential generators) that is managed on the local host (via duffle) and injected into the invocation container on demand.

### On-disk format

The `$HOME/.duffle/` directory is where user-specific Duffle configuration files are placed. Inside of this directory, Duffle stores credential information in its own subdirectory:

```
$HOME/.duffle/
  |- credentials/
          |- production.yaml
          |- staging.yaml
```

NOTE: YAML is not a required format, but it's easy to write as a real human. So... I'll start there

A credential YAML file contains a set of named credentials that are resolved locally (if necessary) and then pushed into the container.

Example (`staging.yaml`):

```yaml
name: staging     # Must match the name portion of the file name (staging.yaml)
credentials:
  - name: read_file
    source:
      path: $SOMEPATH/testdata/someconfig.txt  # credential will be read from this file
                                               # In 'path', env vars are evaluated.
    destination:
      # credential data will be presented as environment variable $TEST_READ_FILE
      env: TEST_READ_FILE    
  - name: run_program
    source:
      command: "echo wildebeest" # The command `echo wildebeest` will be executed
                                 # An error will cause the process to exit
    destination:
      env: TEST_RUN_PROGRAM  # Results will be placed as an env var.
  - name: use_var
    source:
      env: TEST_USE_VAR      # This will read an env var from local, and copy to dest
      value: "this space intentionally left non-blank"
    destination:
      env: TEST_USE_VAR
  - name: fallthrough
    source:
      name: NO_SUCH_VAR      # Assuming this is not set....
      value: quokka          # Then this will be used as the default value
    destination:
      env: TEST_FALLTHROUGH     # The result will be written to env var...
      path: animals/quokka.txt  # and also to a file path.
  - name: plain_value
    source:
      value: cassowary       # Load this literal value.
    destination:
      path: animals/cassowary.txt  # Save the value to a file on dest.
```

The above shows several examples of how credentials can be loaded from a local source and
sent to an in-image destination.

Loading from source is done from four potential inputs:

- `value` is a literal value
- `env` is loaded from an environment variable (and can fall back to `value` as a default)
- `path` is loaded from a file at the given path (or else it errors)
- `command` executes a command, and returns the output as the value (or else it errors)

Data can then be passed into the image in one of two ways:

- `env` will store the data as an environment variable
- `path` will store the data as the contents of a file located at the given path

Note that both `env` and `path` can be specified, which will result in the data being stored in both.

Credential sets are specified when needed:

```console
$ duffle run --credentials=staging example/myapp:1.0.0
> loading credentials from $HOME/.duffle/credentials/staging.yaml
> running example/myapp:1.0.0
```

Credential sets are loaded locally. All commands are executed locally. Then the results are injected into the image at startup.

## Default Credential Sets

A default credential set may be specified in the Duffle preferences. They may be set in a project's `duffle.yaml`, or (more frequently) per user in the `$HOME/.duffle/preferences.yaml` file:

```yaml
defaultCredentials: "staging"
```

## Limitations

In this model, credentials can only be injected as files and environment variables. Some systems may not be satisfied with this limitation, in which case additional scripting may be required inside of the invocation image.

Other:

- We might be able to put all credentials in one large YAML file. Credentials may include x509 certs or other large things
- There is no way to specify in a CNAB bundle what credentials are required on the host system, other than by documentation. We might have to figure that out at some point
- We don't address how a credential might be injected into a file on the image. The assumption is that such a thing would be scripted
