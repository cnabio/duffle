BINDIR    := $(CURDIR)/bin
GO        ?= go
GOFLAGS   :=
TAGS      :=
LDFLAGS   := -w -s

ifeq ($(OS),Windows_NT)
	TARGET = duffle.exe
else
	TARGET = duffle
	HAS_DEP := $(shell command -v dep;)
endif

GIT_TAG  := $(shell git describe --tags --always)
VERSION  ?= ${GIT_TAG}
LDFLAGS  += -X github.com/deis/duffle/pkg/version.Version=$(VERSION)

.PHONY: build
build:
	$(GO) build $(GOFLAGS) -o $(BINDIR)/$(TARGET) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...

.PHONY: bootstrap
bootstrap:
ifndef HAS_DEP
	go get -u github.com/golang/dep/cmd/dep
endif
	dep ensure -v

