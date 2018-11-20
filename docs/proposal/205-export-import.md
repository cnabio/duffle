# Duffle Export/Import

This document describes `$ duffle export` and `$ duffle import`.

Duffle export can be used to package a bundle into a compressed archive.
The export command takes an argument to a bundle. It then looks for the bundle manifest file (`bundle.json` or `bundle.cnab`). It uses the `name` and `version` information in the bundle manifest to create a directory. It copies the bundle manifest to the newly created directory. Within the newly created directory, it creates an `artifacts/` directory and pull a tar archive of each image and invocation image specified in the bundle manifest file.

It the compresses the bundle
