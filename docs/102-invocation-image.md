# The Invocation Image

The `invocationImage` section of a `bundle.json` must point to exactly one image (the invocation image). This image must be formatted according to the specification laid out in the present document.

When a bundle is executed, the invocation image will be retrieved (if necessary) and started. Credential and parameter data is passed to it, and then its `run` tool is executed. (See [The Bundle Runtime](103-bundle-runtime.md) for details).

This section describes the layout of an invocation image.

## Components of an Invocation Image

An invocation image is composed of the following:

- A file system hierarchy following a defined pattern (below)
- A main entry point, which is an executable (often a script) called the _run tool_
- Runtime metadata (Helm charts, Terraform templates, etc)
- The bundle file, which enumerates images used and available parameters, along with other metadata.
- The material necessary for reproducing the invocation image (`Dockerfile` and `packer.json` are two examples)

### The File System Layout

The following exhibits the filesystem layout:

```yaml
cnab/
├── bundle.json​      # Required -- but also stored outside of the image
└── Dockerfile​         # Optional
└── app​                # Required
    ├── run​            # Required: This is the main entrypoint, and must be executable
    ├── charts​         # Example: Helm charts might go here
    │   └── azure-voting-app​
    │       ├── Chart.yaml​
    │       ├── templates​​
    │       │   └── ...
    │       └── values.yaml​
    └── sfmesh​         # Example: Service Fabric definitions might go here
        └── sfmesh-deploy.json
```

The `app/` directory contains subdirectories, each of which stores configuration for a particular target environment. The `app/run` file _must be an executable file_ that will act as the "main" installer for this CNAB bundle.

The contents beneath `/cnab/app/SUBDIRECTORY` are undefined by the spec. `run` is considered the only reserved word underneath `/cnab/app/`

_NOTE:_ While _ad hoc_ data may be placed under `/cnab/app/`, all possible names under the root container directory (`/cnab/`) are considered reserved, and no additional names may be added beyond the spec. For example, creating `/cnab/bin/` would violate the specification.

### Why a `cnab/` directory?

Earlier versions of the spec did not include a top-level `cnab` directory, and dropped all files in `/`. However, this would put packages into an unregulated namespace (the root directory). Existing images may already contain or require files and directories like `run` or `app`. In turn, this leads to the potential for unintended namespace collisions.

By declaring a `cnab` directory, we reduce the likelihood of collisions, while also making it possible for base images to scaffold out CNAB constructs that can be leveraged.

For example, a base CNAB image may provide something like this:

```Dockerfile
FROM ubuntu:latest

COPY ./some-chart /cnab/app/charts/some-chart
```

Then a later `Dockerfile` could do this:

```Dockerfile
FROM above-image:latest

RUN helm inspect values /cnab/app/charts/some-chart > ./myvals.yaml &&  sed ...
```

The example above is simply intended to show how by reserving the `/cnab` directory, we can make images composable, while not worrying about non-CNAB images putting data in places CNAB treats as special.

## The bundle.json File

The `bundle.json` file included inside of the CNAB image _must_ be identical to the version stored outside of the bundle. The `bundle.json` is required to make the image format portable.

This format is defined in the previous [bundle.json definition](101-bundle-json.md).

_Note:_ The `bundle.json` file that exists inside of the bundle is not a signed bundle, because signing requires calculating the hash of the `invocationImage`.

## Image Construction Files

Including a Dockerfile is _recommended_ for all images built with Docker. It is useful for reproducing a bundle. For other build tools, the build tool's definition may be included instead (e.g. `packer.json` for VM images built with Packer).

The remainder of this subsection is non-normative.

The `Dockerfile` used to build the invocation image must be stored inside of the invocation image. This is to ensure reproducibility, and in order to allow rename operations that require a rebuild.

This is a normal Dockerfile, and may derive from any base image.

Example:

```Dockerfile
FROM ubuntu:latest

COPY ./Dockerfile /cnab/Dockerfile
COPY ./bundle.json /cnab/manfiest.json
COPY ./parameters.json /cnab/parameters.json

RUN curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
RUN helm init --client-only
RUN helm repo add example-stable https://repo.example.com/stable

CMD /cnab/app/run
```

The above example installs and configures Helm inside of a base Ubuntu image. Note that there are no restrictions on what tools may be installed.

## The Run Tool

The run tool _must_ be located at the path `/cnab/app/run`. It _must_ be executable. It _must_ react to the `CNAB_ACTION` provided to it.

The specification does not define what language(s) the tool must be written in, or any details about how it processes the information. However, the following is a non-normative example:

```bash
#!/bin/sh

action=$CNAB_ACTION
name=$CNAB_INSTALLATION_NAME 

case $action in
    install)
    echo "Install action"
    ;;
    uninstall)
    echo "uninstall action"
    ;;
    upgrade)
    echo "Upgrade action"
    ;;
    status)
    echo "Status action"
    ;;
    *)
    echo "No action for $action"
    ;;
esac
echo "Action $action complete for $name"
```

The run tool above is implemented as a shell script, and merely reacts to each given `CNAB_ACTION` by printing a message.

See [The Bundle Runtime](103-bundle-runtime.md) for a description on how this tool is used.

Next section: [The Bundle Runtime](103-bundle-runtime.md)