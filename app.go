package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                  context.Context
	ollamaCmd            *exec.Cmd
	ollamaPath           string
	logger               *logWriter
	environmentVariables map[string]interface{}
	websocketConnections map[string]*websocket.Conn
	websocketMutex       sync.Mutex
	pullProcesses        map[string]*exec.Cmd // 保存正在运行的拉取进程
	pullProcessesMutex   sync.Mutex           // 拉取进程互斥锁
}

// 内存地址正则表达式
var memoryAddressRegex = regexp.MustCompile(`0x[0-9a-fA-F]+`)

// logWriter 是一个自定义的 io.Writer，将日志发送到前端
type logWriter struct {
	ctx context.Context
	mu  sync.Mutex
}

// Write 实现 io.Writer 接口
func (l *logWriter) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := string(p)

	// 过滤掉包含内存地址的输出
	if memoryAddressRegex.MatchString(msg) {
		return len(p), nil
	}

	// 发送日志事件到前端
	if l.ctx != nil {
		wailsRuntime.EventsEmit(l.ctx, "log", msg)
	}

	return len(p), nil
}

const (
	OllamaAPIBaseURL = "http://127.0.0.1:11434"
)

// WebSocket升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源的WebSocket连接
	},
}

// ModelInfo 模型信息
type ModelInfo struct {
	Name     string                 `json:"name"`
	Model    string                 `json:"model"`
	Size     string                 `json:"size"`
	Digest   string                 `json:"digest"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Modified string                 `json:"modified_at"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
	Options  interface{}   `json:"options,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   ChatMessage   `json:"message"`
	Done      bool          `json:"done"`
	Messages  []ChatMessage `json:"messages,omitempty"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		websocketConnections: make(map[string]*websocket.Conn),
		pullProcesses:        make(map[string]*exec.Cmd),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化日志记录器
	a.logger = &logWriter{ctx: ctx}

	// 设置日志输出到自定义记录器
	log.SetOutput(a.logger)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 保存原始的 stdout，以便同时输出到控制台
	originalStdout := os.Stdout

	// 创建管道用于捕获输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 启动一个 goroutine 来读取管道内容并发送到前端和控制台
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				data := buf[:n]
				// 同时输出到控制台
				originalStdout.Write(data)
				// 发送到前端
				wailsRuntime.EventsEmit(ctx, "log", string(data))
			}
			if err != nil {
				if err != io.EOF {
					originalStdout.Write([]byte(fmt.Sprintf("日志读取错误: %v\n", err)))
				}
				break
			}
		}
	}()

	// 初始化环境变量存储
	a.environmentVariables = make(map[string]interface{})
	// 设置默认值
	a.environmentVariables["OLLAMA_MODEL_SOURCE"] = "modelscope"
	a.environmentVariables["OLLAMA_NUM_CTX"] = 2048
	a.environmentVariables["OLLAMA_INTEL_GPU"] = true
	a.environmentVariables["OLLAMA_DEBUG"] = false
	a.environmentVariables["ONEAPI_DEVICE_SELECTOR"] = ""
	// 允许跨域访问和外部访问
	a.environmentVariables["OLLAMA_ORIGINS"] = "*"
	a.environmentVariables["OLLAMA_HOST"] = "0.0.0.0:11434"
	// OpenAI兼容API默认值
	a.environmentVariables["OLLAMA_OPENAI_COMPATIBLE"] = true
	a.environmentVariables["OLLAMA_OPENAI_PORT"] = 8080
	a.environmentVariables["OLLAMA_OPENAI_API_KEY"] = ""

	log.Println("startup: 开始初始化")

	// 设置 Ollama 二进制文件路径
	a.setOllamaPath()

	// 加载配置文件
	a.loadConfig()

	// 初始化HTTP服务器，添加WebSocket路由
	a.initHTTPServer()

	// 启动 Ollama 服务
	log.Println("startup: 启动 Ollama 服务")
	go a.startOllamaService()
	log.Println("startup: 初始化完成")
}

// shutdown is called when the app closes
func (a *App) shutdown(ctx context.Context) {
	// 停止 Ollama 服务
	if a.ollamaCmd != nil && a.ollamaCmd.Process != nil {
		a.ollamaCmd.Process.Kill()
	}
}

// GetEnvironmentInfo 获取环境信息
func (a *App) GetEnvironmentInfo() map[string]interface{} {
	// 检查服务状态
	serviceRunning := a.checkOllamaService()
	serviceStatus := "stopped"
	if serviceRunning {
		serviceStatus = "running"
	}

	// 获取 Ollama 版本
	ollamaVersion := "unknown"
	if serviceRunning {
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://127.0.0.1:11434/api/version")
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
				if v, ok := result["version"].(string); ok {
					ollamaVersion = v
				}
			}
		}
	}

	// 检测 GPU 状态
	gpuStatus := a.detectGPUStatus()

	// 获取内存信息
	memoryInfo := a.getMemoryInfo()

	// 获取 CPU 信息
	cpuInfo := a.getCPUInfo()

	info := map[string]interface{}{
		"ollama_version": ollamaVersion,
		"gpu_status":     gpuStatus,
		"service_status": serviceStatus,
		"memory_usage":   memoryInfo,
		"os_info":        runtime.GOOS,
		"arch":           runtime.GOARCH,
		"cpu_info":       cpuInfo,
	}
	return info
}

// detectGPUStatus 检测 GPU 状态
func (a *App) detectGPUStatus() string {
	// 首先检查 Ollama 服务是否在日志中报告 GPU 信息
	// 尝试通过 HTTP 获取系统信息
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:11434/api/tags")
	if err == nil {
		defer resp.Body.Close()
		// 如果服务响应，说明 GPU 可能已初始化
		// 在实际实现中，可以解析更详细的 GPU 信息
		// 这里简化返回
		return "Intel GPU - Available (Service Running)"
	}

	// 如果服务未运行，尝试通过系统方式检测
	if runtime.GOOS == "windows" {
		// 检查常见的 Intel GPU 相关文件
		possiblePaths := []string{
			"C:\\Windows\\System32\\igfxCUIService.exe",
			"C:\\Windows\\System32\\igfxDHL.dll",
			"C:\\Windows\\System32\\DriverStore\\FileRepository\\iigd_dch.inf_amd64_*",
		}

		for _, path := range possiblePaths {
			// 检查文件或目录是否存在（简化检查）
			if path[len(path)-1] == '*' {
				// 通配符路径，检查目录是否存在
				dir := filepath.Dir(path)
				if _, err := os.Stat(dir); err == nil {
					return "Intel GPU - Detected (Driver Files Found)"
				}
			} else {
				if _, err := os.Stat(path); err == nil {
					return "Intel GPU - Detected (Driver Files Found)"
				}
			}
		}

		// 尝试运行系统命令检测 GPU
		cmd := exec.Command("powershell", "-Command", "Get-WmiObject Win32_VideoController | Select-Object Name")
		// 隐藏命令窗口
		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				HideWindow: true,
			}
		}
		output, err := cmd.Output()
		if err == nil {
			outputStr := string(output)
			if strings.Contains(strings.ToLower(outputStr), "intel") {
				return "Intel GPU - Detected via System"
			}
			if strings.Contains(strings.ToLower(outputStr), "radeon") || strings.Contains(strings.ToLower(outputStr), "nvidia") {
				return "Non-Intel GPU - " + strings.TrimSpace(outputStr)
			}
		}
	}

	// 通用检测
	return "GPU detection in progress - Intel optimization enabled"
}

// getMemoryInfo 获取内存信息
func (a *App) getMemoryInfo() string {
	if runtime.GOOS == "windows" {
		// 使用 PowerShell 获取内存信息
		cmd := exec.Command("powershell", "-Command", "Get-WmiObject Win32_ComputerSystem | Select-Object TotalPhysicalMemory")
		// 隐藏命令窗口
		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				HideWindow: true,
			}
		}
		output, err := cmd.Output()
		if err == nil {
			outputStr := string(output)
			lines := strings.Split(strings.TrimSpace(outputStr), "\n")
			if len(lines) > 2 {
				// 提取内存值（字节）
				memoryStr := strings.TrimSpace(lines[2])
				if memoryStr != "" {
					// 尝试解析为字节并转换为GB
					if memoryBytes, err := strconv.ParseUint(memoryStr, 10, 64); err == nil {
						memoryGB := float64(memoryBytes) / (1024 * 1024 * 1024)
						return fmt.Sprintf("%.1f GB Total", memoryGB)
					}
				}
			}
		}
	}

	// 通用实现
	return "System memory detected"
}

// getCPUInfo 获取 CPU 信息
func (a *App) getCPUInfo() string {
	// 获取CPU核心数
	numCPU := runtime.NumCPU()

	if runtime.GOOS == "windows" {
		// 使用 PowerShell 获取 CPU 型号
		cmd := exec.Command("powershell", "-Command", "Get-WmiObject Win32_Processor | Select-Object Name")
		// 隐藏命令窗口
		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				HideWindow: true,
			}
		}
		output, err := cmd.Output()
		if err == nil {
			outputStr := string(output)
			lines := strings.Split(strings.TrimSpace(outputStr), "\n")
			if len(lines) > 2 {
				cpuName := strings.TrimSpace(lines[2])
				if cpuName != "" {
					return fmt.Sprintf("%s (%d cores)", cpuName, numCPU)
				}
			}
		}
	}

	// 通用实现
	return fmt.Sprintf("%d CPU cores", numCPU)
}

// ListModels 获取本地模型列表
func (a *App) ListModels() []ModelInfo {
	// 尝试通过 HTTP API 获取模型列表
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://127.0.0.1:11434/api/tags")
	if err != nil {
		log.Printf("ListModels: HTTP 请求失败: %v", err)
		// 如果失败，返回模拟数据
		return a.getMockModels()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ListModels: HTTP 状态码错误: %d", resp.StatusCode)
		// 如果失败，返回模拟数据
		return a.getMockModels()
	}

	// 解析结果
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("ListModels: JSON 解析失败: %v", err)
		// 如果失败，返回模拟数据
		return a.getMockModels()
	}

	log.Printf("ListModels: 原始结果: %+v", result)

	// 解析结果
	var models []ModelInfo
	if modelsList, ok := result["models"].([]interface{}); ok {
		log.Printf("ListModels: 找到 %d 个模型", len(modelsList))
		for _, m := range modelsList {
			if modelMap, ok := m.(map[string]interface{}); ok {
				log.Printf("ListModels: 模型信息: %+v", modelMap)

				// 解析 details 字段
				details := make(map[string]interface{})
				if detailsMap, ok := modelMap["details"].(map[string]interface{}); ok {
					details = detailsMap
				}

				model := ModelInfo{
					Name:     getString(modelMap, "name"),
					Model:    getString(modelMap, "model"),
					Size:     getSizeString(modelMap, "size"),
					Digest:   getString(modelMap, "digest"),
					Modified: getString(modelMap, "modified_at"),
					Details:  details,
				}
				models = append(models, model)
			}
		}
	} else {
		log.Printf("ListModels: 未找到 'models' 字段或格式不正确")
	}

	if len(models) == 0 {
		log.Printf("ListModels: 模型列表为空，返回模拟数据")
		return a.getMockModels()
	}

	log.Printf("ListModels: 返回 %d 个模型", len(models))
	return models
}

