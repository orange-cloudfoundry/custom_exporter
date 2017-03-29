# Copyright 2017 Orange
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO     ?= GO15VENDOREXPERIMENT=1 go
GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
SCPATH := $(GOPATH)/src/github.com/orange-cloudfoundry/custom_exporter
LCPATH := $(shell pwd)

PROMU       ?= $(GOPATH)/bin/promu
STATICCHECK ?= $(GOPATH)/bin/staticcheck
pkgs         = $(shell $(GO) list ./... )

CUR_DIR                 ?= $(shell basename $(pwd))
BIN_DIR                 ?= $(GOPATH)/bin
SRC_DIR			?= $(GOPATH)/src
PKG_DIR			?= $(GOPATH)/pkg

PREFIX                  ?= $(shell pwd) 

DOCKER_IMAGE_NAME       ?= custom_exporter
DOCKER_IMAGE_TAG        ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

ifeq ($(OS),Windows_NT)
    OS_detected := Windows
else
    OS_detected := $(shell uname -s)
endif

all: format vet test staticcheck build 

pre-build:
	@echo ">> get dependancies"
	@$(GO) get .

style: pre-build
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -prune -o -name '*.go' -print) | grep '^'

test: pre-build
	@echo ">> running tests"
	@$(GO) test -race -short $(pkgs)

format: pre-build
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet: pre-build
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

staticcheck: $(STATICCHECK)
	@echo ">> running staticcheck"
	@$(STATICCHECK) $(pkgs)

buildbin: $(PROMU)
	@echo ">> building binaries"
	@$(PROMU) build --prefix $(PREFIX)

build: pre-build buildbin 

tarball: $(PROMU)
	@echo ">> building release tarball"
	@$(PROMU) tarball --prefix $(PREFIX)

$(GOPATH)/bin/promu promu:
	@GOOS= GOARCH= $(GO) get -u github.com/prometheus/promu

$(GOPATH)/bin/staticcheck:
	@GOOS= GOARCH= $(GO) get -u honnef.co/go/tools/cmd/staticcheck


.PHONY: all style format build test vet tarball docker promu staticcheck

# Declaring the binaries at their default locations as PHONY targets is a hack
# to ensure the latest version is downloaded on every make execution.
# If this is not desired, copy/symlink these binaries to a different path and
# set the respective environment variables.
.PHONY: $(GOPATH)/bin/promu $(GOPATH)/bin/staticcheck

