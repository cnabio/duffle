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

Subsequently, and satisfactory solution must be able to accomodate a wide variety of configurational permutations, ideally without dictating that credentialing tools change in any way.

## Credential Sets

A *credential set* is a named set of credentials (or credential generators) that can is managed on the local host (via duffle) and injected into the invocation container on demand.

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
credentials:
    - name: kubeconfig
      source:
        # Where is it on the host system?
        type: file
        path: $HOME/.kube/kubeconfig
      destination:
        # Where does it go in the container?
        type: file
        path: /root/.kube/kubeconfig
   - name: service-token
     source:
       # Run a command on localhost and capture the output
       type: command
       command: "gen-token example.com"
     destination:
       # Expose the output to the container as an environment variable
       type: env
       name: SERVICE_TOKEN
```

The above declares two different credentails. One is a kubeconfig file that will be copied from the localhost into the container verbatim. The second is a security token that will be generated on the localhost, and passed into the container as an environment variable (`$SERVICE_TOKEN`).

Credential sets are specified when needed:

```console
$ duffle run --credentials=staging example/myapp:1.0.0
> loading credentials from $HOME/.duffle/credentails/staging.yaml
> running example/myapp:1.0.0
```

## Default Credentail Sets

A default credential set may be specified in the Duffle preferences. They may be set in a project's `duffle.yaml`, or (more frequently) per user in the `$HOME/.duffle/preferences.yaml` file:

```yaml
defaultCredentials: "staging"
```

## Limitations

In this model, credentials can only be injected as files and environment variables. Some systems may not be satisfied with this limitation, in which case additional scripting may be required inside of the invocation image.

Other:

- We might be able to put all credentials in one large YAML file. Credentails may include x509 certs or other large things
- There is no way to specify in a CNAB bunfle what credentials are required on the host system, other than by documentation. We might have to figure that out at some point
- We don't address how a credential might be injected into a file on the image. The assumption is that such a thing would be scripted
