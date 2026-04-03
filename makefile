# 项目名称
APP_NAME := pomodora

# 入口文件
MAIN := ./cmd/pomodoro/main.go

# 输出目录
BIN_DIR := bin

# --- 操作系统检测逻辑 ---
ifeq ($(OS),Windows_NT)
    # Windows 环境
    BINARY_NAME := $(APP_NAME).exe
    # 如果是 Windows 且是 GUI 程序，默认加上隐藏控制台参数
    LDFLAGS := -H=windowsgui
    # 清理命令在 Windows 下的兼容处理（如果是用的 MinGW 的 rm 则不需要改）
    RM := rm -rf
else
    # Linux / macOS 环境
    BINARY_NAME := $(APP_NAME)
    LDFLAGS := 
    RM := rm -rf
endif

# 默认目标
.PHONY: all
all: build

# 本机编译（智能识别当前系统）
.PHONY: build
build:
	@echo "Building for current platform..."
	go build -v -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN)

# Linux (交叉编译强制不带 .exe)
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -v -o $(BIN_DIR)/$(APP_NAME)-linux $(MAIN)

# Windows (交叉编译强制带 .exe)
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
	$(RM) $(BIN_DIR)/*