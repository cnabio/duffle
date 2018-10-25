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
	golangci-lint run --config ./golangci.yml

HAS_DEP          := $(shell $(CHECK) dep)
HAS_GOLANGCI     := $(shell $(CHECK) golangci-lint)

.PHONY: build-drivers
build-drivers:
	cp drivers/azure-vm/duffle-azvm.sh bin/duffle-azvm
	cd drivers/azure-vm && pip3 install -r requirements.txt

.PHONY: bootstrap
bootstrap:
ifndef HAS_DEP
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
endif
ifndef HAS_GOLANGCI
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin
endif
	dep ensure -vendor-only -v
