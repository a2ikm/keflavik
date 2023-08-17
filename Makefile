cmd/migrate/migrate:
	cd cmd/migrate
	go build
	cd -

cmd/server/server:
	cd cmd/server
	go build
	cd -

all: cmd/migrate/migrate cmd/server/server
.PHONY: all
