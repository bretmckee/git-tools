export GOFLAGS=
export GO111MODULE=on

PROGS=create-reviews rebase-prs submit-pr
INSTALL_DIR=$(HOME)/bin

VERSION := $(shell git describe --tags)
BUILD_TIME := $(shell date +'%Y-%m-%dT%T')
BIN_DIR := $(PWD)/bin

all: $(PROGS)

$(BIN_DIR) $(INSTALL_DIR):
	mkdir -p $@

${PROGS}:
	cd cmd/$@ && CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildTime=$(BUILD_TIME)" -o ${BIN_DIR}/$@

install: $(PROGS) $(INSTALL_DIR)
	cp $(addprefix $(BIN_DIR)/,$(PROGS)) $(INSTALL_DIR)
	for script in scripts/*; do \
		ln -s -f $${PWD}/$${script} $${HOME}/scripts;\
	done
