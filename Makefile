GOTOOLS = github.com/golang/lint/golint \
          gopkg.in/alecthomas/gometalinter.v2

# Temporary patch to avoid build failing because of the outdated documentation example
PKGS = $(shell go list ./... | egrep -v "\/examples")

all: lint test

deps: tools
	@go get -v -d -t $(PKGS)

test: deps
	@go test -race $(PKGS)

tools:
	@go get $(GOTOOLS)
	@gometalinter.v2 --install > /dev/null

tools-update:
	@go get -u $(GOTOOLS)
	@gometalinter.v2 --install

lint: deps
	@gometalinter.v2 --config=.gometalinter.json $(PKGS)

lint-all: deps
	@gometalinter.v2 --config=.gometalinter.json --enable=interfacer --enable=gosimple $(PKGS)

.PHONY: all deps test lint lint-all tools tools-update