GOVERSION=$(shell go version)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

# if tag doesnt exists, show revision
VERSION=$(shell git describe --tags --abbrev=7 --always)
# REVISION=$(shell git rev-parse --short HEAD)

ifeq ($(VERSION),$(REVISION))
VERSION=v0.0.0
endif

# ANSI color
RED=\033[31m
GREEN=\033[32m
RESET=\033[0m

COLORIZE_PASS=sed ''/PASS/s//$$(printf "$(GREEN)PASS$(RESET)")/''
COLORIZE_FAIL=sed ''/FAIL/s//$$(printf "$(RED)FAIL$(RESET)")/''

NAME=strsql

.PHONY: \
	runner-test \
	test \
	install \
	build \
	runner \
	dep-clean \
	vet \
	fmt-diff \
	install-lint \
	lint \


runner-test:
	go test -v ./... | $(COLORIZE_PASS) | $(COLORIZE_FAIL)

test:
	go test -v -race ./...

runner:
	realize start

fmt-diff:
	bash -c "diff -u <(echo -n) <(gofmt -d ./)"

clean-dep:
	go mod tidy

vet: 
	go vet ./...

install-lint:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b build/bin

lint:
	./build/bin/golangci-lint run --tests --disable-all --enable=golint --enable=govet --enable=unused --enable=deadcode --enable=ineffassign --enable=structcheck

build:
	go build -trimpath -ldflags "-X github.com/smith-30/strsql/cmd.version=$(VERSION) -X github.com/smith-30/strsql/cmd.appName=$(NAME)" -o build/${GOOS}_${GOARCH}/${NAME} main.go

install:
	go install ./cmd/${CMD}
