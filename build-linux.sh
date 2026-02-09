#!/bin/bash

# Ollama 英特尔优化版 - Linux 桌面程序构建脚本
# 在 Linux 系统上运行此脚本以构建 Linux 桌面应用

set -e

echo "========================================"
echo "  Ollama 英特尔优化版 - Linux 构建脚本"
echo "========================================"
echo ""

# 检查是否安装了必要的工具
command -v wails >/dev/null 2>&1 || { echo "错误: 未安装 Wails CLI。请先安装 Wails (https://wails.io/docs/gettingstarted/installation)。" >&2; exit 1; }
command -v go >/dev/null 2>&1 || { echo "错误: 未安装 Go。请先安装 Go (https://golang.org/doc/install)。" >&2; exit 1; }
command -v npm >/dev/null 2>&1 || { echo "错误: 未安装 npm。请先安装 Node.js 和 npm。" >&2; exit 1; }

# 构建 Linux 桌面应用
echo "构建 Linux 桌面应用..."
wails build -platform linux -debug

# 复制 Ollama Intel 二进制文件到构建目录
echo "复制 Ollama Intel 二进制文件..."
if [ -d "ollama-intel-ubuntu" ]; then
    cp -r ollama-intel-ubuntu build/bin/
    echo "已复制 Linux 版本二进制文件"
else
    echo "警告: ollama-intel-ubuntu 文件夹不存在"
fi

echo ""
echo "========================================"
echo "  构建完成！"
echo "  可执行文件位置: ./build/bin/ollama-desktop-intel"
echo "  英特尔优化版二进制文件已包含在: ./build/bin/ollama-intel-ubuntu/"
echo "========================================"
echo ""
echo "运行应用:"
echo "  ./build/bin/ollama-desktop-intel"
echo ""