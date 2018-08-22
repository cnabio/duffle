# Duffle: The CNAB Installer

## Getting Started

1. Clone this repo: `$ git clone git@github.com:deis/duffle.git`
2. Vendor dependencies and build the `duffle` binary: `$ make bootstrap build`
3. Set `$DUFFLE_HOME` to `tests/testdata/home`

In powershell:

```
$ $env:DUFFLE_HOME = ".\tests\testdata\home"
$ .\bin\duffle.exe help
```

For everyone else other than Fisher, Radu and Ivan:

```
$ export DUFFLE_HOME="$PWD/tests/testdata/home"
$ ./bin/duffle help
```

Once you have everything configured, you can start installing bundles!

```
$ duffle search
$ duffle install foo foo
```


For more information, see the [docs](docs/).
