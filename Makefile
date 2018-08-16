BINDIR    := $(CURDIR)/bin
GO        ?= go
GOFLAGS   :=
TAGS      :=
LDFLAGS   := -w -s

.PHONY: build
build:
	GOBIN=$(BINDIR) $(GO) install $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/duffle/cmd/...
