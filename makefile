SHELL := /bin/bash

# ==============================================================================
# Running tests

test:
	go test ./... -count=1

# ==============================================================================
# Modules support

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	go get -u -t -d -v ./...
	go mod tidy
	go mod vendor

# ==============================================================================
