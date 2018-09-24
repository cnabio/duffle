BINDIR    := $(CURDIR)/bin
GOFLAGS   :=
LDFLAGS   := -w -s
TESTFLAGS :=

ifeq ($(OS),Windows_NT)
	TARGET = duffle.exe
	SHELL  = cmd.exe
	CHECK  = where.exe
else
	TARGET = duffle
	SHELL  = bash
	CHECK  = command -v
endif

GIT_TAG  := $(shell git describe --tags --always)
VERSION  := ${GIT_TAG}
LDFLAGS  += -X github.com/deis/duffle/pkg/version.Version=$(VERSION)

.PHONY: default
default: build

.PHONY: build
build:
	go build $(GOFLAGS) -o $(BINDIR)/$(TARGET) -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...

.PHONY: debug
debug:
	go build $(GOFLAGS) -o $(BINDIR)/$(TARGET) github.com/deis/duffle/cmd/...

.PHONY: test
test:
	go test $(TESTFLAGS) ./...

.PHONY: lint
lint:
	gometalinter --config ./gometalinter.json ./...

HAS_DEP          := $(shell $(CHECK) dep)
HAS_GOMETALINTER := $(shell $(CHECK) gometalinter)

.PHONY: build-drivers
build-drivers:
	cp drivers/azure-vm/duffle-azvm.sh bin/duffle-azvm
	cd drivers/azure-vm && pip3 install -r requirements.txt

.PHONY: bootstrap
bootstrap:
ifndef HAS_DEP
	go get -u github.com/golang/dep/cmd/dep
endif
ifndef HAS_GOMETALINTER
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
endif
	dep ensure -vendor-only -v
