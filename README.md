# Duffle: The CNAB Installer

## Getting Started

1. Clone this repo: `$ git clone git@github.com:deis/duffle.git`
2. Vendor dependencies and build the `duffle` binary: `$ make bootstrap build`
3. Call `duffle init`

Once you have everything configured, you can start installing bundles!

```
$ duffle install foo -f examples/helloworld/cnab/bundle.json
```


For more information, see the [docs](docs/).
