GO_FILES:=$(shell find . -type f -name '*.go' -print)

GOBIN:=$(shell go env GOPATH)/bin

.PHONY: build_migrate
build_migrate:
	cd cmd/migrate && go build

.PHONY: build_server
build_server:
	cd cmd/server && go build

.PHONY: all
all: cmd/migrate/migrate cmd/server/server

$(GOBIN)/air:
	curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

.PHONY: install-tools
install-tools: $(GOBIN)/air

.PHONY: start
start: install-tools
	air -c .air.toml
