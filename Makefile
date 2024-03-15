include .env
PACKAGES	?= $(shell go list ./...)

test:
	 @POSTGRES_URI=${POSTGRES_URI} REDIS_URI=${REDIS_URI} go test $(PACKAGES) -v -cover

