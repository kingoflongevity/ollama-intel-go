# 构建 Ollama Intel 安装程序
# 此脚本会自动下载 NSIS 并构建安装程序

$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "构建 Ollama Intel 安装程序" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# 检查 NSIS 是否已安装
$nsisPath = $null
$possiblePaths = @(
    "C:\Program Files (x86)\NSIS\makensis.exe",
    "C:\Program Files\NSIS\makensis.exe"
)

foreach ($path in $possiblePaths) {
    if (Test-Path $path) {
        $nsisPath = $path
        break
    }
}

if ($null -eq $nsisPath) {
    Write-Host "NSIS 未安装，正在下载..." -ForegroundColor Yellow
    
    $nsisUrl = "https://sourceforge.net/projects/nsis/files/NSIS%203/3.09/nsis-3.09-setup.exe/download"
    $nsisInstaller = "$env:TEMP\nsis-setup.exe"
    
    try {
        Write-Host "下载 NSIS 安装程序..." -ForegroundColor Yellow
        Invoke-WebRequest -Uri $nsisUrl -OutFile $nsisInstaller -UseBasicParsing
        
        Write-Host "安装 NSIS..." -ForegroundColor Yellow
        Start-Process -FilePath $nsisInstaller -ArgumentList "/S" -Wait
        
        # 重新检查安装路径
        foreach ($p in $possiblePaths) {
            if (Test-Path $p) {
                $nsisPath = $p
                break
            }
        }
        
        if ($null -ne $nsisPath) {
            Write-Host "NSIS 安装成功!" -ForegroundColor Green
        } else {
            throw "NSIS 安装失败"
        }
    }
    catch {
        Write-Host "自动安装 NSIS 失败: $_" -ForegroundColor Red
        Write-Host "请手动下载并安装 NSIS: https://nsis.sourceforge.io/Download" -ForegroundColor Yellow
        exit 1
    }
}

Write-Host "使用 NSIS: $nsisPath" -ForegroundColor Green

# 创建 installer 目录
if (Test-Path "installer") {
    Remove-Item -Recurse -Force "installer"
}
New-Item -ItemType Directory -Path "installer" -Force | Out-Null

# 构建安装程序
Write-Host "构建安装程序..." -ForegroundColor Yellow
$result = & $nsisPath "installer.nsi" 2>&1
Write-Host $result

if (Test-Path "installer\Ollama-Intel-Setup.exe") {
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host "安装程序构建成功!" -ForegroundColor Green
    Write-Host "输出文件: installer\Ollama-Intel-Setup.exe" -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Green
} else {
    Write-Host "安装程序构建失败!" -ForegroundColor Red
    exit 1
}