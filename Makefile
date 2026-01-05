GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)

BASE_PATH := $(shell pwd)
BUILD_PATH = $(BASE_PATH)/build
WEB_PATH=$(BASE_PATH)/frontend
CORE_PATH=$(BASE_PATH)/core
WEB_DIST_PATH=$(CORE_PATH)/web/dist
CORE_MAIN=$(CORE_PATH)/main.go
CORE_NAME=gpanel

clean:
	rm -rf $(BUILD_PATH)
	rm -rf $(WEB_DIST_PATH)
	cd $(CORE_PATH) && $(GOCLEAN)

build_frontend:
	cd $(WEB_PATH) && npm install && npm run build
	mkdir -p $(WEB_DIST_PATH)
	cp -r $(WEB_PATH)/dist/* $(WEB_DIST_PATH)/

build_core:
	cd $(CORE_PATH) \
	&& CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(CORE_NAME) $(CORE_MAIN)

build_core_linux:
	cd $(CORE_PATH) \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -trimpath -ldflags '-s -w' -o $(BUILD_PATH)/$(CORE_NAME) $(CORE_MAIN)

build: build_frontend build_core

build_linux: build_frontend build_core_linux

.PHONY: clean build build_frontend build_core build_core_linux