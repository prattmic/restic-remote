all: client

VERSION=$(shell git describe --long --tags --dirty --always)

client:
	go build -ldflags='-X "main.versionStr=$(VERSION)"' github.com/prattmic/restic-remote/cmd/client

.PHONY: all client
