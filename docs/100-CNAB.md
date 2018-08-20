# CNAB: Cloud Native Application Bundles

CNAB is a _standard packaging format_ for multi-container cloud native applications. It allows packages to target different runtimes and architectures. It empowers application distributors to package applications for deployment on a wide variety of cloud platforms, cloud providers, and cloud services. It also provides the utilities necessary for delivering multi-container applications in disconnected environments.

CNAB is not a platform-specific tool. While it uses containers for encapsulating installation logic, it remains un-opinionated about what cloud environment it runs in. CNAB developers can bundle applications targeting environments spanning IaaS (like OpenStack or Azure), container orchestrators (like Kubernetes or Nomad), container runtimes (like local Docker or ACI), and cloud services (like object storage or Database as a Service).

This is a working document tracking the current design for CNAB. At a later point, this document will be superseded by a formal specification.

## Approach

The CNAB specification builds on the [Open Container Initiative (OCI)](https://www.opencontainers.org/) family of specifications. It relies upon OCI's content addressable storage (CAS) and multi-architecture support, as well as OCI's specification for cryptographic integrity checking.

The core concept of CNAB is that a bundle is comprised of a _lightweight invocation image_ whose job is to install zero or more cloud components, including (but not limited to): containers, IaaS and PaaS layers, and service frameworks.

The invocation image contains a standardized filesystem layout where metadata and installation data is stored in predictable places. A "main" script or tool contains top-layer orchestration logic. Parameterization and credentialing allow injection of configuration data into the main image.

Actions are sent to the `run` command via environment variable. Actions determine whether a bundle is to be installed, upgraded, downgraded, uninstalled, or merely queried for status.

### Key Terms:

- Bundle: the collection of CNAB data and metadata necessary for installing an application on the designated cloud services
- Invocation Image: The OCI image that contains the bootstrapping and installation logic for the bundle
- Container: An OCI container
- Image: An OCI container image
- Manifest.json: The CNAB file that enumerates the images that are compositionally part of this application, and enumerates which parameters can be overridden.

## Container Registry and Bundle

A CNAB bundle must point to one image (the invocation image). It _may_ point to other images in its `manifest.json` file. Logically, we can talk about those containers as _parts of_ the application. But when stored in a container registry, they will each be stored as independent images.

In other words, while CNAB bundles contain fixed and verifiable references to containers, it is not a precondition of a CNAB bundle that it actually contain the binary data of those containers. This is similar to the way in which Docker and OCI registries store layers independently while representing the top-level layer as a single object (though it is composed of separately stored layers).

## The Invocation Image

An invocation image is composed of the following:

- A file system hierarchy following a defined pattern (below)
- A main entry point, which is an executable (often a script)
- Runtime metadata (Helm charts, Terraform templates, etc)
- The manifest file, which enumerates images used and available parameters, along with other metadata.
- The Dockerfile used to produce the invocation image

### The File System Layout

The following exhibits the filesystem layout:

```yaml
cnab/
├── manifest.json​      # Required
└── Dockerfile​         # Required
└── app​                # Required
    ├── run​            # Required: This is the main entrypoint, and must be executable
    ├── charts​         # Example: Helm charts might go here
    │   └── azure-voting-app​
    │       ├── Chart.yaml​
    │       ├── templates​​
    │       │   └── ...
    │       └── values.yaml​
    └── sfmesh​         # Example: Service Mesh definitions might go here
        └── sfmesh-deploy.json
```

The `app/` directory contains subdirectories, each of which stores configuration for a particular target environment. The `app/run` file must be an executable file that will act as the "main" installer for this CNAB bundle. [NB: Why is this not at the top level? I think it should be]

The contents beneath `app/` are undefined by the spec.

_NOTE:_ While _ad hoc_ data may be placed under `app/`, all possible names under the root container directory (`cnab/`) are considered reserved, and no additional names may be added beyond the spec. For example, creating `cnab/bin/` would violate the specification.

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

## Manifest

The `manifest.json` maps container metadata (name, repository, tag) to placeholders within the bundle. This allows images to be renamed, relabeled, or replaced during the CNAB bundle build operation. It also specifies the parameters that may be overridden in this image, giving tooling the ability to expose configuration options.

Supported substitution formats:

- JSON
- YAML
- XML

```json
{ ​
  "name": "vote",​
  "version": "0.1",​
  "images": [​
    { ​
      "name": "frontend",​
      "uri": "gabrtv.azurecr.io/gabrtv/vote-frontend:a5ff67...",​
      "refs": [​
        {​
          "path": "./charts/azure-voting-app/values.yaml",​
          "field": "AzureVoteFront.deployment.image"​
        }​
      ]​
    },​
    { ​
      "name": "backend",​
      "uri": "gabrtv.azurecr.io/gabrtv/vote-backend:a5ff67...",​
      "refs": [​
        {​
          "path": "./charts/azure-voting-app/values.yaml",​
          "field": "AzureVoteBack.deployment.image"​
        }​
      ]​
    }​
  ]​,
  "parameters": {
    "backend_port" : {
      "type" : "int",
      "defaultValue": 80,
      "minValue": 10,
      "maxValue": 10240,
      "metadata": {
        "description": "The port that the backend will listen on" 
      }
    }
  }
}
```

*TODO:* `manifest.json` probably requires a few more top-level fields, such as something about who published it, and something about the license. A decision on this is deferred until after the PoC

Fields:

- name: the name of the application
- version: The version of the application, which MUST comply with SemVer 2
- images: The list of dependent images
    - name: The image name
    - URI: The image reference (REGISTRY/NAME:TAG). Note that _should_ be a CAS SHA, not a version tag as in the example above.
    - refs: An array listing the locations which refer to this image, and whose values should be replaced by the value specified in URI. Each entry contains the following properties:
        - path: the path of the file where the value should be replaced
        - field:a selector specifying a location (or locations) within that file where the value should be replaced
- parameters: name/value pairs describing a user-overridable parameter
  - <name>: The name of the parameter. In the example above, this is `backend_port`. This
    is mapped to a value definition, which contains the following fields:
    - type: one of string, int, boolean
    - defaultValue: The default value (optional)
    - allowedValues: an array of allowed values (optional)
    - minValue: Minimum value (for ints) (optional)
    - maxValue: Maximum value (for ints) (optional)
    - minLength: Minimum number of characters allowed in the field (for strings) (optional)
    - maxLength: Maximum number of characters allowed in the field (for strings) (optional)
    - metadata: Holds fields that are not used in validation
      - description: A user-friendly description of the parameter

### Field Selectors

For fields, the selectors are based on the _de facto_ format used in tools like `jq`, which is a subset of the [CSS selector](https://www.w3.org/TR/selectors-3/) path. Examples:

- `foo.bar.baz` is interpreted as "find element baz whose parent is bar and whose grandparent is foo".
- `#baz` in XML is "the element whose ID attribute is set to "baz"". It is a no-op in YAML and JSON.
- TODO: Will we need to support attribute selectors?

TODO: How do we specify multiple replacements within a single file?

TODO: How do we specify URI is a VM image (or Jar or other) instead of a Docker-style image? Or do we? And if not, why not?

### Format of Parameters

The structure of a parameters section looks like this:

```json
"parameters": {
    "<parameter-name>" : {
        "type" : "<type-of-parameter-value>",
        "defaultValue": "<default-value-of-parameter>",
        "allowedValues": [ "<array-of-allowed-values>" ],
        "minValue": <minimum-value-for-int>,
        "maxValue": <maximum-value-for-int>,
        "minLength": <minimum-length-for-string-or-array>,
        "maxLength": <maximum-length-for-string-or-array-parameters>,
        "metadata": {
            "description": "<description-of-the parameter>" 
        }
    }
}
```

## Dockerfile

The `Dockerfile` used to build the invocation image must be stored inside of the invocation image. This is to ensure reproducibility, and in order to allow rename operations that require a rebuild.

This is a normal Dockerfile, and may derive from any base image.

Example:

```Dockerfile
FROM ubuntu:latest

COPY ./Dockerfile /cnab/Dockerfile
COPY ./manifest.json /cnab/manfiest.json
COPY ./parameters.json /cnab/parameters.json

RUN curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
RUN helm init --client-only
RUN helm repo add example-stable https://repo.example.com/stable

CMD /cnab/app/run
```

The above example installs and configures Helm inside of a base Ubuntu image. Note that there are no restrictions on what tools may be installed.

## The Main Entry Point

Convention suggests that the main entry point be at `/app/run`. The the `Dockerfile`'s `exec` array must point to this entry point. The main entry point must be an executable of some sort, whether a script or a binary.

The environment will provide the name of the current installation as `$CNAB_INSTALLATION_NAME` and the name of the action will be passed as `$CNAB_ACTION`.

Example:

```bash.
#!/bin/bash
action=$CNAB_ACTION

# TODO: probably do a switch here?
if [[ action == "install" ]]; then
  helm install example-stable/wordpress -n $CNAB_INSTALLATION_NAME
elif [[ action == "delete" ]]; then
  helm delete $CNAB_INSTALLATION_NAME
fi
```

This simple example merely executes Helm, installing the Wordpress chart with the default settings if `install` is sent, or deleting the installation if `delete` is sent.

The following actions are supported:

- install
- upgrade
- delete
- downgrade
- status

None of the actions are required to be implemented. However, none of the actions may return an error if not implemented.

## Overriding Parameters

Parameters that are passed into the container are passed in as environment variables, where each environment variable begins with the prefix `CNAB_P_` and to which the uppercased parameter name is appended. For example `backend_port` will be exposed inside the container as `CNAB_P_BACKEND_PORT`, and thus can be accessed inside of the `run` script:

```bash
#!/bin/sh

echo $CNAB_P_BACKEND_PORT
```

The validation of parameters must happen outside of the CNAB bundle, with the results injected into the bundle before the bundle's `run` is executed (and also in such a way that the `run` has access to these variables.) This works analogously to `CNAB_ACTION` and `CNAB_INSTALLATION_NAME`.

Optionally, installers may additionally mount a file inside of the image that also declares the variables, so that other scripts may take advantage of these values:

```
#!/bin/sh

export CNAB_P_BACKEND_PORT=8080

export CNAB_ACTION=upgrade
export CNAB_INSTALLATION_NAME=wild-panda
```

## TODO

The following topics must be added in the future:

- Signing and security
- Handling of VM images
- Examples of installing services, along with guidance if necessary
- Specification of how data is to be injected into the container (and when we do env vars vs when we do file mounts)