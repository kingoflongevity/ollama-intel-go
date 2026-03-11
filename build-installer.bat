@echo off
chcp 65001
setlocal enabledelayedexpansion

echo ==========================================
echo 构建 Ollama 英特尔优化版安装程序
echo ==========================================

:: 检查是否安装了 WiX Toolset
where candle >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 WiX Toolset。请先安装 WiX Toolset。
    echo 下载地址: https://wixtoolset.org/releases/
    pause
    exit /b 1
)

:: 清理旧的构建文件
echo 清理旧的构建文件...
if exist "installer" rmdir /s /q "installer"
mkdir "installer"

:: 编译 WiX 源文件
echo 编译安装程序...
candle -out "installer\installer.wixobj" "installer.wxs"
if %errorlevel% neq 0 (
    echo 错误: 编译 WiX 源文件失败
    pause
    exit /b 1
)

:: 链接生成 MSI 安装程序
echo 生成 MSI 安装程序...
light -out "installer\Ollama-Intel-Setup.msi" "installer\installer.wixobj" -ext WixUIExtension
if %errorlevel% neq 0 (
    echo 错误: 链接安装程序失败
    pause
    exit /b 1
)

echo ==========================================
echo 安装程序构建成功！
echo 输出文件: installer\Ollama-Intel-Setup.msi
echo ==========================================

pause