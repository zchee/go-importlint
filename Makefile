GOPATH ?= $(shell go env GOPATH)
GO_SRCS = $(shell find . -type f -name '*.go' -and -not -iwholename '*vendor*' -and -not -iwholename '*testdata*')
GO_VENDOR_PACKAGES = $(shell go list ./vendor/...)

GO_BUILD_FLAGS := -v
GO_BUILD_TAGS ?=
GO_TEST_FLAGS := -v

.DEFAULT_GOAL = build

build: bin/importlint

bin:
	@mkdir ./bin

bin/importlint: ${GOPATH}/pkg/darwin_amd64/github.com/zchee/go-importlint $(GO_SRCS)
	$(CGO_FLAGS) go build $(GO_BUILD_FLAGS) -o ./bin/importlint ./cmd/importlint

${GOPATH}/pkg/darwin_amd64/github.com/zchee/go-importlint:
	go install $(GO_BUILD_FLAGS) ./vendor/...
