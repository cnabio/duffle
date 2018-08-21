BINDIR    := $(CURDIR)/bin
GO        ?= go
GOFLAGS   :=
TAGS      :=
LDFLAGS   := -w -s


UNAME := $(shell uname)

ifneq ($(findstring $(UNAME), "Linux", "Darwin"),)
TARGET = duffle
else
TARGET = duffle.exe
endif

GIT_TAG  := $(shell git describe --tags --always)
VERSION  ?= ${GIT_TAG}
LDFLAGS  += -X github.com/deis/duffle/pkg/version.Version=$(VERSION)

.PHONY: build
build:
	$(GO) build $(GOFLAGS) -o $(BINDIR)/$(TARGET) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...
