# Duffle: The CNAB Installer

🚨 **The Duffle project has been archived and is no longer maintained. See https://cnab.io/community-projects/ for current implementations of the CNAB specification.**

Duffle is the reference implementation of the [CNAB specification][cnab]. It
provides a comprehensive mapping of all features of the CNAB Core specification
as of version [1.0.1](https://github.com/cnabio/cnab-spec/tree/cnab-core-1.0.1).

Duffle's main utility, now that much of its internal code has migrated to the
official [cnab-go] Golang library, is to demonstrate working proof-of-concepts
of additions or modifications to applicable CNAB specifications. Of course,
Duffle may still be used to install, manage and author bundles at a low level.

Future conformance updates per the Core and other CNAB specifications are
intended to be added in the aforementioned [cnab-go] library, rather than
in Duffle, which will not be updated going forward. Duffle will remain a reference
implementation of the CNAB spec as of v1.0.1. If you are interested in using an 
up-to-date CNAB tool, check out https://cnab.io/community-projects/#tools.

The community has created implementations of the CNAB spec with
[opinionated takes on authoring bundles][cnab-tools]. Some even use Duffle's
[libraries][cnab-sdk] to handle the CNAB implementation. If you want to make your own CNAB tooling, that is a great place to start!

Learn more about about CNAB and Duffle, check out our [docs](docs/README.md).

## Getting Started

1. [Get the latest Duffle release for your operating system](https://github.com/cnabio/duffle/releases).


2. Run the command to set `duffle` up on your machine:
    ```console
    $ duffle init
    ==> The following new directories will be created:
    /home/janedoe/.duffle
    /home/janedoe/.duffle/bundles
    /home/janedoe/.duffle/logs
    /home/janedoe/.duffle/plugins
    /home/janedoe/.duffle/claims
    /home/janedoe/.duffle/credentials
    ==> The following new files will be created:
    /home/janedoe/.duffle/repositories.json
    ```

3. Build and install your first bundle (you can find the `examples` directory in this repository):
    ```console
    $ duffle build ./examples/helloworld/
    Step 1/6 : FROM alpine:latest
     ---> e21c333399e0
    Step 2/6 : RUN apk update
     ---> Running in 93480e25ef09
    fetch http://dl-cdn.alpinelinux.org/alpine/v3.7/main/x86_64/APKINDEX.tar.gz
    fetch http://dl-cdn.alpinelinux.org/alpine/v3.7/community/x86_64/APKINDEX.tar.gz
    v3.7.3-40-g354ae2b18a [http://dl-cdn.alpinelinux.org/alpine/v3.7/main]
    v3.7.3-38-gb9b86f0506 [http://dl-cdn.alpinelinux.org/alpine/v3.7/community]
    OK: 9055 distinct packages available
     ---> 4123d6b1dfc5
    Step 3/6 : RUN apk add -u bash
     ---> Running in 3db9dd96e10b
    (1/10) Upgrading busybox (1.27.2-r6 -> 1.27.2-r11)
    Executing busybox-1.27.2-r11.post-upgrade
    (2/10) Upgrading libressl2.6-libcrypto (2.6.3-r0 -> 2.6.5-r0)
    (3/10) Installing libressl2.6-libtls (2.6.5-r0)
    (4/10) Installing ssl_client (1.27.2-r11)
    (5/10) Installing pkgconf (1.3.10-r0)
    (6/10) Installing ncurses-terminfo-base (6.0_p20171125-r1)
    (7/10) Installing ncurses-terminfo (6.0_p20171125-r1)
    (8/10) Installing ncurses-libs (6.0_p20171125-r1)
    (9/10) Installing readline (7.0.003-r0)
    (10/10) Installing bash (4.4.19-r1)
    Executing bash-4.4.19-r1.post-install
    Executing busybox-1.27.2-r11.trigger
    OK: 13 MiB in 19 packages
     ---> 5a3670bf25d9
    Step 4/6 : COPY Dockerfile /cnab/Dockerfile
     ---> 58548d5a8553
    Step 5/6 : COPY app /cnab/app
     ---> 46ce2cca5f93
    Step 6/6 : CMD ["/cnab/app/run"]
     ---> Running in d2294cc8b7fd
     ---> 69abe3476d43
    Successfully built 69abe3476d43
    Successfully tagged cnab/helloworld-cnab:87d786be507769a4913c90d85134c85727c85f41
    ==> Successfully built bundle helloworld:0.1.1
    ```

4. Check that it was built:
    ```console
    $ duffle bundle list
    NAME            VERSION DIGEST
    helloworld      0.1.1   fae0c3a28bd850f6a9a2631b9abe4f8244c83ee4
    ```

5. Now run it:
    ```console
    $ duffle credentials generate helloworld-creds helloworld:0.1.1
    $ duffle install helloworld-demo -c helloworld-creds helloworld:0.1.1
    Executing install action...
    hello world
    Install action
    Action install complete for helloworld-demo
    ```

6. Clean up:
    ```console
    $ duffle uninstall helloworld-demo
    Executing uninstall action...
    hello world
    uninstall action
    Action uninstall complete for helloworld-demo
    ```

    *Notes:*
    * To build and install bundles, you need access to a Docker engine - it can be Docker for Mac, Docker for Windows, Docker on Linux, or a remote Docker engine. Duffle uses the Docker engine to build the invocation images, as well as for running actions inside invocation images.
    * Duffle has a driver architecture for different ways of executing actions inside invocation images, and more drivers will be available in the future.
    * Learn more about what a bundle is and its components [here](https://github.com/cnabio/cnab-spec/blob/main/100-CNAB.md).
    * Get a feel for what CNAB bundles written for Duffle look like by referencing the [examples directory](./examples).

## Developing Duffle

See the [Developer's Guide](docs/developing.md).

[cnab]: https://cnab.io
[cnab-tools]: https://cnab.io/community-projects/#tools
[cnab-sdk]: https://cnab.io/community-projects/#sdk
[cnab-go]: https://github.com/cnabio/cnab-go