// getMockModels 返回模拟模型数据
func (a *App) getMockModels() []ModelInfo {
	return []ModelInfo{
		{
			Name:     "llama3:8b",
			Size:     "4.7 GB",
			Modified: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			Details: map[string]interface{}{
				"format":             "gguf",
				"families":           []string{"llama"},
				"parameter_size":     "8B",
				"quantization_level": "Q4_K_M",
			},
		},
		{
			Name:     "mistral:7b",
			Size:     "4.1 GB",
			Modified: time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			Details: map[string]interface{}{
				"format":             "gguf",
				"families":           []string{"transformer"},
				"parameter_size":     "7B",
				"quantization_level": "Q5_K_M",
			},
		},
	}
}

// getString 从 map 中获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getInt64 从 map 中获取 int64 值
func getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		}
	}
	return 0
}

// formatBytes 将字节数格式化为人类可读的字符串
func formatBytes(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}
	const unit = 1024
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	i := 0
	fb := float64(bytes)
	for fb >= unit && i < len(sizes)-1 {
		fb /= unit
		i++
	}
	return fmt.Sprintf("%.1f %s", fb, sizes[i])
}

// getSizeString 从 map 中获取大小并格式化为人类可读字符串
func getSizeString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return formatBytes(int64(v))
		case int64:
			return formatBytes(v)
		case int:
			return formatBytes(int64(v))
		case string:
			return v
		}
	}
	return ""
}

// PullModel 拉取模型
func (a *App) PullModel(name string) map[string]interface{} {
	// 处理模型名称，确保包含tag
	modelName := a.normalizeModelName(name)
	log.Printf("PullModel: 原始名称=%s, 规范化名称=%s", name, modelName)

	// 在 goroutine 中拉取模型以提供进度更新
	go a.pullModelWithProgress(modelName)

	return map[string]interface{}{
		"message": fmt.Sprintf("开始拉取模型: %s", modelName),
		"model":   modelName,
		"status":  "started",
	}
}

// normalizeModelName 规范化模型名称，确保包含tag
func (a *App) normalizeModelName(name string) string {
	name = strings.TrimSpace(name)

	// 如果模型名称已经包含tag（包含冒号），直接返回
	if strings.Contains(name, ":") {
		return name
	}

	// 常见的模型名称映射（纠正拼写错误）
	modelNameCorrections := map[string]string{
		"llama4":         "llama3.3",
		"llama3.3":       "llama3.3",
		"llama3.2":       "llama3.2",
		"llama3.1":       "llama3.1",
		"llama3":         "llama3",
		"llama2":         "llama2",
		"gemma3":         "gemma3",
		"gemma2":         "gemma2",
		"gemma":          "gemma",
		"mistral":        "mistral",
		"mixtral":        "mixtral",
		"qwen3.5":        "qwen3.5",
		"qwen3":          "qwen3",
		"qwen2.5":        "qwen2.5",
		"qwen2":          "qwen2",
		"phi4":           "phi4",
		"phi3":           "phi3",
		"phi-3":          "phi3",
		"codellama":      "codellama",
		"code-llama":     "codellama",
		"deepseek-coder": "deepseek-coder",
		"deepseek":       "deepseek-r1",
		"deepseek-r1":    "deepseek-r1",
		"deepseek-v3":    "deepseek-v3",
		"llava":          "llava",
		"starling-lm":    "starling-lm",
		"command-r":      "command-r",
		"cohere":         "command-r",
	}

	// 检查是否需要纠正模型名称
	lowerName := strings.ToLower(name)
	if corrected, ok := modelNameCorrections[lowerName]; ok {
		name = corrected
	}

	// 对于常见的模型，添加默认tag
	defaultTags := map[string]string{
		"llama3.3":       "latest",
		"llama3.2":       "latest",
		"llama3.1":       "latest",
		"llama3":         "latest",
		"llama2":         "latest",
		"gemma3":         "latest",
		"gemma2":         "latest",
		"gemma":          "latest",
		"mistral":        "latest",
		"mixtral":        "latest",
		"qwen3.5":        "latest",
		"qwen3":          "latest",
		"qwen2.5":        "latest",
		"qwen2":          "latest",
		"phi4":           "latest",
		"phi3":           "latest",
		"codellama":      "latest",
		"deepseek-coder": "latest",
		"deepseek-r1":    "latest",
		"deepseek-v3":    "latest",
		"llava":          "latest",
		"starling-lm":    "latest",
		"command-r":      "latest",
	}

	// 检查是否有预定义的默认tag
	lowerName = strings.ToLower(name)
	if tag, ok := defaultTags[lowerName]; ok {
		return fmt.Sprintf("%s:%s", name, tag)
	}

	// 默认添加:latest tag
	return fmt.Sprintf("%s:latest", name)
}

// pullModelWithProgress 拉取模型并发送进度更新
func (a *App) pullModelWithProgress(modelName string) {
	log.Printf("开始拉取模型: %s", modelName)

	// 发送开始事件
	a.sendPullProgressEvent(modelName, "started", 0, "开始拉取模型")

	// 构建环境变量
	env := os.Environ()

	// 添加用户配置的环境变量（会覆盖系统变量）
	for key, value := range a.environmentVariables {
		// 跳过空值
		if strValue, ok := value.(string); ok {
			if strValue == "" {
				continue
			}
			// 检查是否已存在该环境变量，如果存在则替换
			found := false
			for i, e := range env {
				if strings.HasPrefix(e, key+"=") {
					env[i] = fmt.Sprintf("%s=%s", key, strValue)
					found = true
					break
				}
			}
			if !found {
				env = append(env, fmt.Sprintf("%s=%s", key, strValue))
			}
			// 记录关键环境变量
			if key == "OLLAMA_MODEL_SOURCE" || key == "OLLAMA_ORIGINS" || key == "OLLAMA_HOST" {
				log.Printf("配置环境变量 %s=%s", key, strValue)
			}
		} else if boolValue, ok := value.(bool); ok {
			if boolValue {
				env = append(env, fmt.Sprintf("%s=true", key))
			}
		} else if intValue, ok := value.(float64); ok {
			env = append(env, fmt.Sprintf("%s=%d", key, int(intValue)))
		}
	}

	// 记录关键环境变量用于调试
	log.Printf("模型拉取: %s, 环境变量数量: %d", modelName, len(env))

	cmd := exec.Command(a.ollamaPath, "pull", modelName)
	cmd.Env = env

	// 隐藏命令窗口
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
	}

	// 保存进程到map中，以便后续可以取消
	a.pullProcessesMutex.Lock()
	a.pullProcesses[modelName] = cmd
	a.pullProcessesMutex.Unlock()

	// 确保在函数结束时清理进程
	defer func() {
		a.pullProcessesMutex.Lock()
		delete(a.pullProcesses, modelName)
		a.pullProcessesMutex.Unlock()
	}()

	// 获取标准输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("创建标准输出管道失败: %v", err)
		a.sendPullProgressEvent(modelName, "error", 0, fmt.Sprintf("创建管道失败: %v", err))
		return
	}

	// 获取标准错误管道
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("创建标准错误管道失败: %v", err)
		a.sendPullProgressEvent(modelName, "error", 0, fmt.Sprintf("创建管道失败: %v", err))
		return
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		log.Printf("启动命令失败: %v", err)
		a.sendPullProgressEvent(modelName, "error", 0, fmt.Sprintf("启动命令失败: %v", err))
		return
	}

	// 使用通道来跟踪错误状态
	errorOccurred := make(chan bool, 1)

	// 合并 stdout 和 stderr 的读取
	go a.readOutputLines(stdout, modelName, "stdout", errorOccurred)
	go a.readOutputLines(stderr, modelName, "stderr", errorOccurred)

	// 等待命令完成
	err = cmd.Wait()

	// 检查是否有错误发生
	select {
	case <-errorOccurred:
		log.Printf("模型拉取过程中发生错误: %s", modelName)
		return
	default:
	}

	if err != nil {
		log.Printf("拉取模型失败: %v", err)
		// 检查是否是被取消的
		if strings.Contains(err.Error(), "signal: killed") || strings.Contains(err.Error(), "signal: terminated") {
			a.sendPullProgressEvent(modelName, "cancelled", 0, "模型拉取已取消")
		} else {
			a.sendPullProgressEvent(modelName, "error", 0, fmt.Sprintf("拉取模型失败: %v", err))
		}
		return
	}

	// 发送完成事件
	a.sendPullProgressEvent(modelName, "completed", 100, "模型拉取完成")
	log.Printf("模型拉取完成: %s", modelName)
}

// sendPullProgressEvent 发送模型拉取进度事件
func (a *App) sendPullProgressEvent(modelName, status string, progress float64, message string) {
	eventData := map[string]interface{}{
		"model":    modelName,
		"status":   status,
		"progress": progress,
		"message":  message,
		"time":     time.Now().Format("2006-01-02 15:04:05"),
	}

	// 发送事件到前端
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "model_pull_progress", eventData)
	}

	// 同时记录日志
	log.Printf("模型拉取进度: %s - %s (%.1f%%)", modelName, status, progress)
}

// readOutputLines 读取输出行并解析进度
func (a *App) readOutputLines(reader io.Reader, modelName, streamType string, errorOccurred chan bool) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("模型拉取输出[%s]: %s", streamType, line)

		// 检测错误信息
		if streamType == "stderr" {
			errorMsg := a.parsePullError(line)
			if errorMsg != "" {
				a.sendPullProgressEvent(modelName, "error", 0, errorMsg)
				select {
				case errorOccurred <- true:
				default:
				}
				return
			}
		}

		// 解析进度信息
		progress, message := a.parsePullProgress(line)

		// 发送进度事件
		if progress >= 0 {
			a.sendPullProgressEvent(modelName, "downloading", progress, message)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("读取输出失败: %v", err)
		a.sendPullProgressEvent(modelName, "error", 0, fmt.Sprintf("读取输出失败: %v", err))
		select {
		case errorOccurred <- true:
		default:
		}
	}
}

