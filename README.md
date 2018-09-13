# Duffle: The CNAB Installer
[![Build Status](https://cnlabs.visualstudio.com/duffle/_apis/build/status/deis.duffle)](https://cnlabs.visualstudio.com/duffle/_build/latest?definitionId=1)

Duffle is a command line tool that allows you to install and manage CNAB bundles. To learn more about about CNAB and duffle, check out the [docs](docs/000-index.md).

## Getting Started

1. Ensure you're running the latest version of Go (1.11+) by running `$ go version`
```console
$ go version
go version go1.11 darwin/amd64
```

2. Clone this repo:
```console
$ cd $GOPATH/src/github.com/deis/
$ git clone git@github.com:deis/duffle.git
$ cd duffle
```

3. Vendor dependencies and build the `duffle` binary:
```
$ make bootstrap build
```

4. Run the command to set duffle up on your machine:
```console
duffle init
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

5. Search for available bundles:
```console
$ duffle search
NAME      	REPOSITORY                 	VERSION
helloazure	github.com/deis/bundles.git	0.1.0
hellohelm 	github.com/deis/bundles.git	0.1.0
helloworld	github.com/deis/bundles.git	1.0
```
_Notes:_
_* These bundles are hosted in the [bundles repo](https://github.com/deis/bundles) on github. There is on going work on the bundle repository design and to support http based repositories much of which you can see via [this pull request](https://github.com/deis/duffle/pull/184).
* Learn more about what a bundle is and its components [here](https://github.com/deis/duffle/blob/master/docs/100-CNAB.md)._

6. Install your first bundle:
```
$ duffle install foo -f examples/helloworld/cnab/bundle.json
```


For more information, see the [docs](docs/000-index.md).
