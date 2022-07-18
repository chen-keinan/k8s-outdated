SHELL := /bin/bash

GOCMD=go
GOMOCKS=$(GOCMD) generate ./...
GOMOD=$(GOCMD) mod
GOTEST=$(GOCMD) test


all:
	$(info  "completed running make file for go-simple-config")
fmt:
	@go fmt ./...
lint:
	./lint.sh
tidy:
	$(GOMOD) tidy -v
test:
	$(GOTEST) ./... -coverprofile coverage.md fmt
	$(GOCMD) tool cover -html=coverage.md -o coverage.html
	$(GOCMD) tool cover  -func coverage.md

.PHONY: install-req fmt lint tidy test imports .
