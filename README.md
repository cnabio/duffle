# Duffle: The CNAB Installer

## Getting Started

1. Ensure you're running the latest version of Go (1.11+)
2. Clone this repo: `$ git clone git@github.com:deis/duffle.git`
3. Vendor dependencies and build the `duffle` binary: `$ make bootstrap build`
4. Call `duffle init`

Once you have everything configured, you can start installing bundles!

```
$ duffle install foo -f examples/helloworld/cnab/bundle.json
```


For more information, see the [docs](docs/).
