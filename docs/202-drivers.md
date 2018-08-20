# Duffle Drivers

Duffle can perform various actions, such as `install` and `uninstall`. These actions interact with CNAB images. For example, a `duffle install my_app example:1.2.3` will execute the invocation image for the CNAB `example:1.2.3`.

It is possible that users may want to determine which runtime is used to execute the invocation image. For example, if the image is a Docker image, a user may prefer to run it with Docker, or they may prefer to run it with an alternative client (like `rkt`), or they may prefer to execute it within a cloud service like Azure Container Instances.

To accommodate this case, Duffle provides multiple drivers.

## Image Types

Duffle inspects the manifest.json to determine the _image type_ of the image. (FIXME: Actually, right now it just assumes Docker. See `cmd/duffle/install.go`.) Each image can be a particular type, such as a Docker image or a QCOW image.

Drivers can support different image types. The `docker` driver supports `oci` and `docker` image types. The `debug` driver supports all image types. When creating a new driver, developers must specify which image types that driver can support.

## Requesting a Driver

By default, the `docker` driver is used. But a user may choose to override the driver by specify the `-d DRIVERNAME` flag on the relevant operation.

A driver will (MUST) fail if the given driver cannot handle the CNAB bundle's invocation image.

## Built-In Drivers

Duffle has a few default drivers:

- `docker`: Runs OCI and Docker images using a local Docker client (currently requires the Docker CLI)
- `aci`: Runs a Docker image inside of Azure ACI (currently requires the `az` commandline client)
- `debug`: Dumps the info that was sent to the driver, and exits
- `???`: Runs a VM image on... (TODO: We want a VM version if possible. Maybe `az` for this?)

## Driver-Specific Configuration

Configuration is sent to drivers via environment variables. For example, setting the environment variable `$DEBUG` will turn on debugging for most drivers, while setting the environment variable `$AZ_RESOURCE_GROUP` will set the resource group setting on drivers that use `az`.

## Custom Drivers

Custom drivers are implemented following the pattern of `git` plugins: When a driver is requested (`-d mydriver`) and Duffle does not have built-in support for that driver name, it will seek `$PATH` for a program named `duffle-mydriver` (prepending `duffle-` to the driver name).

If a suitable executable is found, Duffle will execute that program, using the action requested. The environment in which that command executes will be pre-populated with the current environment variables. Credential sets will be passed as well. And the operation will be sent as a JSON body on STDIN:

```json
{
  "Installation": "foo",
  "Action": "install",
  "Parameters": {
      "backend_port": 80,
      "hostname": "localhost"
  },  
  "Credentials": [
      {
          "type": "env",
          "name": "SERVICE_TOKEN",
          "value": "secret"
      }
  ],
  "Image": "bar:1.2.3",
  "ImageType": "docker",
  "Revision": "aaaaaa1234567890"
}
```

The custom driver is expected to take that information and execute the appropriate action for the given image.

### Parameters and Credentials for Custom Drivers

The parameters and credentials that are sent to a custom driver will have already been verified.

Parameters will contain the validated, merged parameters. They will be validated against the parameters specification contained in the `manifest.json` file.

Credentials will be loaded and converted to their `destination` format.

A driver MAY withhold some credentials from the underlying system it represents, but it MUST inform the user if doing so.

A driver MUST NOT remove any of the parameters, and must inject them into the image in the format specified by the CNAB specification.