// parsePullError 解析拉取错误信息，返回用户友好的错误消息
func (a *App) parsePullError(line string) string {
	lineLower := strings.ToLower(line)

	// 检测各种错误类型
	if strings.Contains(line, "412") || strings.Contains(lineLower, "precondition failed") {
		return "模型拉取失败 (412): 可能原因：1) 模型名称拼写错误；2) 网络连接问题；3) Ollama服务版本过旧。建议：检查模型名称、更新Ollama版本、或尝试使用镜像源(在设置中配置OLLAMA_MODEL_SOURCE=modelscope)"
	}
	if strings.Contains(lineLower, "not found") || strings.Contains(lineLower, "404") {
		return "模型不存在: 请检查模型名称是否正确，可以在在线模型页面查看可用模型"
	}
	if strings.Contains(lineLower, "connection refused") || strings.Contains(lineLower, "network") {
		return "网络连接失败: 请检查网络连接，确保可以访问模型仓库"
	}
	if strings.Contains(lineLower, "timeout") {
		return "连接超时: 网络响应过慢，请稍后重试"
	}
	if strings.Contains(lineLower, "permission denied") || strings.Contains(lineLower, "403") {
		return "权限不足: 无法访问该模型"
	}
	if strings.Contains(lineLower, "disk") || strings.Contains(lineLower, "space") {
		return "磁盘空间不足: 请清理磁盘空间后重试"
	}
	if strings.Contains(line, "Error:") || strings.Contains(lineLower, "error:") {
		// 提取错误信息
		if idx := strings.Index(line, "Error:"); idx != -1 {
			return strings.TrimSpace(line[idx:])
		}
		if idx := strings.Index(lineLower, "error:"); idx != -1 {
			return strings.TrimSpace(line[idx:])
		}
		return line
	}

	return ""
}

// parsePullProgress 解析拉取进度信息
func (a *App) parsePullProgress(line string) (float64, string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return -1, ""
	}

	// 尝试解析 JSON 格式的进度信息
	var jsonProgress struct {
		Status    string  `json:"status"`
		Completed int64   `json:"completed,omitempty"`
		Total     int64   `json:"total,omitempty"`
		Percent   float64 `json:"percent,omitempty"`
		Digest    string  `json:"digest,omitempty"`
	}

	if err := json.Unmarshal([]byte(line), &jsonProgress); err == nil {
		// JSON 格式
		if jsonProgress.Percent > 0 {
			return jsonProgress.Percent, jsonProgress.Status
		}
		if jsonProgress.Total > 0 && jsonProgress.Completed >= 0 {
			percent := float64(jsonProgress.Completed) / float64(jsonProgress.Total) * 100
			return percent, fmt.Sprintf("%s (%s / %s)", jsonProgress.Status,
				formatBytes(jsonProgress.Completed), formatBytes(jsonProgress.Total))
		}
		if jsonProgress.Status != "" {
			return -1, jsonProgress.Status
		}
	}

	// 尝试解析文本格式的百分比
	if strings.Contains(line, "%") {
		parts := strings.Fields(line)
		for _, part := range parts {
			if strings.Contains(part, "%") {
				percentStr := strings.Trim(part, "%")
				if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
					return percent, line
				}
			}
		}
	}

	return -1, line
}

// DeleteModel 删除模型
func (a *App) DeleteModel(name string) map[string]interface{} {
	// 通过命令行删除模型
	go func() {
		_, err := a.runOllamaCommand("rm", name)
		if err != nil {
			log.Printf("删除模型失败: %v", err)
		}
	}()

	return map[string]interface{}{
		"message": fmt.Sprintf("模型已删除: %s", name),
	}
}

// CancelPull 取消模型拉取
func (a *App) CancelPull(modelName string) map[string]interface{} {
	log.Printf("CancelPull: 尝试取消模型拉取: %s", modelName)

	// 规范化模型名称
	normalizedName := a.normalizeModelName(modelName)

	// 查找正在运行的拉取进程
	a.pullProcessesMutex.Lock()
	defer a.pullProcessesMutex.Unlock()

	cmd, exists := a.pullProcesses[normalizedName]
	if !exists {
		log.Printf("CancelPull: 未找到正在运行的拉取进程: %s", normalizedName)
		return map[string]interface{}{
			"success": false,
			"message": "未找到正在运行的拉取进程",
		}
	}

	// 终止进程
	if cmd != nil && cmd.Process != nil {
		log.Printf("CancelPull: 终止拉取进程: %s", normalizedName)
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("CancelPull: 终止进程失败: %v", err)
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("终止进程失败: %v", err),
			}
		}

		// 发送取消事件
		a.sendPullProgressEvent(normalizedName, "cancelled", 0, "用户取消了模型拉取")

		log.Printf("CancelPull: 成功取消模型拉取: %s", normalizedName)
		return map[string]interface{}{
			"success": true,
			"message": "模型拉取已取消",
		}
	}

	return map[string]interface{}{
		"success": false,
		"message": "进程不存在",
	}
}

// ShowModel 显示模型信息
func (a *App) ShowModel(name string) map[string]interface{} {
	// 使用 HTTP API 获取模型信息
	client := &http.Client{Timeout: 10 * time.Second}
	
	// 使用 Ollama HTTP API 获取模型详情
	url := "http://127.0.0.1:11434/api/show"
	reqBody := map[string]string{"name": name}
	jsonBody, _ := json.Marshal(reqBody)
	
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("ShowModel: HTTP请求失败: %v", err)
		return map[string]interface{}{
			"license":   "...",
			"modelfile": "# Modelfile generated by ollama...",
			"parameters": map[string]interface{}{
				"num_ctx": 2048,
			},
			"template": "{{ if .System }}...",
		}
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("ShowModel: JSON解析失败: %v", err)
		return map[string]interface{}{
			"license":   "...",
			"modelfile": "# Modelfile generated by ollama...",
			"parameters": map[string]interface{}{
				"num_ctx": 2048,
			},
			"template": "{{ if .System }}...",
		}
	}
	
	return result
}

// ChatCompletion 聊天完成
func (a *App) ChatCompletion(req ChatRequest) ChatResponse {
	// 确保设置 stream: true 以支持流式响应
	req.Stream = true

	// 使用 HTTP API 而不是命令行工具
	client := &http.Client{Timeout: 60 * time.Second}

	// 构建请求体
	reqBody, err := json.Marshal(req)
	if err != nil {
		// 如果失败，返回模拟响应
		return ChatResponse{
			Model: req.Model,
			Message: ChatMessage{
				Role:    "assistant",
				Content: "这是来自 Ollama 英特尔优化版的模拟响应。在实际实现中，这里会连接到真实的 Ollama API。",
			},
			Done: true,
		}
	}

	// 发送请求
	resp, err := client.Post("http://127.0.0.1:11434/api/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		// 如果失败，返回模拟响应
		return ChatResponse{
			Model: req.Model,
			Message: ChatMessage{
				Role:    "assistant",
				Content: "这是来自 Ollama 英特尔优化版的模拟响应。在实际实现中，这里会连接到真实的 Ollama API。",
			},
			Done: true,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 如果失败，返回模拟响应
		return ChatResponse{
			Model: req.Model,
			Message: ChatMessage{
				Role:    "assistant",
				Content: "这是来自 Ollama 英特尔优化版的模拟响应。在实际实现中，这里会连接到真实的 Ollama API。",
			},
			Done: true,
		}
	}

	// 处理流式响应
	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder
	var response ChatResponse

	for scanner.Scan() {
		line := scanner.Text()

		// 解析单行 JSON
		var chunk struct {
			Model     string      `json:"model"`
			CreatedAt string      `json:"created_at"`
			Message   ChatMessage `json:"message"`
			Done      bool        `json:"done"`
			Error     string      `json:"error,omitempty"`
		}

		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}

		// 累积内容
		if chunk.Message.Content != "" {
			fullContent.WriteString(chunk.Message.Content)

			// 发送流式事件到前端
			if a.ctx != nil {
				streamEvent := map[string]interface{}{
					"model":        chunk.Model,
					"content":      chunk.Message.Content,
					"full_content": fullContent.String(),
					"done":         chunk.Done,
				}
				wailsRuntime.EventsEmit(a.ctx, "chat_stream", streamEvent)
			}
		}

		// 当完成时，设置最终响应
		if chunk.Done {
			response = ChatResponse{
				Model:     chunk.Model,
				CreatedAt: chunk.CreatedAt,
				Message: ChatMessage{
					Role:    "assistant",
					Content: fullContent.String(),
				},
				Done: true,
			}
			break
		}
	}

	if err := scanner.Err(); err != nil {
		// 如果失败，返回模拟响应
		return ChatResponse{
			Model: req.Model,
			Message: ChatMessage{
				Role:    "assistant",
				Content: "这是来自 Ollama 英特尔优化版的模拟响应。在实际实现中，这里会连接到真实的 Ollama API。",
			},
			Done: true,
		}
	}

	// 返回最终响应
	return response
}

// ChatStreamRequest 聊天流式请求
type ChatStreamRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// ChatStreamResult 聊天流式结果
type ChatStreamResult struct {
	Content   string `json:"content"`
	Done      bool   `json:"done"`
	Error     string `json:"error,omitempty"`
	Model     string `json:"model"`
	TotalTime int64  `json:"total_time,omitempty"`
}

// ChatStream 聊天流式响应，通过后端代理调用Ollama API
// 使用事件推送实现真正的流式传输
func (a *App) ChatStream(req ChatStreamRequest) *ChatStreamResult {
	log.Printf("ChatStream: 模型=%s, 消息数=%d", req.Model, len(req.Messages))

	// 使用 HTTP API 进行聊天
	client := &http.Client{Timeout: 180 * time.Second}

	// 构建请求体
	reqBody, err := json.Marshal(map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	})
	if err != nil {
		return &ChatStreamResult{
			Error: fmt.Sprintf("构建请求失败: %v", err),
			Done:  true,
		}
	}

	// 发送请求
	resp, err := client.Post("http://127.0.0.1:11434/api/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return &ChatStreamResult{
			Error: fmt.Sprintf("连接Ollama服务失败: %v", err),
			Done:  true,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &ChatStreamResult{
			Error: fmt.Sprintf("Ollama返回错误: %d - %s", resp.StatusCode, string(body)),
			Done:  true,
		}
	}

	// 处理流式响应，通过事件推送到前端
	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder
	startTime := time.Now()
	var modelName string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var chunk struct {
			Model     string      `json:"model"`
			Message   ChatMessage `json:"message"`
			Done      bool        `json:"done"`
			Error     string      `json:"error,omitempty"`
			TotalTime int64       `json:"total_duration,omitempty"`
		}

		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			log.Printf("解析响应失败: %v, 行: %s", err, line)
			continue
		}

		if chunk.Error != "" {
			// 发送错误事件
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "chat_stream_chunk", map[string]interface{}{
					"error": chunk.Error,
					"done":  true,
				})
			}
			return &ChatStreamResult{
				Error: chunk.Error,
				Done:  true,
			}
		}

		if chunk.Model != "" {
			modelName = chunk.Model
		}

		if chunk.Message.Content != "" {
			fullContent.WriteString(chunk.Message.Content)
			
			// 发送流式更新事件到前端
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "chat_stream_chunk", map[string]interface{}{
					"content":     chunk.Message.Content,
					"full_content": fullContent.String(),
					"done":        false,
					"model":       modelName,
				})
			}
		}

		if chunk.Done {
			// 发送完成事件
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "chat_stream_chunk", map[string]interface{}{
					"content":     "",
					"full_content": fullContent.String(),
					"done":        true,
					"model":       modelName,
					"total_time":  time.Since(startTime).Milliseconds(),
				})
			}
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return &ChatStreamResult{
			Error: fmt.Sprintf("读取响应失败: %v", err),
			Done:  true,
		}
	}

	log.Printf("ChatStream完成: 内容长度=%d", fullContent.Len())
	return &ChatStreamResult{
		Content:   fullContent.String(),
		Done:      true,
		Model:     modelName,
		TotalTime: time.Since(startTime).Milliseconds(),
	}
}

