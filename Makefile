export GOFLAGS=
export GO111MODULE=on

PROGS=create-reviews rebase-prs submit-pr

VERSION := $(shell git describe --tags)
BUILD_TIME := $(shell date +'%Y-%m-%dT%T')
BIN_DIR := ${PWD}/bin

all: ${PROGS}

bin:
	mkdir -p bin

${PROGS}:
	cd cmd/$@ && CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=amd64 go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildTime=$(BUILD_TIME)" -o ${BIN_DIR}/$@

install: ${PROGS}
	cp $(addprefix ${BIN_DIR}/,${PROGS}) ${GOPATH}/bin
