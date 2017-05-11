.PHONY: clean local linux macos docker push

BINARY ?= fake-teams-api
BUILD_FLAGS ?= -v
LOCAL_BUILD_FLAGS ?= $(BUILD_FLAGS) -i
DOCKERFILE = docker/Dockerfile
IMAGE ?= ikitiki/$(BINARY)
VERSION ?= $(shell git describe --tags --always --dirty)
TAG ?= $(VERSION)
SOURCES = main.go

PATH := $(GOPATH)/bin:$(PATH)
SHELL := env PATH=$(PATH) $(SHELL)

default: local

clean:
	rm -rf build

local: build/${BINARY}
linux: build/linux/${BINARY}
macos: build/macos/${BINARY}

build/${BINARY}: ${SOURCES}
	go build -o $@ $(LOCAL_BUILD_FLAGS) $^

build/linux/${BINARY}: ${SOURCES}
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $@ ${BUILD_FLAGS} $^

build/macos/${BINARY}: ${SOURCES}
	GOOS=darwin GOARCH=amd64 go build -o $@ ${BUILD_FLAGS} $^

docker-context: linux
	mkdir -p docker/build/
	cp build/linux/${BINARY} docker/build/

docker: ${DOCKERFILE} docker-context
	cd docker && docker build --rm -t "$(IMAGE):$(TAG)" .

push:
	docker push "$(IMAGE):$(TAG)"