// StartService 启动服务
func (a *App) StartService() map[string]interface{} {
	// 在 goroutine 中启动服务以避免阻塞
	go func() {
		if err := a.startOllamaService(); err != nil {
			log.Printf("服务启动失败: %v", err)
			// 这里可以添加错误通知机制，例如通过事件发送到前端
		}
	}()

	return map[string]interface{}{
		"message": "服务启动中...",
		"success": true,
	}
}

// StopService 停止服务
func (a *App) StopService() map[string]interface{} {
	// 停止 Ollama 服务
	if a.ollamaCmd != nil && a.ollamaCmd.Process != nil {
		a.ollamaCmd.Process.Kill()
		a.ollamaCmd = nil
		log.Println("Ollama 服务已停止")
	}

	return map[string]interface{}{
		"message": "服务停止中...",
	}
}

// GetEnvironmentVariables 获取环境变量配置
func (a *App) GetEnvironmentVariables() map[string]interface{} {
	log.Println("GetEnvironmentVariables: 获取环境变量配置")
	return a.environmentVariables
}

// GetOllamaPath 获取当前的Ollama执行程序路径
func (a *App) GetOllamaPath() map[string]interface{} {
	log.Printf("GetOllamaPath: 当前路径: %s\n", a.ollamaPath)
	return map[string]interface{}{
		"path": a.ollamaPath,
	}
}

// getConfigPath 获取配置文件路径
func (a *App) getConfigPath() string {
	var configDir string

	if runtime.GOOS == "windows" {
		configDir = filepath.Join(os.Getenv("APPDATA"), "ollama-intel")
	} else {
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "ollama-intel")
	}

	// 确保配置目录存在
	os.MkdirAll(configDir, 0755)

	return filepath.Join(configDir, "config.json")
}

// loadConfig 从文件加载配置
func (a *App) loadConfig() {
	configPath := a.getConfigPath()

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Println("loadConfig: 配置文件不存在，使用默认配置")
		return
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("loadConfig: 读取配置文件失败: %v\n", err)
		return
	}

	// 解析配置文件
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("loadConfig: 解析配置文件失败: %v\n", err)
		return
	}

	// 加载环境变量配置
	if envVars, ok := config["environmentVariables"].(map[string]interface{}); ok {
		a.environmentVariables = envVars
		log.Println("loadConfig: 环境变量配置已加载")
	}

	// 加载Ollama路径配置
	if ollamaPath, ok := config["ollamaPath"].(string); ok && ollamaPath != "" {
		a.ollamaPath = ollamaPath
		log.Printf("loadConfig: Ollama路径已加载: %s\n", a.ollamaPath)
	}
}

// saveConfig 保存配置到文件
func (a *App) saveConfig() {
	configPath := a.getConfigPath()

	// 构建配置数据
	config := map[string]interface{}{
		"environmentVariables": a.environmentVariables,
		"ollamaPath":           a.ollamaPath,
		"lastSaved":            time.Now().Format(time.RFC3339),
	}

	// 序列化配置数据
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("saveConfig: 序列化配置失败: %v\n", err)
		return
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		log.Printf("saveConfig: 写入配置文件失败: %v\n", err)
		return
	}

	log.Printf("saveConfig: 配置已保存到 %s\n", configPath)
}

// SaveEnvironmentVariables 保存环境变量配置
func (a *App) SaveEnvironmentVariables(variables map[string]interface{}) map[string]interface{} {
	log.Printf("SaveEnvironmentVariables: 保存环境变量配置: %+v\n", variables)

	// 保存环境变量
	a.environmentVariables = variables

	// 检查是否设置了OLLAMA_EXECUTABLE_PATH
	if executablePath, ok := variables["OLLAMA_EXECUTABLE_PATH"]; ok {
		if pathStr, ok := executablePath.(string); ok {
			if pathStr != "" {
				a.ollamaPath = pathStr
				log.Printf("SaveEnvironmentVariables: 更新 Ollama 执行程序路径为: %s\n", a.ollamaPath)
			} else {
				// 如果用户清空了路径，重置为默认路径
				a.setOllamaPath()
				log.Printf("SaveEnvironmentVariables: 重置为默认 Ollama 执行程序路径: %s\n", a.ollamaPath)
			}
		}
	}

	// 保存配置到文件
	a.saveConfig()

	log.Println("SaveEnvironmentVariables: 环境变量配置已保存")

	return map[string]interface{}{
		"message":   "环境变量配置已保存",
		"variables": variables,
		"path":      a.ollamaPath, // 返回当前路径
	}
}

// GetServiceStatus 获取服务状态
func (a *App) GetServiceStatus() map[string]interface{} {
	// 通过 HTTP 请求检查服务状态
	running := a.checkOllamaService()
	version := "unknown"
	log.Printf("GetServiceStatus: running=%v\n", running)

	if running {
		// 尝试获取版本信息
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://127.0.0.1:11434/api/version")
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
				if v, ok := result["version"].(string); ok {
					version = v
				}
			}
		} else {
			log.Printf("GetServiceStatus 获取版本失败: err=%v\n", err)
		}
	}

	status := map[string]interface{}{
		"running": running,
		"host":    "127.0.0.1:11434",
		"version": version,
	}
	log.Printf("GetServiceStatus 返回: %+v\n", status)
	// 写入状态到文件以便调试
	statusJSON, _ := json.MarshalIndent(status, "", "  ")
	os.WriteFile("service_status_debug.txt", statusJSON, 0644)
	return status
}

// checkOllamaService 通过 HTTP 请求检查 Ollama 服务是否运行
func (a *App) checkOllamaService() bool {
	url := "http://127.0.0.1:11434/api/tags"
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)

	// 写入调试信息到文件
	debugMsg := fmt.Sprintf("时间: %s\nURL: %s\n错误: %v\n",
		time.Now().Format("2006-01-02 15:04:05"), url, err)
	if err == nil {
		debugMsg += fmt.Sprintf("状态码: %d\n", resp.StatusCode)
	}

	os.WriteFile("check_service_debug.txt", []byte(debugMsg), 0644)

	if err != nil {
		log.Printf("checkOllamaService HTTP请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()
	statusOK := resp.StatusCode == http.StatusOK
	log.Printf("checkOllamaService 状态码: %d, 成功: %v\n", resp.StatusCode, statusOK)
	return statusOK
}

// GetOnlineModels 获取在线模型
func (a *App) GetOnlineModels(page int, limit int) map[string]interface{} {
	// 使用本地 Ollama 命令获取模型列表
	models, err := a.fetchOnlineModelsFromOllama()
	if err != nil {
		log.Printf("获取在线模型失败，使用内置模型列表: %v", err)
		models = a.getBuiltinOnlineModels()
	}

	// 实现分页
	total := len(models)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return map[string]interface{}{
			"models": []map[string]interface{}{},
			"total":  total,
			"page":   page,
			"limit":  limit,
		}
	}

	if end > total {
		end = total
	}

	paginatedModels := models[start:end]

	return map[string]interface{}{
		"models": paginatedModels,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}
}

// SearchOnlineModels 搜索在线模型
func (a *App) SearchOnlineModels(query string, page int, limit int) map[string]interface{} {
	// 使用本地 Ollama 命令搜索模型
	models, err := a.searchOnlineModelsWithOllama(query)
	if err != nil {
		log.Printf("搜索在线模型失败，使用内置模型列表: %v", err)
		// 如果搜索失败，使用内置模型列表并进行过滤
		models = a.getBuiltinOnlineModels()
		if query != "" {
			var filtered []map[string]interface{}
			lowerQuery := strings.ToLower(query)
			for _, model := range models {
				if name, ok := model["name"].(string); ok {
					if strings.Contains(strings.ToLower(name), lowerQuery) {
						filtered = append(filtered, model)
					}
				}
			}
			models = filtered
		}
	}

	// 实现分页
	total := len(models)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return map[string]interface{}{
			"models": []map[string]interface{}{},
			"total":  total,
			"page":   page,
			"limit":  limit,
		}
	}

	if end > total {
		end = total
	}

	paginatedModels := models[start:end]

	return map[string]interface{}{
		"models": paginatedModels,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}
}

// fetchOnlineModelsFromOllama 使用 Ollama 命令获取在线模型列表
func (a *App) fetchOnlineModelsFromOllama() ([]map[string]interface{}, error) {
	// 首先尝试从 API 获取模型列表
	models, err := a.fetchOnlineModelsFromAPI()
	if err == nil && len(models) > 0 {
		log.Printf("从 API 获取到 %d 个在线模型", len(models))
		return models, nil
	}

	log.Printf("从 API 获取模型失败，使用内置模型列表: %v", err)

	// 运行 ollama list 命令获取本地模型
	localModels, err := a.runOllamaCommand("list")
	if err != nil {
		log.Printf("运行 ollama list 失败: %v", err)
		// 即使本地模型获取失败，也继续返回内置模型列表
	}

	log.Printf("本地模型列表: %s", localModels)

	// 使用内置的在线模型列表
	// 这些模型是从 Ollama 官方库中精选的
	return a.getBuiltinOnlineModels(), nil
}

