## commom Makefile

COMMIT_HASH=$(shell git rev-parse --verify head | cut -c-1-8)
DATETIME=$(shell date +%Y-%m-%dT%H:%M:%S%z)
PACKAGES=$(shell go list ./... | grep -v /vendor/)
VETPACKAGES=$(shell go list ./... | grep -v /vendor/ | grep -v /examples/)
GOFILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*")

all: fmt build

.PHONY: fmt vet build

list:
	@echo ${DATETIME}
	@echo ${PACKAGES}
	@echo ${VETPACKAGES}
	@echo ${GOFILES}

fmt:
	@gofmt -s -w ${GOFILES}

test:
	@go test -cpu=1,2,4 -v -tags integration ./...

vet:
	@go vet $(VETPACKAGES)

build:
	go build -o dist/bin/{{.Name}} -ldflags "-X main.buildDate=$(DATETIME) -X main.gitCommit=$(COMMIT_HASH)" main.go

clean:
	@if [ -f dist/bin/{{.Name}} ] ; then rm dist/bin/{{.Name}} ; fi


