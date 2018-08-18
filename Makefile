BINDIR    := $(CURDIR)/bin
GO        ?= go
GOFLAGS   :=
TAGS      :=
LDFLAGS   := -w -s


ifeq ($(OS),Windows_NT)
	TARGET = duffle.exe
else
	TARGET = duffle
endif

GIT_TAG  := $(shell git describe --tags --always)
VERSION  ?= ${GIT_TAG}
LDFLAGS  += -X github.com/deis/duffle/pkg/version.Version=$(VERSION)

.PHONY: build
build:
<<<<<<< HEAD
	$(GO) build $(GOFLAGS) -o $(BINDIR)/$(TARGET) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...
||||||| merged common ancestors
	GOBIN=$(BINDIR) $(GO) install $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...

.PHONY: build-win
build-win:
	$(GO) build -o $(BINDIR)/duffle.exe $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...
=======
	GOBIN=$(BINDIR) $(GO) install $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...

.PHONY: build-win
build-win:
	$(GO) build -o $(BINDIR)/duffle.exe $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...
>>>>>>> feat: prototype a 'duffle install' command
