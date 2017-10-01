all: client

VERSION=$(shell git describe --long --tags --dirty --always)

LDFLAGS=-X "main.versionStr=$(VERSION)"

client:
	go build -ldflags='$(LDFLAGS)' github.com/prattmic/restic-remote/cmd/client

client.exe:
	GOOS=windows go build -ldflags='$(LDFLAGS)' github.com/prattmic/restic-remote/cmd/client

.PHONY: all client client.exe
