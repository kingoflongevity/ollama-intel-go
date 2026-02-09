@echo off
chcp 65001 > nul
echo.
echo ========================================
echo   Ollama 英特尔优化版 - Windows 构建脚本
echo ========================================
echo.

REM 检查是否安装了必要的工具
where wails >nul 2>nul
if errorlevel 1 (
    echo 错误: 未安装 Wails CLI。请先安装 Wails (https://wails.io/docs/gettingstarted/installation)。
    pause
    exit /b 1
)

where go >nul 2>nul
if errorlevel 1 (
    echo 错误: 未安装 Go。请先安装 Go (https://golang.org/doc/install)。
    pause
    exit /b 1
)

where npm >nul 2>nul
if errorlevel 1 (
    echo 错误: 未安装 npm。请先安装 Node.js 和 npm。
    pause
    exit /b 1
)

REM 构建 Windows 桌面应用
echo 构建 Windows 桌面应用...
wails build -platform windows

REM 复制 Ollama Intel 二进制文件到构建目录
echo 复制 Ollama Intel 二进制文件...
if exist ollama-intel-win (
    xcopy /E /I /Y ollama-intel-win build\bin\ollama-intel-win
    echo 已复制 Windows 版本二进制文件
) else (
    echo 警告: ollama-intel-win 文件夹不存在
)

echo.
echo ========================================
echo   构建完成！
echo   可执行文件位置: build\bin\ollama-desktop-intel.exe
echo   英特尔优化版二进制文件已包含在: build\bin\ollama-intel-win\
echo ========================================
echo.
echo 运行应用:
echo   build\bin\ollama-desktop-intel.exe
echo.
pause