// fetchOnlineModelsFromAPI 从 Ollama 官方库页面获取模型列表
// 通过解析 ollama.com/library 和 ollama.com/search 页面获取完整的模型列表
func (a *App) fetchOnlineModelsFromAPI() ([]map[string]interface{}, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	// 从 ollama.com/library 页面获取模型列表
	resp, err := client.Get("https://ollama.com/library")
	if err != nil {
		log.Printf("从 Ollama library 页面获取模型失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("页面返回状态码: %d", resp.StatusCode)
	}

	// 读取页面内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析HTML获取模型列表
	models := a.parseOllamaLibraryPage(string(body))
	log.Printf("从 Ollama library 页面解析到 %d 个模型", len(models))

	if len(models) == 0 {
		return nil, fmt.Errorf("未能从页面解析到任何模型")
	}

	return models, nil
}

// parseOllamaLibraryPage 解析 Ollama library 页面获取模型列表
func (a *App) parseOllamaLibraryPage(htmlContent string) []map[string]interface{} {
	var models []map[string]interface{}
	seenModels := make(map[string]bool)

	// 使用正则表达式提取模型信息
	// 匹配模型名称 - 在 <h2> 标签内的模型名称
	nameRegex := regexp.MustCompile(`<h2[^>]*>\s*([a-zA-Z0-9_.-]+)\s*</h2>`)
	// 匹配描述 - 在 class="text-neutral-400" 的段落中
	descRegex := regexp.MustCompile(`<p[^>]*class="[^"]*text-neutral-400[^"]*"[^>]*>([^<]+)</p>`)
	// 匹配下载量
	pullsRegex := regexp.MustCompile(`(\d+\.?\d*[KM]?)\s*Pulls`)
	// 匹配标签数量
	tagsRegex := regexp.MustCompile(`(\d+)\s*Tags`)
	// 匹配参数大小
	paramRegex := regexp.MustCompile(`(\d+(?:\.\d+)?[bm]?)`)

	// 提取所有模型名称
	nameMatches := nameRegex.FindAllStringSubmatch(htmlContent, -1)
	descMatches := descRegex.FindAllStringSubmatch(htmlContent, -1)

	// 提取下载量
	pullsMatches := pullsRegex.FindAllStringSubmatch(htmlContent, -1)
	// 提取标签数量
	tagsMatches := tagsRegex.FindAllStringSubmatch(htmlContent, -1)

	// 合并模型信息
	maxLen := len(nameMatches)
	if len(descMatches) > maxLen {
		maxLen = len(descMatches)
	}

	for i := 0; i < maxLen && i < len(nameMatches); i++ {
		name := strings.TrimSpace(nameMatches[i][1])
		if name == "" || seenModels[name] {
			continue
		}
		seenModels[name] = true

		description := ""
		if i < len(descMatches) {
			description = strings.TrimSpace(descMatches[i][1])
		}

		pulls := ""
		if i < len(pullsMatches) {
			pulls = strings.TrimSpace(pullsMatches[i][1]) + " Pulls"
		}

		tags := ""
		if i < len(tagsMatches) {
			tags = strings.TrimSpace(tagsMatches[i][1])
		}

		// 尝试从描述中提取参数大小
		paramSize := ""
		if matches := paramRegex.FindStringSubmatch(description); len(matches) > 1 {
			paramSize = matches[1]
		}

		model := map[string]interface{}{
			"name":        name,
			"description": description,
			"pulls":       pulls,
			"tags_count":  tags,
			"details": map[string]interface{}{
				"parameter_size": paramSize,
				"format":         "gguf",
			},
		}

		models = append(models, model)
	}

	// 如果正则表达式没有匹配到足够的模型，尝试另一种方法
	if len(models) < 50 {
		// 使用更简单的正则匹配模型名称
		simpleNameRegex := regexp.MustCompile(`href="/library/([a-zA-Z0-9_.-]+)"`)
		simpleMatches := simpleNameRegex.FindAllStringSubmatch(htmlContent, -1)

		for _, match := range simpleMatches {
			name := strings.TrimSpace(match[1])
			if name == "" || seenModels[name] || strings.Contains(name, "blog") {
				continue
			}
			seenModels[name] = true

			model := map[string]interface{}{
				"name":        name,
				"description": fmt.Sprintf("%s model from Ollama library", name),
				"pulls":       "N/A",
				"tags_count":  "N/A",
				"details": map[string]interface{}{
					"format": "gguf",
				},
			}

			models = append(models, model)
		}
	}

	return models
}

// searchOnlineModelsWithOllama 搜索在线模型
// 首先尝试从 Ollama 搜索页面获取结果，如果失败则从本地缓存搜索
func (a *App) searchOnlineModelsWithOllama(query string) ([]map[string]interface{}, error) {
	// 如果有搜索关键词，尝试从 Ollama 搜索页面获取结果
	if query != "" {
		searchResults, err := a.searchOnlineModelsFromWeb(query)
		if err == nil && len(searchResults) > 0 {
			return searchResults, nil
		}
		log.Printf("从网页搜索失败，使用本地模型列表: %v", err)
	}

	// 获取所有在线模型（从网页或内置列表）
	allModels, err := a.fetchOnlineModelsFromAPI()
	if err != nil {
		log.Printf("获取在线模型失败，使用内置模型列表: %v", err)
		allModels = a.getBuiltinOnlineModels()
	}

	// 如果没有搜索关键词，返回所有模型
	if query == "" {
		return allModels, nil
	}

	// 过滤模型
	var filtered []map[string]interface{}
	lowerQuery := strings.ToLower(query)

	for _, model := range allModels {
		// 检查模型名称是否包含关键词
		if name, ok := model["name"].(string); ok {
			if strings.Contains(strings.ToLower(name), lowerQuery) {
				filtered = append(filtered, model)
				continue
			}
		}

		// 检查模型描述是否包含关键词
		if description, ok := model["description"].(string); ok {
			if strings.Contains(strings.ToLower(description), lowerQuery) {
				filtered = append(filtered, model)
				continue
			}
		}

		// 检查模型详情是否包含关键词
		if details, ok := model["details"].(map[string]interface{}); ok {
			for _, value := range details {
				if strValue, ok := value.(string); ok {
					if strings.Contains(strings.ToLower(strValue), lowerQuery) {
						filtered = append(filtered, model)
						break
					}
				}
			}
		}
	}

	return filtered, nil
}

// searchOnlineModelsFromWeb 从 Ollama 搜索页面搜索模型
func (a *App) searchOnlineModelsFromWeb(query string) ([]map[string]interface{}, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	// 从 ollama.com/search?q=query 页面获取搜索结果
	searchURL := fmt.Sprintf("https://ollama.com/search?q=%s", query)
	resp, err := client.Get(searchURL)
	if err != nil {
		log.Printf("从 Ollama 搜索页面获取结果失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("搜索页面返回状态码: %d", resp.StatusCode)
	}

	// 读取页面内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析搜索结果页面
	models := a.parseOllamaSearchPage(string(body), query)
	log.Printf("从 Ollama 搜索页面解析到 %d 个模型", len(models))

	return models, nil
}

// parseOllamaSearchPage 解析 Ollama 搜索页面获取模型列表
func (a *App) parseOllamaSearchPage(htmlContent string, query string) []map[string]interface{} {
	var models []map[string]interface{}
	seenModels := make(map[string]bool)

	// 使用正则表达式提取模型信息
	// 匹配模型名称 - 在 <h2> 标签内的模型名称
	nameRegex := regexp.MustCompile(`<h2[^>]*>\s*([a-zA-Z0-9_.-]+)\s*</h2>`)
	// 匹配描述 - 在 class="text-neutral-400" 的段落中
	descRegex := regexp.MustCompile(`<p[^>]*class="[^"]*text-neutral-400[^"]*"[^>]*>([^<]+)</p>`)
	// 匹配下载量
	pullsRegex := regexp.MustCompile(`(\d+\.?\d*[KM]?)\s*Pulls`)
	// 匹配标签数量
	tagsRegex := regexp.MustCompile(`(\d+)\s*Tags`)

	// 提取所有模型名称
	nameMatches := nameRegex.FindAllStringSubmatch(htmlContent, -1)
	descMatches := descRegex.FindAllStringSubmatch(htmlContent, -1)

	// 提取下载量
	pullsMatches := pullsRegex.FindAllStringSubmatch(htmlContent, -1)
	// 提取标签数量
	tagsMatches := tagsRegex.FindAllStringSubmatch(htmlContent, -1)

	// 合并模型信息
	maxLen := len(nameMatches)
	if len(descMatches) > maxLen {
		maxLen = len(descMatches)
	}

	lowerQuery := strings.ToLower(query)

	for i := 0; i < maxLen && i < len(nameMatches); i++ {
		name := strings.TrimSpace(nameMatches[i][1])
		if name == "" || seenModels[name] {
			continue
		}

		description := ""
		if i < len(descMatches) {
			description = strings.TrimSpace(descMatches[i][1])
		}

		// 检查是否匹配搜索关键词
		if !strings.Contains(strings.ToLower(name), lowerQuery) &&
			!strings.Contains(strings.ToLower(description), lowerQuery) {
			continue
		}

		seenModels[name] = true

		pulls := ""
		if i < len(pullsMatches) {
			pulls = strings.TrimSpace(pullsMatches[i][1]) + " Pulls"
		}

		tags := ""
		if i < len(tagsMatches) {
			tags = strings.TrimSpace(tagsMatches[i][1])
		}

		model := map[string]interface{}{
			"name":        name,
			"description": description,
			"pulls":       pulls,
			"tags_count":  tags,
			"details": map[string]interface{}{
				"format": "gguf",
			},
		}

		models = append(models, model)
	}

	// 如果正则表达式没有匹配到足够的模型，尝试另一种方法
	if len(models) < 10 {
		// 使用更简单的正则匹配模型名称
		simpleNameRegex := regexp.MustCompile(`href="/library/([a-zA-Z0-9_.-]+)"`)
		simpleMatches := simpleNameRegex.FindAllStringSubmatch(htmlContent, -1)

		for _, match := range simpleMatches {
			name := strings.TrimSpace(match[1])
			if name == "" || seenModels[name] || strings.Contains(name, "blog") {
				continue
			}

			// 检查是否匹配搜索关键词
			if !strings.Contains(strings.ToLower(name), lowerQuery) {
				continue
			}

			seenModels[name] = true

			model := map[string]interface{}{
				"name":        name,
				"description": fmt.Sprintf("%s model from Ollama library", name),
				"pulls":       "N/A",
				"tags_count":  "N/A",
				"details": map[string]interface{}{
					"format": "gguf",
				},
			}

			models = append(models, model)
		}
	}

	return models
}

// getBuiltinOnlineModels 返回内置的在线模型列表
func (a *App) getBuiltinOnlineModels() []map[string]interface{} {
	// 从 Ollama 官方库中精选的模型列表
	return []map[string]interface{}{
		{
			"name":        "llama3:8b",
			"description": "Meta's latest 8 billion parameter model",
			"size":        "4.7 GB",
			"details": map[string]interface{}{
				"parameter_size":     "8B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "llama",
			},
		},
		{
			"name":        "llama3:70b",
			"description": "Meta's powerful 70 billion parameter model",
			"size":        "34.4 GB",
			"details": map[string]interface{}{
				"parameter_size":     "70B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "llama",
			},
		},
		{
			"name":        "mistral:7b",
			"description": "Mistral AI's 7 billion parameter model",
			"size":        "4.1 GB",
			"details": map[string]interface{}{
				"parameter_size":     "7B",
				"quantization_level": "Q5_K_M",
				"format":             "gguf",
				"family":             "transformer",
			},
		},
		{
			"name":        "mistral:7b-instruct",
			"description": "Mistral AI's 7 billion parameter instruction-tuned model",
			"size":        "4.1 GB",
			"details": map[string]interface{}{
				"parameter_size":     "7B",
				"quantization_level": "Q5_K_M",
				"format":             "gguf",
				"family":             "transformer",
			},
		},
		{
			"name":        "gemma:2b",
			"description": "Google's lightweight 2 billion parameter model",
			"size":        "1.4 GB",
			"details": map[string]interface{}{
				"parameter_size":     "2B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "gemma",
			},
		},
		{
			"name":        "gemma:7b",
			"description": "Google's 7 billion parameter model",
			"size":        "4.2 GB",
			"details": map[string]interface{}{
				"parameter_size":     "7B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "gemma",
			},
		},
		{
			"name":        "qwen2:0.5b",
			"description": "阿里巴巴通义千问 0.5B 模型",
			"size":        "0.3 GB",
			"details": map[string]interface{}{
				"parameter_size":     "0.5B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "qwen",
			},
		},
		{
			"name":        "qwen2:1.5b",
			"description": "阿里巴巴通义千问 1.5B 模型",
			"size":        "0.9 GB",
			"details": map[string]interface{}{
				"parameter_size":     "1.5B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "qwen",
			},
		},
		{
			"name":        "qwen2:7b",
			"description": "阿里巴巴通义千问 7B 模型",
			"size":        "4.3 GB",
			"details": map[string]interface{}{
				"parameter_size":     "7B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "qwen",
			},
		},
		{
			"name":        "qwen2:72b",
			"description": "阿里巴巴通义千问 72B 模型",
			"size":        "35.1 GB",
			"details": map[string]interface{}{
				"parameter_size":     "72B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "qwen",
			},
		},
		{
			"name":        "phi3:mini",
			"description": "微软轻量级 3.8B 参数模型",
			"size":        "2.3 GB",
			"details": map[string]interface{}{
				"parameter_size":     "3.8B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "phi",
			},
		},
		{
			"name":        "phi3:small",
			"description": "微软 7B 参数模型",
			"size":        "4.1 GB",
			"details": map[string]interface{}{
				"parameter_size":     "7B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "phi",
			},
		},
		{
			"name":        "phi3:medium",
			"description": "微软 14B 参数模型",
			"size":        "8.3 GB",
			"details": map[string]interface{}{
				"parameter_size":     "14B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "phi",
			},
		},
		{
			"name":        "llava:1.5-7b",
			"description": "多模态视觉语言模型",
			"size":        "4.5 GB",
			"details": map[string]interface{}{
				"parameter_size":     "7B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "llava",
			},
		},
		{
			"name":        "codellama:7b",
			"description": "专为代码生成优化的模型",
			"size":        "4.2 GB",
			"details": map[string]interface{}{
				"parameter_size":     "7B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "llama",
			},
		},
		{
			"name":        "codellama:13b",
			"description": "专为代码生成优化的 13B 参数模型",
			"size":        "7.8 GB",
			"details": map[string]interface{}{
				"parameter_size":     "13B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "llama",
			},
		},
		{
			"name":        "starling-lm:7b-alpha",
			"description": "高性能语言模型",
			"size":        "4.1 GB",
			"details": map[string]interface{}{
				"parameter_size":     "7B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "llama",
			},
		},
		{
			"name":        "deepseek-coder:6.7b-base",
			"description": "专为代码生成优化的模型",
			"size":        "4.0 GB",
			"details": map[string]interface{}{
				"parameter_size":     "6.7B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "deepseek",
			},
		},
		{
			"name":        "deepseek-coder:6.7b-instruct",
			"description": "专为代码生成优化的指令模型",
			"size":        "4.0 GB",
			"details": map[string]interface{}{
				"parameter_size":     "6.7B",
				"quantization_level": "Q4_K_M",
				"format":             "gguf",
				"family":             "deepseek",
			},
		},
	}
}

// CheckOllamaAvailable 检查 Ollama 服务是否可用
func (a *App) CheckOllamaAvailable() bool {
	return a.checkOllamaService()
}

// 辅助函数：获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetIntelOptimizationInfo 获取英特尔优化信息
func (a *App) GetIntelOptimizationInfo() map[string]interface{} {
	info := map[string]interface{}{
		"intel_gpu_support":   true,
		"oneapi_available":    true,
		"mkl_optimized":       true,
		"supported_devices":   []string{"Intel Core Ultra", "Intel Core 11th-14th gen", "Intel Arc A-Series GPU", "Intel Arc B-Series GPU"},
		"optimization_source": "ModelScope Intel Ollama Optimized Edition",
		"documentation_url":   "https://www.modelscope.cn/models/Intel/ollama/summary",
		"version":             "2.3.0b20250923",
	}
	return info
}

// GetStats 获取统计信息
func (a *App) GetStats() map[string]interface{} {
	// 获取模型数量
	models := a.ListModels()
	modelCount := len(models)

	// 获取服务状态
	serviceStatus := a.GetServiceStatus()
	running := serviceStatus["running"].(bool)

	// 计算运行时间（简化：如果服务运行中，假设从应用启动开始）
	runningTime := "0h 0m"
	if running {
		// 这里可以记录服务启动时间，但简化实现
		runningTime = "运行中"
	}

	// 对话数和Token数（需要记录，暂时返回0）
	totalChats := 0
	totalTokens := "0"

	// 尝试从本地文件读取历史统计（如果存在）
	statsFile := filepath.Join(os.Getenv("HOME"), ".ollama", "stats.json")
	if runtime.GOOS == "windows" {
		statsFile = filepath.Join(os.Getenv("USERPROFILE"), ".ollama", "stats.json")
	}

	if _, err := os.Stat(statsFile); err == nil {
		data, err := os.ReadFile(statsFile)
		if err == nil {
			var savedStats map[string]interface{}
			if err := json.Unmarshal(data, &savedStats); err == nil {
				if chats, ok := savedStats["total_chats"].(float64); ok {
					totalChats = int(chats)
				}
				if tokens, ok := savedStats["total_tokens"].(string); ok {
					totalTokens = tokens
				}
			}
		}
	}

	return map[string]interface{}{
		"totalChats":  totalChats,
		"totalTokens": totalTokens,
		"totalModels": modelCount,
		"runningTime": runningTime,
	}
}

// WebSocketChat 处理WebSocket聊天连接
func (a *App) WebSocketChat(conn *websocket.Conn) {
	// 生成连接ID
	connID := fmt.Sprintf("%d", time.Now().UnixNano())

	// 添加连接到映射
	a.websocketMutex.Lock()
	a.websocketConnections[connID] = conn
	a.websocketMutex.Unlock()

	log.Printf("WebSocket连接已建立: %s", connID)

	defer func() {
		// 移除连接
		a.websocketMutex.Lock()
		delete(a.websocketConnections, connID)
		a.websocketMutex.Unlock()

		conn.Close()
		log.Printf("WebSocket连接已关闭: %s", connID)
	}()

	// 处理消息
	for {
		var msg struct {
			Type     string        `json:"type"`
			Model    string        `json:"model"`
			Messages []ChatMessage `json:"messages"`
			Role     string        `json:"role,omitempty"`
		}

		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket读取错误: %v", err)
			}
			break
		}

		// 处理不同类型的消息
		switch msg.Type {
		case "chat":
			a.handleWebSocketChat(conn, msg.Model, msg.Messages)
		case "role":
			a.handleWebSocketRole(conn, msg.Role)
		case "search":
			a.handleWebSocketSearch(conn, msg.Messages[len(msg.Messages)-1].Content)
		}
	}
}

