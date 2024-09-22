# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
BINARY_NAME = useq

.PHONY: all clean test build run deps setup-hooks

all: clean test build

build:
	$(GOBUILD) -v ./...
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/useq

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -rf bin

run: build
	./bin/$(BINARY_NAME)

deps:
	$(GOMOD) download

setup-hooks:
	pre-commit install
