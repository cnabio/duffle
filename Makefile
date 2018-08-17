BINDIR    := $(CURDIR)/bin
GO        ?= go
GOFLAGS   :=
TAGS      :=
LDFLAGS   := -w -s

GIT_TAG  := $(shell git describe --tags --always)
VERSION  ?= ${GIT_TAG}
LDFLAGS  += -X github.com/deis/duffle/pkg/version.Version=$(VERSION)

.PHONY: build
build:
	GOBIN=$(BINDIR) $(GO) install $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...

.PHONY: build-win
build-win:
	$(GO) build -o $(BINDIR)/duffle.exe $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...
