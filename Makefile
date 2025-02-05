.PHONY: all build test docs docker docker-push clean

# Variables
APP_NAME := bouncer
VERSION := $(shell git describe --tags --always --dirty)
DOCKER_REPO := pjwerneck/bouncer
DOCKER_TAG := $(VERSION)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Build flags
LDFLAGS=-ldflags "-w -s -X main.version=$(VERSION) -X main.commit=$(shell git rev-parse HEAD) -X main.buildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

all: clean build test docs docker

# Change build target to depend on docs
build: docs
	$(GOBUILD) $(LDFLAGS) -o $(APP_NAME)

test:
	$(GOTEST) -v -race ./...

# Add swag init command to docs target
docs:
	swag init -g bouncermain/handlers.go --markdownFiles ./descriptions --output docs

docker:
	docker build -t $(DOCKER_REPO):$(DOCKER_TAG) .
	docker tag $(DOCKER_REPO):$(DOCKER_TAG) $(DOCKER_REPO):latest

docker-push:
	docker push $(DOCKER_REPO):$(DOCKER_TAG)
	docker push $(DOCKER_REPO):latest

clean:
	rm -f $(APP_NAME)
	rm -rf docs

# Development helpers
deps:
	$(GOGET) -u github.com/swaggo/swag/cmd/swag
	$(GOGET) -u github.com/swaggo/http-swagger
	$(GOGET) ./...

run: build
	./$(APP_NAME)

dev: build
	./$(APP_NAME) -debug
