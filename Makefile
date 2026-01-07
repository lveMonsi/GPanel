GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)

VERSION ?= dev
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

BASE_PATH := $(shell pwd)
BUILD_PATH = $(BASE_PATH)/build
WEB_PATH=$(BASE_PATH)/frontend
CORE_PATH=$(BASE_PATH)/core
WEB_DIST_PATH=$(CORE_PATH)/web/dist
CORE_MAIN=$(CORE_PATH)/main.go
CORE_NAME=gpanel

LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -s -w"

.PHONY: clean build build_frontend build_core build_linux install deploy help

help:
	@echo "GPanel 构建和管理命令"
	@echo ""
	@echo "可用的命令:"
	@echo "  make build           - 构建前端和后端（当前平台）"
	@echo "  make build_linux     - 构建前端和后端（Linux 平台）"
	@echo "  make build_frontend  - 仅构建前端"
	@echo "  make build_core      - 仅构建后端（当前平台）"
	@echo "  make build_core_linux- 仅构建后端（Linux 平台）"
	@echo "  make clean           - 清理构建产物"
	@echo "  make install         - 安装到系统（需要 root 权限）"
	@echo "  make deploy          - 快速部署到 /opt/gpanel"
	@echo "  make help            - 显示此帮助信息"
	@echo ""
	@echo "环境变量:"
	@echo "  VERSION              - 版本号（默认: dev）"

clean:
	@echo "清理构建产物..."
	rm -rf $(BUILD_PATH)
	rm -rf $(WEB_DIST_PATH)
	cd $(CORE_PATH) && $(GOCLEAN)
	rm -rf $(WEB_PATH)/dist
	rm -rf $(WEB_PATH)/node_modules
	@echo "清理完成"

build_frontend:
	@echo "构建前端..."
	cd $(WEB_PATH) && npm install && npm run build:pro
	mkdir -p $(WEB_DIST_PATH)
	cp -r $(WEB_PATH)/dist/* $(WEB_DIST_PATH)/
	@echo "前端构建完成"

build_core:
	@echo "构建后端 ($(GOOS)/$(GOARCH))..."
	cd $(CORE_PATH) \
	&& CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -trimpath $(LDFLAGS) -o $(BUILD_PATH)/$(CORE_NAME) $(CORE_MAIN)
	@echo "后端构建完成"

build_core_linux:
	@echo "构建后端 (linux/amd64)..."
	cd $(CORE_PATH) \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -trimpath $(LDFLAGS) -o $(BUILD_PATH)/$(CORE_NAME) $(CORE_MAIN)
	@echo "后端构建完成"

build: clean build_frontend build_core
	@echo "构建完成！二进制文件位于: $(BUILD_PATH)/$(CORE_NAME)"

build_linux: clean build_frontend build_core_linux
	@echo "Linux 构建完成！二进制文件位于: $(BUILD_PATH)/$(CORE_NAME)"

install:
	@echo "安装 GPanel 到系统..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "错误: 需要 root 权限来安装"; \
		exit 1; \
	fi
	mkdir -p /opt/gpanel
	mkdir -p /var/log/gpanel
	cp $(BUILD_PATH)/$(CORE_NAME) /opt/gpanel/
	chmod +x /opt/gpanel/$(CORE_NAME)
	cp config.yaml /opt/gpanel/ 2>/dev/null || echo "警告: config.yaml 不存在，将使用默认配置"
	cp gpanel.service /etc/systemd/system/
	systemctl daemon-reload
	@echo "安装完成！"
	@echo "使用以下命令启动服务:"
	@echo "  sudo systemctl enable gpanel"
	@echo "  sudo systemctl start gpanel"

deploy: build_linux
	@echo "部署 GPanel 到 /opt/gpanel..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "错误: 需要 root 权限来部署"; \
		exit 1; \
	fi
	mkdir -p /opt/gpanel
	mkdir -p /var/log/gpanel
	cp $(BUILD_PATH)/$(CORE_NAME) /opt/gpanel/
	chmod +x /opt/gpanel/$(CORE_NAME)
	cp config.yaml /opt/gpanel/ 2>/dev/null || true
	cp gpanel.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl enable gpanel
	@echo "部署完成！"
	@echo "使用以下命令启动服务:"
	@echo "  sudo systemctl start gpanel"