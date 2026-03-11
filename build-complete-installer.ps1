# Build complete installer for ollama intel
# This script creates a complete installer with all necessary files

$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Building ollama intel Complete Installer" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# Define paths
$buildDir = "build\bin"
$ollamaBinSource = "ollama-bin"
$ollamaBinTarget = "$buildDir\ollama-bin"

# Step 1: Build the application with Wails
Write-Host "`nStep 1: Building application with Wails..." -ForegroundColor Yellow
& wails build -nsis
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build application!" -ForegroundColor Red
    exit 1
}

# Step 2: Copy ollama-bin to build directory
Write-Host "`nStep 2: Copying ollama-bin files..." -ForegroundColor Yellow
if (Test-Path $ollamaBinTarget) {
    Remove-Item -Recurse -Force $ollamaBinTarget
}
Copy-Item -Path $ollamaBinSource -Destination $ollamaBinTarget -Recurse -Force

# Step 3: Create a ZIP package with all files
Write-Host "`nStep 3: Creating distribution package..." -ForegroundColor Yellow
$releaseDir = "release\ollama-intel"
if (Test-Path "release") {
    Remove-Item -Recurse -Force "release"
}
New-Item -ItemType Directory -Path $releaseDir -Force | Out-Null

# Copy main executable (the one without "installer" in name)
$exeFiles = Get-ChildItem "$buildDir\*.exe" | Where-Object { $_.Name -notlike "*installer*" -and $_.Name -notlike "*debug*" }
foreach ($file in $exeFiles) {
    Copy-Item $file.FullName $releaseDir -Force
    Write-Host "Copied: $($file.Name)" -ForegroundColor Green
}

# Copy ollama-bin directory
$ollamaBinDest = "$releaseDir\ollama-bin"
Copy-Item -Path $ollamaBinTarget -Destination $ollamaBinDest -Recurse -Force
Write-Host "Copied: ollama-bin directory" -ForegroundColor Green

# Step 4: Create README file
Write-Host "`nStep 4: Creating README file..." -ForegroundColor Yellow
$readmeContent = @"
# ollama intel - Intel Optimized Ollama Desktop Application

## System Requirements
- Windows 10/11 64-bit
- Intel Core Ultra or Intel Arc GPU (recommended)
- 8GB+ RAM

## Installation
1. Extract all files to a directory (e.g., C:\Program Files\ollama-intel)
2. Run the main executable
3. The application will automatically start the built-in Ollama service

## Features
- Built-in Intel optimized Ollama (version 2.3.0b20250923)
- Intel GPU acceleration support
- Graphical management interface
- Model download and management
- AI chat functionality

## Directory Structure
- Main executable - Main application
- ollama-bin/ - Ollama binary files and dependencies

## Notes
- First startup may take some time to initialize
- Model files are stored in the .ollama folder in the user directory by default
- To uninstall, simply delete the installation directory

## Support
- Project: https://github.com/ollama/ollama
- Intel Optimized: https://www.modelscope.cn/models/Intel/ollama/summary
"@

Set-Content -Path "$releaseDir\README.md" -Value $readmeContent -Encoding UTF8

# Step 5: Create ZIP package
Write-Host "`nStep 5: Creating ZIP package..." -ForegroundColor Yellow
$zipPath = "release\ollama-intel-windows.zip"
if (Test-Path $zipPath) {
    Remove-Item -Force $zipPath
}
Compress-Archive -Path "$releaseDir\*" -DestinationPath $zipPath -Force

# Step 6: Summary
Write-Host "`n==========================================" -ForegroundColor Green
Write-Host "Build completed successfully!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Generated files:" -ForegroundColor Cyan
Write-Host "  - Distribution package: $zipPath" -ForegroundColor White
Write-Host "  - Wails installer: $buildDir\ollama英特尔-amd64-installer.exe" -ForegroundColor White
Write-Host ""
Write-Host "Note: The Wails installer may not include ollama-bin directory." -ForegroundColor Yellow
Write-Host "Use the ZIP package for complete distribution." -ForegroundColor Yellow