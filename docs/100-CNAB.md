# CNAB: Cloud Native Application Bundles

CNAB is a _standard packaging format_ for multi-container cloud native applications. It allows packages to target different runtimes and architectures. It empowers application distributors to package applications for deployment on a wide variety of cloud platforms, cloud providers, and cloud services. It also provides the utilities necessary for delivering multi-container applications in disconnected environments.

CNAB is not a platform-specific tool. While it uses containers for encapsulating installation logic, it remains un-opinionated about what cloud environment it runs in. CNAB developers can bundle applications targeting environments spanning IaaS (like OpenStack or Azure), container orchestrators (like Kubernetes or Nomad), container runtimes (like local Docker or ACI), and cloud services (like object storage or Database as a Service).

This is a working document tracking the current design for CNAB. At a later point, this document will be superseded by a formal specification.

## Summary

The CNAB format is a packaging format for a broad range of cloud services. It specifies a pairing of a `bundle.json` to define the app, and an _invocation image_ to install the app.

CNAB defines a `bundle.json` file, which contains the following information:

- The name and version of the bundle
- Information about locating and running the _invocation image_ (the installer program)
- A list of user-overridable parameters that this package accepts
- The signed list of OCI/Docker images and VM images that this bundle will install
- A list of credential paths or environment variables that this bundle requires to execute

The `bundle.json` can be stored on its own, or as part of a _packed archive_, which is a CNAB bundle that includes the JSON file and exported images (including the invocation image).

Bundles use cryptographic signing extensively. Images and signed, and their signature is then embedded into the `bundle.json`. The `bundle.json` may then be signed using a public/private key system to ensure that it has not been tampered with.

Finally, this document describes a format for invocation images, including file system layout and a functional description of how an invocation image is installed.

## Approach

The current cloud architecture involves three different executable units: Virtual Machines (VMs, Containers (e.g. Docker and OCI) and Functions-as-a-Service (FaaS). Along with these executable units, many managed cloud services (from load balancers to databases) are provisioned via REST APIs. Our goal is to provide a packaging format that can enable application providers and developers with a way of installing a multi-component application into a cloud environment, leveraging all of these cloud components.

The core concept of CNAB is that a bundle is comprised of a _lightweight invocation image_ whose job is to install zero or more cloud components, including (but not limited to): containers, functions, VMs, IaaS and PaaS layers, and service frameworks.

The invocation image contains a standardized filesystem layout where metadata and installation data is stored in predictable places. A _run_ script or tool contains top-layer instructions. Parameterization and credentialing allow injection of configuration data into the main image.

_Actions_ are sent to the `run` command via environment variable. Actions determine whether a bundle is to be installed, upgraded, downgraded, uninstalled, or merely queried for status.

The key data required to bootstrap a CNAB bundle is contained in a `bundle.json` file, which provides name and version information, along with instructions on how to obtain and run the invocation image.

### Key Terms:

- Application: The functional unit composed by the components described in a bundle
- Bundle: the collection of CNAB data and metadata necessary for installing an application on the designated cloud services
- `bundle.json`: The file that contains information on executing a bundle.
- Image: Used generically, a container image (Docker) or a VM image 
- Invocation Image: The  image that contains the bootstrapping and installation logic for the bundle

### Contents

- [The bundle.json File](101-bundle-json.md)
- [The Invocation Image Format](102-invocation-image.md)
- [The Bundle Runtime](103-bundle-runtime.md)
- [The Claims System](104-claims.md)
- [Signing and Provenance](105-signing.md)


## History

- The `bundle.json` is now a stand-alone artifact, not part of the invocation image
- The initial draft of the spec included a `manifest.json`, a `ui.json` and a `parameters.json`. The `bundle.json` is now the only metadata file, containing what was formerly spread across those three.
- The top-level `/cnab` directory was added to the bundle format due to conflicts with file hierarchy
- The signal handling method was discarded after early research showed its limitations. The replacement uses environment variables to trigger actions.
- The credentialset and claims concepts were introduced to cover areas upon which the original spec was silent
- The generic action `run` has been replaced by specific actions: `install`, `uninstall`, `upgrade`, `status`.