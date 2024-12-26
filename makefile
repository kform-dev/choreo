# Copyright 2024 Nokia
# Licensed under the Apache License 2.0
# SPDX-License-Identifier: Apache-2.0

VERSION ?= latest
REGISTRY ?= europe-docker.pkg.dev/srlinux/eu.gcr.io
PROJECT ?= choreo
IMG ?= $(REGISTRY)/${PROJECT}:$(VERSION)

REPO = github.com/kform-dev/choreo
USERID := 10000


## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.15.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec


.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./apis/..."
	GOBIN=$(LOCALBIN) go generate ./...

.PHONY: manifests
manifests: controller-gen generate ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./apis/..." output:crd:artifacts:config=artifacts


.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: genproto
genproto:
	go run ./tools/apiserver-runtime-gen \
		-g go-to-protobuf \
		--module $(REPO) \

.PHONY:
docker-build: ## Build docker image with the manager.
	ssh-add ./keys/id_rsa 2>/dev/null; true
	docker build --build-arg USERID="$(USERID)" . -t ${IMG} --ssh default="$(SSH_AUTH_SOCK)"

.PHONY: docker-push
docker-push: docker-build ## Push docker image with the manager.
	docker push ${IMG}

.PHONY: all
all: fmt vet manifests ## Build manager binary.
	go build -ldflags "-X github.com/kform-dev/choreo/cmd/choreoctl/commands.version=${GIT_COMMIT}" -o $(LOCALBIN)/choreoctl -v cmd/choreoctl/main.go

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)