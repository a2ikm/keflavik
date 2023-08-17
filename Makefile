GO_FILES:=$(shell find . -type f -name '*.go' -print)

.PHONY: build_migrate
build_migrate:
	cd cmd/migrate && go build

.PHONY: build_server
build_server:
	cd cmd/server && go build

.PHONY: all
all: cmd/migrate/migrate cmd/server/server
