# Claims: Tracking an Installation

> This is a proposal for a change in the spec. The original spec does not define how an installation is tracked over time, though it suggests that there are two paths. One is to run the container for the duration of a revision (e.g. the container is the "release record"). The second is to store some sort of release record. This proposal suggests a possible solution for the second method.

A _claim_ (or _claim receipt_) is a record of a CNAB installation. This document describes how the claim system works.

## Concepts of Package Management

A _package_ is a discrete data chunk that can be moved from location to location, and can be unpacked and installed onto a system. All package managers provide some explicit definition of a package and a package format.

When a package is installed, the contents of a package are extracted and placed into the appropriate spaces on the target system. Thus we have an _installation_ (or _instance_) of the package.

There are three core feature categories of a package manager system:

- It can _install_ packages (initially put something onto a system)
- It can _query_ installations (to see what is installed)
- It can _upgrade_ and _delete_ packages (in other words, it can perform additional operations on an existing installation)

Package managers provide a wealth of other features, but those three are standard across all package managers. (For example, most package managers also provide a way to query what packages are available for installation.)

This proposal explains how CNAB records are generated such that continuity can be established across applications.

## Managing State

Fundamentally, package managers provide a state management layer to keep records of what was installed. For example, [homebrew](http://homebrew.sh), a popular macOS package manager, stores records for all installed software in `/usr/local/Cellar`. Helm, the package manager for Kubernetes, stores state records in Kubernetes ConfigMaps located in the system namespace. The Debian Apt system stores state in `/var/run` _WHERE?_. In all of these cases, the stored state allows the package managing system to be able to answer (quickly) the question of whether a given package is installed.

```console
$ brew info cscope
cscope: stable 15.8b (bottled)
Tool for browsing source code
https://cscope.sourceforge.io/
/usr/local/Cellar/cscope/15.8b (10 files, 714.2KB) *
  Poured from bottle on 2017-05-15 at 09:24:58
From: https://github.com/Homebrew/homebrew-core/blob/master/Formula/cscope.rb
```

CNAB does not define where or how records are stored, nor how these records may be used by an implementation. However, it does describe how a CNAB-based system must emit the record to an implementing system.

This is done so that implementors can standardize on a way of relating a release claim (the record of a release) to release operations like `install`, `upgrade`, or `delete`. This, in turn, is necessary if CNAB bundles are expected to be executable by different implementations.

### Anatomy of a Claim

The CNAB claim is defined as a JSON document.

```json
{
    "name": "galloping-pony",
    "revision": "01CN530TF9Q095VTRYP1M8797C",
    "bundle": "technosophos.azurecr.io/cnab/example:0.1.0"
    "created": "TIMESTAMP",
    "modified": "TIMESTAMP",
    "result": {
        "message": "installed vote 0.1.0",
        "action": "install",
        "status": "success"
    },
    "parameters": {
        "SOME_KEY": "SOME_VALUE"
    }
}
```

- name: The name of the _installation_. This can be automatically generated, though humans may need to interact with it. It must be unique within the installation environment, though that constraint must be imposed externally.
- revision: An [ULID](https://github.com/ulid/spec) that must change each time the release is modified.
- bundle: The resource name (e.g. `technosophos.azurecr.io/cnab/example:0.1.0`)
- created: A timestamp indicating when this release claim was first created. This must not be changed after initial creation.
- updated: A timestamp indicating the last time this release claim was modified
- result: The outcome of the bundle's action (e.g. if action is install, this indicates the outcome of the installation.). It is an object with the following fields:
    - message: A human-readable string that communicates the outcome. Error messages may be included in `failure` conditions.
    - action: Indicates the action that the current bundle is in. Valid actions are:
        - install
        - upgrade
        - delete
        - downgrade
        - status
        - unknown
    - status: Indicates the status of the last phase transition. Valid statuses are:
        - success: completed successfully
        - failure: failed before completion
        - underway: in progress. This should only be used if the invocation container must exit before it can determine whether all operations are complete. Note that underway is a _long term status_ that indicates that the installation's final state cannot be determined by the system. For this reason, it should be avoided.
        - unknown: unknown
- parameters: Key/value pairs that were passed in during the operation. These are stored so that the operation can be re-run. Some implementations may choose not to store these for security or portability reasons.

TODO: What is the best timestamp format to use? Does JSON have a preference?

### Why ULIDs?

ULIDs have two properties that are desirable:

- High probability of [uniqueness](https://github.com/ulid/javascript)
- Sortable by time. The first 48 bits contain a timestamp

Compared to a monotonic increment, this has strong advantages when it cannot be assumed that only one actor will be acting upon the CNAB claim record.

### How is the Claim Used?

The claim is used to inform any CNAB tooling about how to address a particular installation. For example, given the claim record, a package manager that implements CNAB should be able to:

- List the _names_ of the installations, given a _bundle name_
- Given an installation's _name_, return the _bundle info_ that is installed under that name
- Given an installation _name_ and a _bundle_, generate a _bundle info_.
    - This is accompanied by running the `install` path in the bundle
- Given an installation's _name_, replace the _bundle info_ with new _bundle info_, and updated the revision with a new ULID, and the modified timestamp with the current time. This is an upgrade operation.
    - This is accompanied by running the `upgrade` path in the bundle
- Given an installation's name, mark the claim as deleted.
    - This is accompanied by running the `uninstall` path in the bundle
    - XXX: Do we want to allow the implementing system to remove the claim from its database (e.g. helm delete --purge) or remain silent on this matter?

To satisfy these requirements, implementations of a CNAB package manager are expected to be able to store and retrieve state information. However, note that nothing in the CNAB specification tells _how or where_ this state information is to be stored. It is _not a requirement_ to store that state information inside of the invocation image. (In fact, this is discouraged.)

## Producing a Claim

The claim is produced outside of the CNAB package. The following claim data is injected
into the invocation container at runtime:

- `$CNAB_INSTALLATION_NAME`: The value of `claim.name`.
- `$CNAB_BUNDLE_NAME`: The name of the present bundle.
- `$CNAB_ACTION`: The action to be performed (install, upgrade, ...)
- `$CNAB_REVISION`: The ULID for the present release revision. (On upgrade, this is the _new_ revision)

The parameters passed in by the user are vetted against `parameters.json` outside of the container, and then injected into the container as environment variables of the form: `$CNAB_P_{parameterName.toUpper}="{parameterValue}"`.

## Calculating the Result

The result object is populated by the result of the invocation image's action. For example, consider the case where an invocation image executes an installation action. The action is represented by the following shell script, and `$CNAB_INSTALLATION_NAME` is set to `my_first_install`:

```bash
#!/bin/bash

set -eo pipefail

helm install stable/wordpress -n $CNAB_INSTALLATION_NAME > /dev/null
kubectl create pod $CNAB_INSTALLATION_NAME > /dev/null
echo "yay!"
```

(Note that we are redirecting data to `/dev/null` just to make the example easier. A production CNAB bundle might choose to include more verbose output.)

If both commands exit with 0, then the resulting claim will look like this:

```json
{
    "name": "my_first_install",
    "revision": "01CN530TF9Q095VTRYP1M8797C",
    "bundle": {
        "uri": "hub.docker.com/technosophos/example_cnab",
        "name": "example_cnab",
        "version": "0.1.0"
    },
    "created": "TIMESTAMP",
    "modified": "TIMESTAMP",
    "result": {
        "message": "yay!",    // From STDOUT (echo)
        "action": "install",  // Determined by the action
        "status": "success"   // if exit code == 0, success, else failure
    }
}
```

## TODO

- Define downgrade
- Define how action is determined, as this is beyond merely running an executable