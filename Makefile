include .env
PACKAGES	?= $(shell go list ./...)

all:
	go build -o build/main cmd/main.go

test_all:
	@POSTGRES_URI=${POSTGRES_URI} REDIS_URI=${REDIS_URI} go test $(PACKAGES) -v -cover -tags=integration

test:
	go test $(PACKAGES) -v -cover

clean:
	rm -rf build