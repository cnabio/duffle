# Image Relocation

Before installing a bundle, the user may wish to _relocate_ the images referenced by the bundle to a suitable registry.

A highly desirable property of image relocation is that the image digests of the relocated images are the same as those of the original images. This gives the user confidence that the relocated images consist of the same bits as the original images.

Using your own registry has several advantages:
* You can control when images are updated or deleted.
* You can host the registry on a private network for security or other reasons.
* You won't be impacted by outages of other registries.

## Relocating Image Names
An image name consists of a domain name (with optional port) and a path. The image name may also contain a tag and/or a digest. The domain name determines the network location of a registry. The path consists of one or more components separated by forward slashes. The first component is, by convention in some registries, a user name providing access control to the image.

Let’s look at some examples:
* The image name `docker.io/istio/proxyv2` refers to an image with user name `istio` residing in the docker hub registry at `docker.io`.
* The image name `projectriff/builder:v1` is short-hand for `docker.io/projectriff/builder:v1` which refers to an image with user name `projectriff` also residing at `docker.io`.
* The image name `gcr.io/cf-elafros/knative-releases/github.com/knative/serving/cmd/autoscaler@sha256:deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef` refers to an image with user name `cf-elafros` residing at `gcr.io`.

When an image is relocated to a registry, the domain name is set to that of the registry and the path modified so that it starts with a given value such as the user name that will own the image.

The path of a relocated image may:
* Include the original user name for readability.
* Be “flattened” to accommodate registries which do not support paths with more than two components.
* End with a hash of the image name (to reduce the risk of collisions).
* Preserve any tag in the original image name.
* Preserve any digest in the original image name.

For instance, when relocated to a registry at `example.com` with user name `user`, the above image names might become something like this:
* `example.com/user/istio-proxyv2-f93a2cacc6cafa0474a2d6990a4dd1a0`
* `example.com/user/projectriff-builder-a4a25a99d48adad8310050be267a10ce:v1`
* `example.com/user/cf-elafros-knative-releases-github.com-knative-serving-cmd-autoscaler-c74d62dc488234d6d1aaa38808898140@sha256:deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef`

The hash added to the end of the relocated image path should not depend on any tag and/or digest in
the original image name. This ensures a one-to-one mapping between repositories. In other words, if:

    x maps to y

where `x` and `y` are image names without tags or digests, then

    x:t maps to y:t (for all tags t)

and

    x@d maps to y@d (for all digests d).

## duffle relocate
`duffle relocate` relocates the images referenced by a bundle and creates a new bundle with an updated image map. Relocation is restricted to images with type "docker" and "oci". Images of other types are not relocated.
 
The `--repository-prefix` flag determines the repositories for the relocated images. Each image is given a name starting with the given prefix and pushed to the repository.

For example, if the repository prefix is `example.com/user`, the image `istio/proxyv2` is relocated
to a name starting with `example.com/user/` and pushed to a repository hosted by `example.com`.

Issue `duffle relocate -h` for more information.

## Example Scenarios

### Thin Bundle Relocation

The [Acme Corporation](https://en.wikipedia.org/wiki/Acme_Corporation) needs to install some "forge" software packaged as a thin bundle (`forge.json`).
Acme is used to things going wrong, so they have evolved sophisticated processes to protect their systems.
In particular, all their production software must be loaded from Acme-internal repositories.
This protects them from outages when an external repository goes down.
It also gives them complete control over what software they run in production.

So Acme needs to pull the images referenced by `forge.json` from external repositories and store them in an Acme-internal registry.
This will be done in a DMZ with access to the internet and write access to the internal registry.

Suppose their internal registry is hosted at `registry.internal.acme.com` and they have created a user `smith` to manage the forge software. They can use `duffle relocate` to
relocate the images to their registry as follows:
```bash
duffle relocate forge.json forge-relocated --input-bundle-is-file --repository-prefix=registry.internal.acme.com/smith

```

They can now install the forge software from the relocated bundle:
```bash
duffle install acme-forge forge-relocated ...

```

The invocation image then installs the software such that its images are loaded from the internal registry.

### Thick Bundle Relocation

Not yet supported. See [issue 705](https://github.com/deislabs/duffle/issues/705)
