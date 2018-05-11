.DEFAULT_GOAL = build
SHELL = /usr/bin/env bash

BIN := importlint
NAME := go-importlint
PKG := github.com/zchee/$(NAME)

VERSION := 0.0.1
GITBRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GITCOMMIT := $(shell git rev-parse --short --quiet HEAD)

CTIMEVAR=-X $(PKG)/version=$(VERSION) -X $(PKG)/gitCommit=$(GITCOMMIT)
GO_LDFLAGS=-ldflags "-s -w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-s -w $(CTIMEVAR) -extldflags='-static'"

GO_TEST_PKGS ?= $(shell go list ./...)
GO_TEST ?= go test
GO_TEST_TARGET ?= .
GOSRCS = $(shell find . -type f \( -name '*.go' -and -not -iwholename '*testdata*' \))

all: test

.PHONY: build
build: $(NAME)  ## Builds a dynamic executable

$(NAME): *.go
	@echo "+ $@"
	go build -v ${GO_LDFLAGS} -o ${BIN} .

.PHONY: build.static
build.static:  ## Builds a static executable
	@echo "+ $@"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -tags "netgo" -installsuffix netgo ${GO_LDFLAGS} -o ${BIN} .

.PHONY: release
release: build.static


.PHONY: vendor.install
vendor.install: $(shell go env GOPATH)/pkg/$(shell go env GOOS)_$(shell go env GOARCH)/${PKG}:

$(shell go env GOPATH)/pkg/$(shell go env GOOS)_$(shell go env GOARCH)/${PKG}:
	go install -v $(shell go list -f='{{if ne .Name "main"}}{{.ImportPath}}{{end}}' ./vendor/...)


.PHONY: test
test: 
	${GO_TEST} -v -race -run=$(GO_TEST_TARGET) $(PKGS)

.PHONY: vet
vet:
	go vet $(PKGS)

.PHONY: coverage
coverage:  ## Runs go test with coverage
	${GO_TEST} -v -race -run=$(GO_TEST_TARGET) -covermode=atomic -coverpkg=./... -coverprofile=$@.out $(PKGS)


IMAGE_TAG ?= ${GITBRANCH}-${GITCOMMIT}
.PHONY: docker.build
docker.build:  ## Builds the container image
	docker build --rm --pull -t gcr.io/zchee-io/$(PKG_NAME):$(IMAGE_TAG) .


.PHONY: changelog
changelog:  ## Create CHANGELOG.md
	@go get -u github.com/git-chglog/git-chglog/cmd/git-chglog
	git-chglog --output CHANGELOG.md


.PHONY: clean
clean:  ## Cleanup any build binaries or packages
	@echo "+ $@"
	rm -f $(NAME) app *.out


.PHONY: help
help:  ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[33m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
