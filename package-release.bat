@echo off
chcp 65001
setlocal enabledelayedexpansion

echo ==========================================
echo 打包 Ollama 英特尔优化版发布版本
echo ==========================================

:: 创建发布目录
set "RELEASE_DIR=release\Ollama-Intel-v1.0.0"
if exist "release" rmdir /s /q "release"
mkdir "%RELEASE_DIR%"
mkdir "%RELEASE_DIR%\ollama-bin"

echo 复制应用程序文件...
:: 复制主程序
copy "build\bin\ollama-desktop-intel.exe" "%RELEASE_DIR%\" >nul

:: 复制Ollama二进制文件
echo 复制 Ollama 二进制文件...
copy "ollama-bin\ollama.exe" "%RELEASE_DIR%\ollama-bin\" >nul
copy "ollama-bin\ollama-lib.exe" "%RELEASE_DIR%\ollama-bin\" >nul
copy "ollama-bin\ollama-serve.bat" "%RELEASE_DIR%\ollama-bin\" >nul
copy "ollama-bin\start-ollama.bat" "%RELEASE_DIR%\ollama-bin\" >nul
copy "ollama-bin\sycl-ls.exe" "%RELEASE_DIR%\ollama-bin\" >nul

:: 复制DLL文件
echo 复制 DLL 文件...
copy "ollama-bin\*.dll" "%RELEASE_DIR%\ollama-bin\" >nul

:: 复制文档文件
echo 复制文档文件...
copy "ollama-bin\README.txt" "%RELEASE_DIR%\ollama-bin\" >nul
copy "ollama-bin\README.zh-CN.txt" "%RELEASE_DIR%\ollama-bin\" >nul
copy "ollama-bin\ollama-version.txt" "%RELEASE_DIR%\ollama-bin\" >nul

:: 创建启动脚本
echo 创建启动脚本...
(
echo @echo off
echo chcp 65001
echo echo 正在启动 Ollama 英特尔优化版...
echo start "" "%%~dp0ollama-desktop-intel.exe"
) > "%RELEASE_DIR%\启动程序.bat"

:: 创建README文件
echo 创建 README 文件...
(
echo # Ollama 英特尔优化版 v1.0.0
echo.
echo ## 系统要求
echo - Windows 10/11 64位
echo - Intel Core Ultra 或 Intel Arc GPU（推荐）
echo - 8GB+ 内存
echo.
echo ## 安装说明
echo 1. 解压本压缩包到任意目录
echo 2. 运行 `启动程序.bat` 或 `ollama-desktop-intel.exe`
echo 3. 程序会自动启动内置的 Ollama 服务
echo.
echo ## 功能特性
echo - 内置 Ollama 英特尔优化版（版本 2.3.0b20250923）
echo - 支持 Intel GPU 加速
echo - 图形化管理界面
echo - 模型下载和管理
echo - AI 对话功能
echo.
echo ## 目录结构
echo - `ollama-desktop-intel.exe` - 主程序
echo - `ollama-bin/` - Ollama 二进制文件和依赖库
echo - `启动程序.bat` - 启动脚本
echo.
echo ## 注意事项
echo - 首次启动可能需要一些时间初始化
echo - 模型文件默认存储在用户目录下的 `.ollama` 文件夹中
echo - 如需卸载，直接删除本目录即可
echo.
echo ## 技术支持
echo - 项目地址: https://github.com/ollama/ollama
echo - Intel 优化版: https://www.modelscope.cn/models/Intel/ollama/summary
echo.
) > "%RELEASE_DIR%\README.md"

:: 打包为ZIP
echo 打包为 ZIP 文件...
cd release
powershell -Command "Compress-Archive -Path 'Ollama-Intel-v1.0.0' -DestinationPath 'Ollama-Intel-v1.0.0-Windows.zip' -Force"
cd ..

echo ==========================================
echo 打包完成！
echo 输出文件: release\Ollama-Intel-v1.0.0-Windows.zip
echo ==========================================

pause