.PHONY: build install snapshot dist test vet lint fmt run clean
OUT := greeter
PKG := github.com/emgag/greeter
VERSION := $(shell git describe --always --dirty --tags)
PKG_LIST := $(shell go list ${PKG}/...)
GO_FILES := $(shell find . -name '*.go')

all: build

build:
	CGO_ENABLED=0 GOOS=linux go build -a -v -o ${OUT} ${PKG}

install:
	CGO_ENABLED=0 GOOS=linux go install -a -v -o ${OUT} ${PKG}

snapshot:
	goreleaser --snapshot --skip-publish --rm-dist

dist:
	goreleaser --rm-dist

test:
	@go test -v ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	revive ./...

fmt:
	@gofmt -l -w -s ${GO_FILES}

clean:
	-@rm ${OUT}

