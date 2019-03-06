# Duffle: The CNAB Installer
![Build Status](http://badges.technosophos.me/v1/github/build/deislabs/duffle/badge.svg?branch=master)
[![Waffle.io - Columns and their card count](https://badge.waffle.io/deislabs/duffle.svg?columns=In%20Progress,Needs%20Review,Done)](https://waffle.io/deislabs/duffle)


Duffle is a command line tool that allows you to install and manage CNAB bundles. To learn more about about CNAB and duffle, check out the [docs](docs/README.md).

## Getting Started

1. [Get the latest Duffle release for your operating system](https://github.com/deislabs/duffle/releases).


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
    ==> Generating a new secret keyring at /home/janedoe/.duffle/secret.ring
    ==> Generating a new signing key with ID janedoe <janedoe@computer>
    ==> Generating a new public keyring at /home/janedoe/.duffle/public.ring
    ```

3. Build and install your first bundle (you can find the `examples` directory in this repository):
    ```console
    $ duffle build ./examples/helloworld/
    Step 1/5 : FROM alpine:latest
    ---> 196d12cf6ab1
    Step 2/5 : RUN apk add -u bash
    ---> Using cache
    ---> 54b3a85c5c2e
    Step 3/5 : COPY Dockerfile /cnab/Dockerfile
    ---> Using cache
    ---> cd6f4ff8d83d
    Step 4/5 : COPY app /cnab/app
    ---> 38a482447ffd
    Step 5/5 : CMD ["/cnab/app/run"]
    ---> Running in 8b22055f0a37
    ---> e5c795c2a1f4
    Successfully built e5c795c2a1f4
    Successfully tagged deislabs/helloworld-cnab:64dfc7c4d825fe87506dbaba6ab038eafe2a486d
    ==> Successfully built bundle helloworld:0.1.0
    ```

4. Check that it was built:
    ```console
    $ duffle bundle list
    NAME            VERSION DIGEST                                          SIGNED?
    helloworld      0.1.0   b2747e5c36369f4c102f4f879caa94e607e5db7e        true
    ```

5. Now run it:
    ```console
    $ duffle credentials generate helloworld-creds helloworld:0.1.0
    $ duffle install helloworld-demo -c helloworld-creds helloworld:0.1.0
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
    * Learn more about what a bundle is and its components [here](https://github.com/deislabs/cnab-spec/blob/master/100-CNAB.md).
    * Get a feel for what CNAB bundles look like by referencing the [bundles repo](https://github.com/deislabs/bundles) on github.

# Debugging using VS Code

For instructions on using VS Code to debug the Duffle binary, see [the debugging document](docs/debugging.md).
