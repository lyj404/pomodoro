# 项目名称
APP_NAME := pomodora

# 入口文件
MAIN := ./cmd/pomodoro/main.go

# 输出目录
BIN_DIR := bin

# 默认目标
.PHONY: all
all: build

# 本机编译（当前系统）
.PHONY: build
build:
	@echo "Building for current platform..."
	go build -v -o $(BIN_DIR)/$(APP_NAME) $(MAIN)

# Linux
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -v -o $(BIN_DIR)/$(APP_NAME)-linux $(MAIN)

# Windows
.PHONY: build-windows
build-windows:
	@echo "Building for Windows (GUI)..."
	GOOS=windows GOARCH=amd64 go build -v -ldflags "-H=windowsgui" -o $(BIN_DIR)/$(APP_NAME).exe $(MAIN)

# 全平台构建
.PHONY: build-all
build-all: build-linux build-windows

# 清理
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)/*
