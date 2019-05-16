PROJECT         := duffle
ORG             := deislabs
DOCKER_REGISTRY ?= $(ORG)
BINDIR          := $(CURDIR)/bin
GOFLAGS         :=
LDFLAGS         := -w -s
TESTFLAGS       :=
INSTALL_DIR     := /usr/local/bin

ifeq ($(OS),Windows_NT)
	TARGET = $(PROJECT).exe
	SHELL  = cmd.exe
	CHECK  = where.exe
else
	TARGET = $(PROJECT)
	SHELL  ?= bash
	CHECK  ?= which
endif

GIT_TAG   := $(shell git describe --tags --always)
VERSION   ?= ${GIT_TAG}
# Replace + with -, for Docker image tag compliance
IMAGE_TAG ?= $(subst +,-,$(VERSION))
LDFLAGS   += -X github.com/$(ORG)/$(PROJECT)/pkg/version.Version=$(VERSION)

.PHONY: default
default: build

.PHONY: build
build:
	go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(TARGET) github.com/$(ORG)/$(PROJECT)/cmd/...

.PHONY: install
install:
	install $(BINDIR)/$(TARGET) $(INSTALL_DIR)

CX_OSES  = linux windows darwin
CX_ARCHS = amd64

.PHONY: build-release
build-release:
	@for os in $(CX_OSES); do \
		echo "building $$os"; \
		for arch in $(CX_ARCHS); do \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(PROJECT)-$$os-$$arch github.com/$(ORG)/$(PROJECT)/cmd/...; \
		done; \
		if [ $$os = 'windows' ]; then \
			mv $(BINDIR)/$(PROJECT)-$$os-$$arch $(BINDIR)/$(PROJECT)-$$os-$$arch.exe; \
		fi; \
	done

.PHONY: debug
debug:
	go build $(GOFLAGS) -o $(BINDIR)/$(TARGET) github.com/$(ORG)/$(PROJECT)/cmd/...

.PHONY: build-docker-bin
build-docker-bin:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(TARGET) github.com/$(ORG)/$(PROJECT)/cmd/...

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_REGISTRY)/$(PROJECT):$(IMAGE_TAG) .

.PHONY: docker-push
docker-push:
	docker push $(DOCKER_REGISTRY)/$(PROJECT):$(IMAGE_TAG)

.PHONY: test
test:
	go test $(TESTFLAGS) ./...

.PHONY: lint
lint:
	golangci-lint run --config ./golangci.yml

HAS_DEP          := $(shell $(CHECK) dep)
HAS_GOLANGCI     := $(shell $(CHECK) golangci-lint)
HAS_GOIMPORTS    := $(shell $(CHECK) goimports)
GOLANGCI_VERSION := v1.16.0
HAS_COCOV_XML := $(shell command -v gocov-xml;)
HAS_GOCOV := $(shell command -v gocov;)
HAS_GO_JUNIT_REPORT := $(shell command -v go-junit-report;)

.PHONY: build-drivers
build-drivers:
	cp drivers/azure-vm/$(PROJECT)-azvm.sh bin/$(PROJECT)-azvm
	cd drivers/azure-vm && pip3 install -r requirements.txt

.PHONY: bootstrap
bootstrap:
ifndef HAS_DEP
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
endif
ifndef HAS_GOLANGCI
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin $(GOLANGCI_VERSION)
endif
ifndef HAS_GOIMPORTS
	go get -u golang.org/x/tools/cmd/goimports
endif
	dep ensure -vendor-only -v
ifndef HAS_COCOV_XML
	go get github.com/AlekSi/gocov-xml
endif
ifndef HAS_GOCOV
	go get -u github.com/axw/gocov/gocov
endif
ifndef HAS_GO_JUNIT_REPORT
	go get github.com/jstemmer/go-junit-report
endif

.PHONY: coverage
coverage:
	go test -v -coverprofile=coverage.txt -covermode count ./... 2>&1 | go-junit-report > report.xml
	gocov convert coverage.txt > coverage.json
	gocov-xml < coverage.json > coverage.xml

.PHONY: goimports
goimports:
	find . -name "*.go" | fgrep -v vendor/ | xargs goimports -w -local github.com/deislabs/duffle
