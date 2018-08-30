# Duffle and CNAB

This document describes the Cloud Native Application Bundle (CNAB) format, together with the Duffle reference implementation of a CNAB installer.


You can think of CNAB (Cloud Native Application Bundles) as a package specification for cloud services. In the cloud native world, the term "application" has grown from meaning "an executable and some supporting files" to "a group of interconnected services that work in concert to perform a set of related tasks." (Think microservices plus cloud services.) So today's cloud application may be comprised of containers, VMs, functions, storage, load balancers, service frameworks/orchestrators, etc., all working together.

The central idea behind CNAB is that we can provide a uniform "package management" metaphor for managing the lifecycle of cloud native apps. Duffle is the reference implementation, illustrating how a package manager is built for CNAB.

1. [CNAB Specification](./100-CNAB.md)
    1. [The bundle.json File](101-bundle-json.md)
    2. [The Invocation Image Format](102-invocation-image.md)
    3. [The Bundle Runtime](103-bundle-runtime.md)
    4. [The Claims System](104-claims.md)
    5. [Signing and Provenance](105-signing.md)
2. [Duffle](./200-CNAB.md): The CNAB package manager
    1. [Credential Sets](201-credentialset.md)
    2. [Drivers](202-drivers.md)
    3. [Duffle Build](203-duffle-build.md)
    4. [Duffle Repositories](204-repositories.md)