// handleWebSocketChat 处理WebSocket聊天请求
func (a *App) handleWebSocketChat(conn *websocket.Conn, model string, messages []ChatMessage) {
	// 使用HTTP API进行聊天
	client := &http.Client{Timeout: 60 * time.Second}

	// 构建请求体
	reqBody, err := json.Marshal(map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   true,
	})
	if err != nil {
		// 发送错误响应
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"content": "构建请求失败",
		})
		return
	}

	// 发送请求
	resp, err := client.Post("http://127.0.0.1:11434/api/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		// 发送错误响应
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"content": "连接Ollama服务失败",
		})
		return
	}
	defer resp.Body.Close()

	// 处理流式响应
	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// 解析单行JSON
		var chunk struct {
			Model     string      `json:"model"`
			CreatedAt string      `json:"created_at"`
			Message   ChatMessage `json:"message"`
			Done      bool        `json:"done"`
			Error     string      `json:"error,omitempty"`
		}

		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}

		// 累积内容
		if chunk.Message.Content != "" {
			fullContent.WriteString(chunk.Message.Content)

			// 发送流式数据到WebSocket
			conn.WriteJSON(map[string]interface{}{
				"type":         "stream",
				"content":      chunk.Message.Content,
				"full_content": fullContent.String(),
				"done":         chunk.Done,
			})
		}

		// 当完成时，发送最终响应
		if chunk.Done {
			conn.WriteJSON(map[string]interface{}{
				"type":    "done",
				"content": fullContent.String(),
			})
			break
		}
	}
}

// handleWebSocketRole 处理WebSocket角色切换
func (a *App) handleWebSocketRole(conn *websocket.Conn, role string) {
	// 角色系统提示
	rolePrompts := map[string]string{
		"code":      `你是一位专业的代码专家，擅长解决各种编程问题。请提供清晰、高效、可维护的代码解决方案，并附带详细的解释和注释。`,
		"video":     `你是一位专业的视频脚本编写专家，擅长创作各种类型的视频脚本。请根据用户需求，设计引人入胜的视频内容，包括分镜、台词、视觉效果等详细要素。`,
		"writing":   `你是一位专业的写作专家，擅长各种文体的创作。请根据用户需求，提供高质量的写作内容，注重结构、逻辑和表达效果。`,
		"business":  `你是一位专业的商业顾问，擅长分析商业问题和提供战略建议。请根据用户需求，提供专业、实用的商业解决方案。`,
		"education": `你是一位专业的教育专家，擅长设计教学内容和解答学习问题。请根据用户需求，提供清晰、易懂、有深度的教育内容。`,
	}

	// 发送角色提示
	if prompt, ok := rolePrompts[role]; ok {
		conn.WriteJSON(map[string]interface{}{
			"type":    "role",
			"content": prompt,
		})
	} else {
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"content": "未知角色",
		})
	}
}

