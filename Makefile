#include .env
PACKAGES	?= $(shell go list ./...)
BUILD_DIR ?= build

all:
	go build -o build/main cmd/main.go

test_all:
	@POSTGRES_URI=${POSTGRES_URI} REDIS_URI=${REDIS_URI} go test $(PACKAGES) -v -cover -tags=integration,test

test:
	go test $(PACKAGES) -v -cover -tags=test -failfast

clean:
	rm -rf build