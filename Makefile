SHELL := /bin/bash
GO ?= go
BUILD_GOPATH = ${GOPATH}:$(CURDIR)
BUILD_USER ?= $(shell whoami)
BUILD_HOST ?= $(shell hostname)
BUILD_DATE ?= $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD := ${BUILD_USER}@${BUILD_HOST} on ${BUILD_DATE}
REV := $(shell git rev-parse --short HEAD 2> /dev/null)
PKG = "github.com/f-secure-foundry/armoryctl"

.PHONY: armoryctl test

all: armoryctl

# requires the following dependencies in $GOPATH
#   honnef.co/go/tools/cmd/staticcheck
#   github.com/kisielk/errcheck
check:
	@${GO} vet ./...
	@${GOPATH}/bin/staticcheck ./...
	@${GOPATH}/bin/errcheck ./...

test:
	@${GO} test ./...

armoryctl:
	@GOPATH="${BUILD_GOPATH}" ${GO} build -v \
	  -gcflags=-trimpath=${CURDIR} -asmflags=-trimpath=${CURDIR} \
	  -ldflags "-s -w -X '${PKG}/internal.Revision=${REV}' -X '${PKG}/internal.Build=${BUILD}'" \
	  armoryctl.go
	@echo -e "compiled armoryctl ${REV} (${BUILD})"
