# Duffle: The CNAB Installer
![Build Status](http://badges.technosophos.me/v1/github/build/deislabs/duffle/badge.svg?branch=master)

Duffle is a command line tool that allows you to install and manage CNAB bundles. To learn more about about CNAB and duffle, check out the [docs](docs/README.md).

## Getting Started

1. Ensure you're running the latest version of Go (1.11+) by running `$ go version`
    ```console
    $ go version
    go version go1.11 darwin/amd64
    ```

2. Clone this repo:
    ```console
    $ cd $GOPATH/src/github.com/deislabs/
    $ git clone git@github.com:deislabs/duffle.git
    $ cd duffle
    ```

3. Vendor dependencies and build the `duffle` binary:
    ```console
    $ make bootstrap build
    ```

4. Run the command to set duffle up on your machine:
    ```console
    $ cd bin
    $ duffle init
    The following new directories will be created:
    /Users/janedoe/.duffle
    /Users/janedoe/.duffle/logs
    /Users/janedoe/.duffle/plugins
    /Users/janedoe/.duffle/repositories
    /Users/janedoe/.duffle/claims
    /Users/janedoe/.duffle/credentials
    ==> Installing default repositories...
    ==> repo added in 1.096263107s
    ```

5. Build and install your first bundle:

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

6. Check that it was built:
    ```console
    $ duffle bundle list
    NAME            VERSION DIGEST                                          SIGNED?
    helloworld      0.1.0   b2747e5c36369f4c102f4f879caa94e607e5db7e        true
    ```

7. Now run it:
    ```console
    $ duffle credentials generate helloworld-creds helloworld:0.1.0
    $ duffle install helloworld-demo -c helloworld-creds helloworld:0.1.0
    Executing install action...
    hello world
    Install action
    Action install complete for helloworld-demo
    ```

8. Clean up:
    ```console
    $ duffle uninstall helloworld-demo
    Executing uninstall action...
    hello world
    uninstall action
    Action uninstall complete for helloworld-demo
    ```

    *Notes:*
    * Learn more about what a bundle is and its components [here](https://github.com/deislabs/cnab-spec/blob/master/100-CNAB.md).
    * Get a feel for what CNAB bundles look like by referencing the [bundles repo](https://github.com/deislabs/bundles) on github.

# Debugging using VS Code

For instructions on using VS Code to debug the Duffle binary, see [the debugging document](docs/debugging.md).
