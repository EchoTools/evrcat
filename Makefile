VERSION=$(shell git describe --tags --always --dirty=+)
COMMIT=$(shell git rev-parse --short HEAD)
PWD=$(shell pwd)

export version=${VERSION}
export commit=${COMMIT}

GCFLAGS=-gcflags "all=-trimpath ${PWD}" -asmflags "all=-trimpath ${PWD}"
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION} -X main.commitID=${COMMIT}"
ASMFLAGS=-asmflags "all=-trimpath ${PWD}"

DEPS=vendor/modules.txt main.go

.PHONY: all clean vendor


all: evrcat evrcat.exe

vendor/modules.txt: go.mod go.sum
	GOWORK=off go mod vendor

# Makefile for evrcat

evrcat: $(DEPS)
	GOWORK=off GOOS=linux go build -o evrcat -trimpath -mod=vendor ${GCFLAGS} ${LDFLAGS} main.go

evrcat.exe: $(DEPS)
	GOWORK=off GOOS=windows go build -o evrcat.exe -trimpath -mod=vendor ${GCFLAGS} ${LDFLAGS} main.go

clean:
	rm -f evrcat evrcat.exe

