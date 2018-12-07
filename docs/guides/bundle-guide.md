# Build Your First Bundle!

A `bundle` is a CNAB package. In its slimmest form, a bundle contains metadata (in a `bundle.cnab` file) which points to a image (we call that the "invocation image") that contains instructions (in a `run` file) on how to install and configure a multi-component cloud native application.

In this guide, you will create a CNAB bundle which does `echo` commands for various actions similar to the [helloworld](https://github.com/deislabs/duffle/blob/master/examples/helloworld/cnab/app/run) example.

## Create the Directory Structure

```console
$ duffle create helloworld
Creating helloworld
$ cd helloworld
```

### The `cnab/` directory

The `cnab/` directory is created for you. It is where the logic and any supporting files for the invocation image lives. In this directory, an `app/run` file exists as the entrypoint to the invocation image.

```console
$ cat cnab/app/run
#!/bin/bash
action=$CNAB_ACTION

if [[ action == "install" ]]; then
echo "hey I am installing things over here"
elif [[ action == "uninstall" ]]; then
echo "hey I am uninstalling things now"
fi
```

For this example, we are going to modify the `run` file to act on environment variables that have been set by the CNAB runtime.

In `cnab/app/run`:

```bash
#!/bin/sh

set -eo pipefail

action=$CNAB_ACTION
name=$CNAB_INSTALLATION_NAME
param=${CNAB_P_HELLO}

echo "$param world"
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
    downgrade)
    echo "Downgrade action"
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

The `$CNAB_ACTION` environment variable describes what action is to be performed. `$CNAB_INSTALLATION_NAME` is the name of the instance of the installation of the bundle. Any environment variable that has a prefix of `$CNAB_P_` is a parameter that either had a default set or was passed in at runtime by the end user.

## Defining the Invocation Image

When we created the application, a Dockerfile for the invocation image was added. This Dockerfile describes the invocation image runtime, copying the `app/run` file to `/cnab/app/run`.

```bash
$ cat cnab/Dockerfile
FROM alpine:latest

COPY Dockerfile /cnab/Dockerfile
COPY cnab/app /cnab/app

CMD ["/cnab/app/run"]
```

## Building the bundle

When we build the bundle, a manifest file is written to `$DUFFLE_HOME/bundles`. This file contains metadata about the bundle, information on the required parameters necessary credentials for a successful installation, and content digests for each image specified for the bundle installation. Let's use the `duffle build` command to build the bundle and inspect its output:

```console
$ duffle build .
[...]
==> Successfully built bundle helloworld:0.1.0
```

After the bundle has been built, we can inspect the bundle:

```console
$ duffle show helloworld:0.1.0 -r
-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA256

{
  "name": "helloworld",
  "version": "0.1.0",
  "description": "A short description of your bundle",
  "invocationImages": [
    {
      "imageType": "docker",
      "image": "helloworld-cnab:f4a5a0dac16c61442ccf19611c06526fcb2e5a74"
    }
  ],
  "images": [],
  "parameters": null,
  "credentials": null
}
-----BEGIN PGP SIGNATURE-----

wsDcBAEBCAAQBQJb9ErXCRA58VPKJbKbxwAARqYMADtWlk3aLj/NVxNpd3GaqlI6
tUiW/1T5zIFEWYsJgSC3ammN9z266Uf2q+tDC+jt7A5+sZTGHujn/8FCuURLRkp7
UVU7ot1xJb8nWUyDLeZjX6yG+eI7XbqjIbt17+bp59XYVRlgJtT1/gLxqm1gh8IQ
D2TLeuOdfI3bstupFEN7AoZWPG5XTYbtQCC9TdBLw70LLGl2f7L4Ll7RFDEJEjx+
NVCjJEWaYAw7DP1kHUpl67vhkFVeptnbr99uC9aEFUo6fImeuczIU0S9K9g+2Vxf
wcs+XgWKDBkAN9hF/tnaIVsIeHrPJZ9oviEbYDeVqIKUlUBBbNblVTVnjC7shfjF
1SQ4AGhkIgf9gFan7KkERlAp3dcjh5XDgZ7/ijVGGItlbIE1p8+KBm2FRwJfox69
L9aitybWBnt5EIm3w4YIYsMuMZuPM/0taoKH9nzNv4lQsKYqeX6tOD36aDx4fys1
NSKekvE5KfHYU3t+3rUtJRphoVsSr3cNFldsVCVuzQ==
=iFSF
-----END PGP SIGNATURE-----
```

As shown by the output, `duffle build` cryptographically signs the bundle to ensure that it has not been tampered with.

## Watch it Work

```console
$ duffle install helloworld helloworld:0.1.0
hello world
Install action
Action install complete for helloworld
```

The output of `duffle install` comes from the run script. `hello world` is printed before the defined action is executed. In this example, the action being executed is the install action. In this example, the install action is running `echo 'Install Action'` At the end, the run script prints a message indication the action has been completed.

## Notes and Next steps

- There are alternatives to defining a custom `run` tool. See examples of more complex and different bundles [here](https://github.com/deislabs/bundles).
- Read more about the CNAB spec in the [docs](https://github.com/deislabs/cnab-spec/blob/master/100-CNAB.md)