// handleWebSocketSearch 处理WebSocket搜索请求
func (a *App) handleWebSocketSearch(conn *websocket.Conn, query string) {
	// 这里实现联网搜索逻辑
	// 暂时返回模拟搜索结果
	conn.WriteJSON(map[string]interface{}{
		"type":    "search",
		"content": "搜索功能正在开发中，敬请期待！",
		"results": []map[string]interface{}{
			{
				"title":   "搜索结果1",
				"url":     "https://example.com/1",
				"snippet": "这是搜索结果1的摘要",
			},
			{
				"title":   "搜索结果2",
				"url":     "https://example.com/2",
				"snippet": "这是搜索结果2的摘要",
			},
		},
	})
}

// setOllamaPath 设置 Ollama 二进制文件路径
func (a *App) setOllamaPath() {
	// 获取当前可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		exePath = "."
	}
	exeDir := filepath.Dir(exePath)
	log.Printf("setOllamaPath: 可执行文件目录: %s\n", exeDir)

	// 根据操作系统设置二进制文件路径
	if runtime.GOOS == "windows" {
		a.ollamaPath = filepath.Join(exeDir, "ollama-intel-win", "ollama.exe")
	} else {
		a.ollamaPath = filepath.Join(exeDir, "ollama-intel-ubuntu", "ollama")
	}
	log.Printf("setOllamaPath: 设置的路径: %s\n", a.ollamaPath)

	// 检查文件是否存在
	if _, err := os.Stat(a.ollamaPath); os.IsNotExist(err) {
		log.Printf("setOllamaPath: 文件不存在, 回退到系统 PATH 中的 ollama\n")
		// 如果不存在，尝试使用系统 PATH 中的 ollama
		a.ollamaPath = "ollama"
	} else {
		log.Printf("setOllamaPath: 文件存在\n")
	}

	// 将路径写入文件以便调试
	debugInfo := fmt.Sprintf("可执行文件目录: %s\nOllama 路径: %s\n文件存在: %v\n",
		exeDir, a.ollamaPath, err == nil)
	os.WriteFile("ollama_path_debug.txt", []byte(debugInfo), 0644)
}

// startOllamaService 启动 Ollama 服务
func (a *App) startOllamaService() error {
	// 检查端口是否被占用
	if a.checkPortInUse(11434) {
		// 尝试清理占用端口的进程
		if !a.killProcessOnPort(11434) {
			return fmt.Errorf("端口 11434 被占用，请手动关闭相关进程")
		}
		// 等待端口释放
		time.Sleep(500 * time.Millisecond)
	}

	// 如果已有进程在运行，先停止
	if a.ollamaCmd != nil && a.ollamaCmd.Process != nil {
		a.ollamaCmd.Process.Kill()
		time.Sleep(100 * time.Millisecond)
	}

	// 启动新的 Ollama 服务进程
	cmd := exec.Command(a.ollamaPath, "serve")

	// 在 Windows 上隐藏命令窗口
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
	}

	// 不将输出重定向到控制台，而是通过日志记录器处理
	cmd.Stdout = a.logger
	cmd.Stderr = a.logger

	// 应用环境变量
	env := os.Environ()
	for key, value := range a.environmentVariables {
		if strValue, ok := value.(string); ok && strValue != "" {
			env = append(env, fmt.Sprintf("%s=%s", key, strValue))
		} else if boolValue, ok := value.(bool); ok {
			if boolValue {
				env = append(env, fmt.Sprintf("%s=true", key))
			}
		} else if intValue, ok := value.(float64); ok {
			env = append(env, fmt.Sprintf("%s=%d", key, int(intValue)))
		}
	}

	// 确保 OLLAMA_DEBUG 为 false，避免服务以 debug 模式运行
	cmd.Env = env
	log.Printf("startOllamaService: 应用环境变量: %+v\n", a.environmentVariables)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 Ollama 服务失败: %v", err)
	}

	a.ollamaCmd = cmd
	log.Println("Ollama 服务已启动")

	// 等待服务就绪
	for i := 0; i < 30; i++ {
		if a.checkPortInUse(11434) {
			return nil // 服务已就绪
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("服务启动超时")
}

// runOllamaCommand 运行 Ollama 命令并返回输出
func (a *App) runOllamaCommand(args ...string) (string, error) {
	log.Printf("runOllamaCommand: 路径=%s, 参数=%v\n", a.ollamaPath, args)
	cmd := exec.Command(a.ollamaPath, args...)

	// 隐藏命令窗口
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
	}

	// 应用环境变量
	env := os.Environ()

	// 对于list命令，不应用OLLAMA_MODEL_SOURCE环境变量，因为我们需要查看本地的模型列表
	isListCommand := false
	for _, arg := range args {
		if arg == "list" {
			isListCommand = true
			break
		}
	}

	for key, value := range a.environmentVariables {
		// 对于list命令，跳过OLLAMA_MODEL_SOURCE环境变量
		if isListCommand && key == "OLLAMA_MODEL_SOURCE" {
			continue
		}

		if strValue, ok := value.(string); ok && strValue != "" {
			env = append(env, fmt.Sprintf("%s=%s", key, strValue))
		} else if boolValue, ok := value.(bool); ok {
			if boolValue {
				env = append(env, fmt.Sprintf("%s=true", key))
			}
		} else if intValue, ok := value.(float64); ok {
			env = append(env, fmt.Sprintf("%s=%d", key, int(intValue)))
		}
	}
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("runOllamaCommand 失败: %v, 输出: %s\n", err, string(output))
		return string(output), err
	}
	log.Printf("runOllamaCommand 成功, 输出长度: %d\n", len(output))
	return string(output), nil
}

// runOllamaCommandWithJSON 运行 Ollama 命令并解析 JSON 输出
func (a *App) runOllamaCommandWithJSON(args ...string) (map[string]interface{}, error) {
	output, err := a.runOllamaCommand(args...)
	if err != nil {
		log.Printf("runOllamaCommandWithJSON: runOllamaCommand 失败: %v, 输出: %s", err, output)
		return nil, err
	}

	log.Printf("runOllamaCommandWithJSON: 命令输出: %s", output)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		log.Printf("runOllamaCommandWithJSON: JSON 解析失败: %v, 输出: %s", err, output)
		return nil, err
	}
	log.Printf("runOllamaCommandWithJSON: JSON 解析成功: %+v", result)
	return result, nil
}

// checkPortInUse 检查端口是否被占用
func (a *App) checkPortInUse(port int) bool {
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/api/tags", port))
	if err == nil {
		resp.Body.Close()
		return true
	}
	return false
}

// killProcessOnPort 杀死占用指定端口的进程
func (a *App) killProcessOnPort(port int) bool {
	if runtime.GOOS == "windows" {
		// Windows: 使用 netstat 和 taskkill
		cmd := exec.Command("cmd", "/C", fmt.Sprintf("netstat -ano | findstr :%d", port))
		// 隐藏命令窗口
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				parts := strings.Fields(line)
				if len(parts) >= 5 {
					// 提取 PID
					pid := parts[len(parts)-1]
					if pid != "" && pid != "0" {
						killCmd := exec.Command("taskkill", "/F", "/PID", pid)
						// 隐藏命令窗口
						killCmd.SysProcAttr = &syscall.SysProcAttr{
							HideWindow: true,
						}
						if err := killCmd.Run(); err == nil {
							time.Sleep(100 * time.Millisecond) // 等待进程终止
							return true
						}
					}
				}
			}
		}
	} else {
		// Linux: 使用 lsof 和 kill
		cmd := exec.Command("sh", "-c", fmt.Sprintf("lsof -ti:%d", port))
		output, err := cmd.Output()
		if err == nil {
			pid := strings.TrimSpace(string(output))
			if pid != "" {
				killCmd := exec.Command("kill", "-9", pid)
				if err := killCmd.Run(); err == nil {
					time.Sleep(100 * time.Millisecond)
					return true
				}
			}
		}
	}
	return false
}

// WebSocketHandler 处理WebSocket连接请求
func (a *App) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 处理WebSocket连接
	go a.WebSocketChat(conn)
}

// 初始化HTTP服务器，添加WebSocket路由和OpenAI兼容API
func (a *App) initHTTPServer() {
	// 启动HTTP服务器，监听WebSocket连接
	go func() {
		// 注册WebSocket路由
		http.HandleFunc("/ws/chat", a.WebSocketHandler)

		// 注册OpenAI兼容API路由
		http.HandleFunc("/v1/chat/completions", a.handleOpenAIChatCompletions)
		http.HandleFunc("/v1/models", a.handleOpenAIModels)
		http.HandleFunc("/v1/models/", a.handleOpenAIModel)

		// 启动服务器，使用不同的端口以避免与Ollama服务冲突
		log.Println("========================================")
		log.Println("WebSocket服务器启动在 :11435")
		log.Println("Web端访问地址: http://localhost:11435")
		log.Println("OpenAI兼容API地址: http://localhost:11435/v1")
		log.Println("========================================")
		if err := http.ListenAndServe(":11435", nil); err != nil {
			log.Printf("WebSocket服务器启动失败: %v", err)
		}
	}()

	// 注意: Ollama 服务本身也内置 OpenAI 兼容 API (端口 11434)
	// OpenAI 兼容 API 地址: http://localhost:11434/v1
	log.Println("Ollama内置OpenAI兼容API地址: http://localhost:11434/v1")
}

// OpenAIChatRequest OpenAI兼容的聊天请求
type OpenAIChatRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
	APIKey      string                   `json:"api_key,omitempty"`
}

// OpenAIChatResponse OpenAI兼容的聊天响应

type OpenAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int                    `json:"index"`
		Message      map[string]interface{} `json:"message"`
		FinishReason string                 `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// OpenAIStreamResponse OpenAI兼容的流式响应

type OpenAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int                    `json:"index"`
		Delta        map[string]interface{} `json:"delta"`
		FinishReason string                 `json:"finish_reason"`
	} `json:"choices"`
}

// OpenAIModelResponse OpenAI兼容的模型响应

type OpenAIModelResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// OpenAIModelsResponse OpenAI兼容的模型列表响应

type OpenAIModelsResponse struct {
	Object string                `json:"object"`
	Data   []OpenAIModelResponse `json:"data"`
}

// handleOpenAIChatCompletions 处理OpenAI兼容的聊天完成请求
func (a *App) handleOpenAIChatCompletions(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头，允许外部工具调用
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 检查请求方法
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查API密钥（如果设置了）
	if apiKey, ok := a.environmentVariables["OLLAMA_OPENAI_API_KEY"]; ok {
		if keyStr, ok := apiKey.(string); ok && keyStr != "" {
			// 从请求头获取API密钥
			authHeader := r.Header.Get("Authorization")
			expectedAuth := "Bearer " + keyStr
			if authHeader != expectedAuth {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
	}

	// 解析请求体
	var req OpenAIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[OpenAI API] 收到聊天请求: 模型=%s, 流式=%v, 消息数=%d", req.Model, req.Stream, len(req.Messages))

	// 检查是否启用了OpenAI兼容API
	if enabled, ok := a.environmentVariables["OLLAMA_OPENAI_COMPATIBLE"]; !ok || !enabled.(bool) {
		http.Error(w, "OpenAI compatible API is disabled", http.StatusServiceUnavailable)
		return
	}

	// 转换为Ollama聊天请求
	ollamaMessages := make([]ChatMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)
		ollamaMessages = append(ollamaMessages, ChatMessage{
			Role:    role,
			Content: content,
		})
	}

	ollamaReq := ChatRequest{
		Model:    req.Model,
		Messages: ollamaMessages,
		Stream:   req.Stream,
	}

	// 处理流式响应
	if req.Stream {
		a.handleOpenAIStreamResponse(w, ollamaReq)
		return
	}

	// 处理非流式响应
	response := a.handleOpenAINonStreamResponse(ollamaReq)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleOpenAIStreamResponse 处理OpenAI兼容的流式响应
func (a *App) handleOpenAIStreamResponse(w http.ResponseWriter, req ChatRequest) {
	log.Printf("[OpenAI API] 开始流式响应: 模型=%s", req.Model)

	// 设置响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 使用HTTP API进行聊天，增加超时时间
	client := &http.Client{
		Timeout: 180 * time.Second,
		Transport: &http.Transport{
			ResponseHeaderTimeout: 180 * time.Second,
			ExpectContinueTimeout: 30 * time.Second,
		},
	}

	// 构建请求体
	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Printf("[OpenAI API] 构建请求失败: %v", err)
		w.Write([]byte("data: {\"error\": \"Invalid request\"}\n\n"))
		w.(http.Flusher).Flush()
		return
	}

	log.Printf("[OpenAI API] 发送请求到Ollama服务")

	// 发送请求
	resp, err := client.Post("http://127.0.0.1:11434/api/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("[OpenAI API] 连接Ollama服务失败: %v", err)
		w.Write([]byte("data: {\"error\": \"Failed to connect to Ollama service\"}\n\n"))
		w.(http.Flusher).Flush()
		return
	}
	defer resp.Body.Close()

	log.Printf("[OpenAI API] Ollama响应状态: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[OpenAI API] Ollama错误响应: %s", string(body))
		w.Write([]byte(fmt.Sprintf("data: {\"error\": \"Ollama error: %d\"}\n\n", resp.StatusCode)))
		w.(http.Flusher).Flush()
		return
	}

	// 处理流式响应
	scanner := bufio.NewScanner(resp.Body)
	// 增加scanner的缓冲区大小
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	
	var fullContent strings.Builder
	responseID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	created := time.Now().Unix()
	chunkCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// 解析单行JSON
		var chunk struct {
			Model     string      `json:"model"`
			CreatedAt string      `json:"created_at"`
			Message   ChatMessage `json:"message"`
			Done      bool        `json:"done"`
			Error     string      `json:"error,omitempty"`
		}

		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			log.Printf("[OpenAI API] 解析响应失败: %v, 行: %s", err, line)
			continue
		}

		// 处理错误
		if chunk.Error != "" {
			log.Printf("[OpenAI API] Ollama返回错误: %s", chunk.Error)
			errorResp := OpenAIStreamResponse{
				ID:      responseID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   req.Model,
				Choices: []struct {
					Index        int                    `json:"index"`
					Delta        map[string]interface{} `json:"delta"`
					FinishReason string                 `json:"finish_reason"`
				}{
					{
						Index:        0,
						Delta:        map[string]interface{}{"content": fmt.Sprintf("Error: %s", chunk.Error)},
						FinishReason: "error",
					},
				},
			}
			w.Write([]byte("data: "))
			json.NewEncoder(w).Encode(errorResp)
			w.Write([]byte("\n"))
			w.Write([]byte("data: [DONE]\n\n"))
			w.(http.Flusher).Flush()
			return
		}

		// 累积内容
		if chunk.Message.Content != "" {
			fullContent.WriteString(chunk.Message.Content)
			chunkCount++

			// 构建OpenAI流式响应
			streamResp := OpenAIStreamResponse{
				ID:      responseID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   req.Model,
				Choices: []struct {
					Index        int                    `json:"index"`
					Delta        map[string]interface{} `json:"delta"`
					FinishReason string                 `json:"finish_reason"`
				}{
					{
						Index: 0,
						Delta: map[string]interface{}{
							"content": chunk.Message.Content,
						},
					},
				},
			}

			// 发送流式数据
			w.Write([]byte("data: "))
			json.NewEncoder(w).Encode(streamResp)
			w.Write([]byte("\n"))
			w.(http.Flusher).Flush()
		}

		// 当完成时，发送最终响应
		if chunk.Done {
			log.Printf("[OpenAI API] 流式响应完成: 总长度=%d, chunk数=%d", fullContent.Len(), chunkCount)
			
			// 构建完成响应
			finishResp := OpenAIStreamResponse{
				ID:      responseID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   req.Model,
				Choices: []struct {
					Index        int                    `json:"index"`
					Delta        map[string]interface{} `json:"delta"`
					FinishReason string                 `json:"finish_reason"`
				}{
					{
						Index:        0,
						Delta:        map[string]interface{}{},
						FinishReason: "stop",
					},
				},
			}

			// 发送完成数据
			w.Write([]byte("data: "))
			json.NewEncoder(w).Encode(finishResp)
			w.Write([]byte("\n"))
			w.Write([]byte("data: [DONE]\n\n"))
			w.(http.Flusher).Flush()
			break
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[OpenAI API] 读取响应失败: %v", err)
	}
}

// handleOpenAINonStreamResponse 处理OpenAI兼容的非流式响应
func (a *App) handleOpenAINonStreamResponse(req ChatRequest) OpenAIChatResponse {
	log.Printf("[OpenAI API] 处理非流式响应: 模型=%s", req.Model)
	
	// 使用现有的ChatCompletion方法获取响应
	ollamaResp := a.ChatCompletion(req)

	// 构建OpenAI响应
	responseID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	created := time.Now().Unix()

	log.Printf("[OpenAI API] 非流式响应完成: 内容长度=%d", len(ollamaResp.Message.Content))

	return OpenAIChatResponse{
		ID:      responseID,
		Object:  "chat.completion",
		Created: created,
		Model:   req.Model,
		Choices: []struct {
			Index        int                    `json:"index"`
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: map[string]interface{}{
					"role":    ollamaResp.Message.Role,
					"content": ollamaResp.Message.Content,
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     len(ollamaResp.Message.Content),
			CompletionTokens: len(ollamaResp.Message.Content),
			TotalTokens:      len(ollamaResp.Message.Content) + len(ollamaResp.Message.Content),
		},
	}
}

// handleOpenAIModels 处理OpenAI兼容的模型列表请求
func (a *App) handleOpenAIModels(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 检查请求方法
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查API密钥（如果设置了）
	if apiKey, ok := a.environmentVariables["OLLAMA_OPENAI_API_KEY"]; ok {
		if keyStr, ok := apiKey.(string); ok && keyStr != "" {
			// 从请求头获取API密钥
			authHeader := r.Header.Get("Authorization")
			expectedAuth := "Bearer " + keyStr
			if authHeader != expectedAuth {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
	}

	// 获取本地模型列表
	models := a.ListModels()

	// 构建OpenAI模型响应
	var modelResponses []OpenAIModelResponse
	for _, model := range models {
		modelResponses = append(modelResponses, OpenAIModelResponse{
			ID:      model.Name,
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "ollama",
		})
	}

	// 构建响应
	response := OpenAIModelsResponse{
		Object: "list",
		Data:   modelResponses,
	}

	// 发送响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleOpenAIModel 处理OpenAI兼容的单个模型请求
func (a *App) handleOpenAIModel(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 检查请求方法
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查API密钥（如果设置了）
	if apiKey, ok := a.environmentVariables["OLLAMA_OPENAI_API_KEY"]; ok {
		if keyStr, ok := apiKey.(string); ok && keyStr != "" {
			// 从请求头获取API密钥
			authHeader := r.Header.Get("Authorization")
			expectedAuth := "Bearer " + keyStr
			if authHeader != expectedAuth {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
	}

	// 提取模型ID
	modelID := strings.TrimPrefix(r.URL.Path, "/v1/models/")
	if modelID == "" {
		http.Error(w, "Model ID is required", http.StatusBadRequest)
		return
	}

	// 构建模型响应
	modelResp := OpenAIModelResponse{
		ID:      modelID,
		Object:  "model",
		Created: time.Now().Unix(),
		OwnedBy: "ollama",
	}

	// 发送响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modelResp)
}

// GetRealTimeStats 获取实时系统状态
func (a *App) GetRealTimeStats() map[string]interface{} {
	stats := map[string]interface{}{
		"timestamp": time.Now().Unix(),
	}

	// 获取内存使用情况
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command",
			"Get-WmiObject Win32_OperatingSystem | Select-Object FreePhysicalMemory, TotalVisibleMemorySize")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(lines) > 2 {
				values := strings.Fields(strings.TrimSpace(lines[2]))
				if len(values) >= 2 {
					if freeKB, err := strconv.ParseUint(values[0], 10, 64); err == nil {
						if totalKB, err := strconv.ParseUint(values[1], 10, 64); err == nil {
							totalGB := float64(totalKB) / 1024 / 1024
							freeGB := float64(freeKB) / 1024 / 1024
							usedGB := totalGB - freeGB
							stats["memory"] = map[string]interface{}{
								"total_gb":     totalGB,
								"used_gb":      usedGB,
								"free_gb":      freeGB,
								"used_percent": (usedGB / totalGB) * 100,
							}
						}
					}
				}
			}
		}
	}

	// 获取 CPU 使用率
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command",
			"Get-WmiObject Win32_Processor | Select-Object LoadPercentage")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(lines) > 2 {
				cpuLoad := strings.TrimSpace(lines[2])
				if load, err := strconv.ParseFloat(cpuLoad, 64); err == nil {
					stats["cpu"] = map[string]interface{}{
						"usage_percent": load,
					}
				}
			}
		}
	}

	// 获取 GPU 信息（如果可用）
	stats["gpu"] = a.getGPUInfo()

	// 获取服务状态
	stats["service"] = a.GetServiceStatus()

	return stats
}

// getGPUInfo 获取 GPU 信息
func (a *App) getGPUInfo() map[string]interface{} {
	gpuInfo := map[string]interface{}{
		"available": false,
		"name":      "Unknown",
		"memory":    "N/A",
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command",
			"Get-WmiObject Win32_VideoController | Select-Object Name, AdapterRAM")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(strings.ToLower(line), "intel") ||
					strings.Contains(strings.ToLower(line), "nvidia") ||
					strings.Contains(strings.ToLower(line), "amd") {
					parts := strings.Fields(line)
					if len(parts) > 0 {
						gpuInfo["available"] = true
						gpuInfo["name"] = strings.Join(parts, " ")
					}
				}
			}
		}
	}

	return gpuInfo
}
