# Ollama 英特尔优化版桌面应用

基于 [Ollama Desktop](https://github.com/jianggujin/ollama-desktop) 项目的复刻版本，专为英特尔硬件优化。

## 功能特性

- **现代化桌面界面**: 使用 Vue 3 + Element Plus 构建的现代化用户界面
- **模型管理**: 可视化本地模型管理，支持拉取、删除、查看模型信息
- **在线模型**: 浏览和下载 Ollama 官方支持的模型
- **聊天对话**: 友好的聊天界面，支持自定义模型参数
- **服务控制**: 一键启动/停止 Ollama 服务
- **英特尔硬件优化**: 针对英特尔 GPU (Arc 系列、集成显卡) 和 CPU 优化
- **跨平台支持**: Windows 和 Linux 桌面应用

## 英特尔优化特性

本应用专门针对英特尔硬件进行优化：

- **英特尔 GPU 加速**: 支持英特尔 Arc 系列 GPU 和集成显卡
- **OneAPI 支持**: 利用英特尔 OneAPI 工具套件进行性能优化
- **MKL 数学库**: 使用英特尔 MKL 库加速数学运算
- **支持设备**:
  - Intel Core Ultra 处理器
  - Intel Core 11th - 14th 代处理器
  - Intel Arc A 系列 GPU
  - Intel Arc B 系列 GPU

更多详细信息请参考: [英特尔优化版 Ollama 文档](https://www.modelscope.cn/models/Intel/ollama/summary)

## 项目结构

```
ollama-desktop-intel/
├── build/                    # 构建输出目录
├── frontend/                 # Vue 前端代码
│   ├── src/
│   │   ├── views/           # 页面组件
│   │   ├── router/          # 路由配置
│   │   └── ...
│   └── wailsjs/             # Wails 自动生成的绑定
├── app.go                   # Go 后端主逻辑
├── main.go                  # 应用入口点
├── wails.json               # Wails 配置文件
├── build-windows.bat        # Windows 构建脚本
├── build-linux.sh           # Linux 构建脚本
└── README.md                # 本文档
```

## 构建说明

### 前置要求

- **Go** 1.21 或更高版本
- **Node.js** 18 或更高版本 (包含 npm)
- **Wails CLI** v2.11.0 或更高版本

安装 Wails CLI:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Windows 构建

1. 打开命令提示符或 PowerShell
2. 进入项目目录
3. 运行构建脚本:
   ```bash
   .\build-windows.bat
   ```
   或手动构建:
   ```bash
   wails build -platform windows
   ```

构建完成后，可执行文件位于 `build\bin\ollama-desktop-intel.exe`

### Linux 构建

1. 在 Linux 系统上打开终端
2. 进入项目目录
3. 运行构建脚本:
   ```bash
   chmod +x build-linux.sh
   ./build-linux.sh
   ```
   或手动构建:
   ```bash
   wails build -platform linux
   ```

构建完成后，可执行文件位于 `build/bin/ollama-desktop-intel`

## 开发模式运行

```bash
# 启动开发服务器
wails dev
```

开发模式下，前端热重载可用，Go 后端代码修改后需要重启应用。

## 使用说明

1. **启动应用**: 运行生成的可执行文件
2. **启动 Ollama 服务**: 在仪表板点击"启动服务"按钮
3. **管理模型**: 在"模型管理"页面查看、拉取、删除模型
4. **开始聊天**: 在"聊天"页面选择模型并开始对话
5. **查看在线模型**: 在"在线模型"页面浏览可下载的模型

## 技术栈

- **后端**: Go (Wails 框架)
- **前端**: Vue 3 + Vite + Element Plus
- **UI 组件**: Element Plus
- **构建工具**: Wails CLI
- **打包**: Wails 内置打包系统

## 许可证

MIT License

## 致谢

- [Ollama Desktop](https://github.com/jianggujin/ollama-desktop) - 原始项目
- [Wails](https://wails.io) - Go 桌面应用框架
- [Vue.js](https://vuejs.org) - 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org) - Vue 3 UI 组件库

## 贡献

欢迎提交 Issue 和 Pull Request。

## 支持

如有问题，请提交 GitHub Issue 或参考相关文档。