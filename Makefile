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

BASE_PACKAGE_NAME := github.com/cnabio/duffle

################################################################################
# Containerized development environment-- or lack thereof                      #
################################################################################

ifneq ($(SKIP_DOCKER),true)
	PROJECT_ROOT := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
	DEV_IMAGE := quay.io/deis/lightweight-docker-go:v0.7.0
	DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-v $(PROJECT_ROOT):/go/src/$(BASE_PACKAGE_NAME) \
		-w /go/src/$(BASE_PACKAGE_NAME) $(DEV_IMAGE)
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

.PHONY: goimports
goimports:
	$(DOCKER_CMD) sh -c "find . -name \"*.go\" | fgrep -v vendor/ | xargs goimports -w -local github.com/cnabio/duffle"

.PHONY: build-drivers
build-drivers:
	mkdir -p bin
	cp drivers/azure-vm/duffle-azvm.sh bin/duffle-azvm
	cd drivers/azure-vm && pip3 install -r requirements.txt

################################################################################
# Tests                                                                        #
################################################################################

.PHONY: lint
lint:
	$(DOCKER_CMD) golangci-lint run --config ./golangci.yml

.PHONY: test
test:
	$(DOCKER_CMD) go test ./...

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

################################################################################
# Example Bundles Build / Validation                                           #
################################################################################

JSON_SCHEMA_URI  := https://cnab.io/v1/bundle.schema.json
JSON_SCHEMA_FILE := /tmp/bundle.schema.json

# bundle-all runs the provided make target on all bundles with a 'duffle.json' file in their directory
define bundle-all
	@for dir in $$(ls -1 examples); do \
		if [[ -e "examples/$$dir/duffle.json" ]]; then \
			BUNDLE=$$dir make --no-print-directory $(1) || exit $$? ; \
		fi ; \
	done
endef

.PHONY: init-validator
init-validator:
	@if ! $$(which ajv > /dev/null 2>&1); then npm install -g ajv-cli; fi
	@if ! [[ -f $(JSON_SCHEMA_FILE) ]]; then curl -sLo $(JSON_SCHEMA_FILE) $(JSON_SCHEMA_URI); fi

.PHONY: validate
validate: init-validator
ifndef BUNDLE
	$(call bundle-all,validate)
else
	@echo "building and validating $(BUNDLE)"
	@cd examples/$(BUNDLE) && \
		duffle build -o bundle.json > /dev/null && \
		ajv test -s $(JSON_SCHEMA_FILE) -d bundle.json --valid
endif

