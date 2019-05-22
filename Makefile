SOURCE_FILES?=./...
TEST_PATTERN?=.
TEST_OPTIONS?=
BUILD_NAME?=kibana-config-controller

export PATH := ./bin:$(PATH)
export GO111MODULE := on

# Install all the build and lint dependencies
setup:
	go mod download
.PHONY: setup

# Run all the tests
test:
	go test $(TEST_OPTIONS) -failfast -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m
.PHONY: test

# Run all the tests and opens the coverage report
cover: test
	go tool cover -html=coverage.txt
.PHONY: cover

# gofmt and goimports all go files
fmt:
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done
.PHONY: fmt

# Run all the linters
lint:
	golangci-lint run --tests=false --enable-all --disable=gochecknoglobals ./...
.PHONY: lint

# Run all the tests and code checks
ci: build test lint
.PHONY: ci

# Build a beta version of goreleaser
build:
	go build -o $(BUILD_NAME)
.PHONY: build

# Show to-do items per file.
todo:
	@grep \
		--exclude-dir=vendor \
		--exclude=Makefile \
		--text \
		--color \
		-nRo -E ' TODO:.*|SkipNow' .
.PHONY: todo

.DEFAULT_GOAL := build
