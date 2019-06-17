SHELL ?= /bin/bash

.DEFAULT_GOAL := build

################################################################################
# Version details                                                              #
################################################################################

# This will reliably return the short SHA1 of HEAD or, if the working directory
# is dirty, will return that + "-dirty"
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty --match=NeVeRmAtCh)

################################################################################
# Go build details                                                             #
################################################################################

BASE_PACKAGE_NAME := github.com/deislabs/duffle

################################################################################
# Containerized development environment-- or lack thereof                      #
################################################################################

ifneq ($(SKIP_DOCKER),true)
	PROJECT_ROOT := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
	DEV_IMAGE := golang:1.12-stretch
	DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-v $(PROJECT_ROOT):/go/src/$(BASE_PACKAGE_NAME) \
		-w /go/src/$(BASE_PACKAGE_NAME) $(DEV_IMAGE)
	INSTALL_DEP := make install-dep &&
	INSTALL_GOLANGCI_LINT := make install-golangci-lint &&
endif

################################################################################
# Binaries and Docker images we build and publish                              #
################################################################################

ifdef DOCKER_REGISTRY
	DOCKER_REGISTRY := $(DOCKER_REGISTRY)/
endif

ifdef DOCKER_ORG
	DOCKER_ORG := $(DOCKER_ORG)/
endif

BASE_IMAGE_NAME := duffle

ifdef VERSION
	MUTABLE_DOCKER_TAG := latest
else
	VERSION            := $(GIT_VERSION)
	MUTABLE_DOCKER_TAG := edge
endif

LDFLAGS              := -w -s -X $(BASE_PACKAGE_NAME)/pkg/version.Version=$(VERSION)

IMAGE_NAME         := $(DOCKER_REGISTRY)$(DOCKER_ORG)$(BASE_IMAGE_NAME):$(VERSION)
MUTABLE_IMAGE_NAME := $(DOCKER_REGISTRY)$(DOCKER_ORG)$(BASE_IMAGE_NAME):$(MUTABLE_DOCKER_TAG)

################################################################################
# Utility targets                                                              #
################################################################################

.PHONY: dep
dep:
	$(DOCKER_CMD) $(INSTALL_DEP) dep ensure -v

.PHONY: install-dep
install-dep:
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | INSTALL_DIRECTORY=/usr/local/bin sh

.PHONY: install-golangci-lint
install-golangci-lint:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b /usr/local/bin v1.16.0

.PHONY: goimports
goimports:
	$(DOCKER_CMD) sh -c "find . -name \"*.go\" | fgrep -v vendor/ | xargs goimports -w -local github.com/deislabs/duffle"

.PHONY: build-drivers
build-drivers:
	mkdir -p bin
	cp drivers/azure-vm/duffle-azvm.sh bin/duffle-azvm
	cd drivers/azure-vm && pip3 install -r requirements.txt

################################################################################
# Tests                                                                        #
################################################################################

# Verifies there are no discrepancies between desired dependencies and the
# tracked, vendored dependencies
.PHONY: verify-vendored-code
verify-vendored-code:
	$(DOCKER_CMD) dep check

.PHONY: lint
lint:
	$(DOCKER_CMD) $(INSTALL_GOLANGCI_LINT) golangci-lint run --config ./golangci.yml

.PHONY: test
test:
	$(DOCKER_CMD) go test -v ./...

################################################################################
# Build / Publish                                                              #
################################################################################

.PHONY: build
build: build-all-bins build-image

.PHONY: build-all-bins
build-all-bins:
	$(DOCKER_CMD) bash -c "LDFLAGS=\"$(LDFLAGS)\" scripts/build.sh"

# You can make this target build for a specific OS and architecture using GOOS
# and GOARCH environment variables.
.PHONY: build-bin
build-bin:
	$(DOCKER_CMD) bash -c "GOOS=\"$(GOOS)\" GOARCH=\"$(GOARCH)\" LDFLAGS=\"$(LDFLAGS)\" scripts/build.sh"

# This target is for contributor convenience.
.PHONY: build-%
build-%:
	$(DOCKER_CMD) bash -c "GOOS=$* LDFLAGS=\"$(LDFLAGS)\" scripts/build.sh"

.PHONY: build-image
build-image:
	docker build \
		-t $(IMAGE_NAME) \
		--build-arg LDFLAGS='$(LDFLAGS)' \
		.
	docker tag $(IMAGE_NAME) $(MUTABLE_IMAGE_NAME)

.PHONY: push
push: push-image

.PHONY: push-image
push-image:
	docker push $(IMAGE_NAME)
	docker push $(MUTABLE_IMAGE_NAME)
