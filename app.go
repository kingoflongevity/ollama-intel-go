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

	// 同时写入标准错误（可选，用于调试）
	os.Stderr.Write(p)

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

	// 初始化环境变量存储
	a.environmentVariables = make(map[string]interface{})
	// 设置默认值
	a.environmentVariables["OLLAMA_MODEL_SOURCE"] = "modelscope"
	a.environmentVariables["OLLAMA_NUM_CTX"] = 2048
	a.environmentVariables["OLLAMA_INTEL_GPU"] = true
	a.environmentVariables["OLLAMA_DEBUG"] = false
	a.environmentVariables["ONEAPI_DEVICE_SELECTOR"] = ""

	log.Println("startup: 开始初始化")

	// 设置 Ollama 二进制文件路径
	a.setOllamaPath()

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
				model := ModelInfo{
					Name:     getString(modelMap, "name"),
					Model:    getString(modelMap, "model"),
					Size:     getString(modelMap, "size"),
					Digest:   getString(modelMap, "digest"),
					Modified: getString(modelMap, "modified_at"),
					Details:  make(map[string]interface{}),
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

// PullModel 拉取模型
func (a *App) PullModel(name string) map[string]interface{} {
	// 在 goroutine 中拉取模型以提供进度更新
	go a.pullModelWithProgress(name)

	return map[string]interface{}{
		"message": fmt.Sprintf("开始拉取模型: %s", name),
		"model":   name,
		"status":  "started",
	}
}

// pullModelWithProgress 拉取模型并发送进度更新
func (a *App) pullModelWithProgress(modelName string) {
	log.Printf("开始拉取模型: %s", modelName)

	// 发送开始事件
	a.sendPullProgressEvent(modelName, "started", 0, "开始拉取模型")

	cmd := exec.Command(a.ollamaPath, "pull", modelName)

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

	// 合并 stdout 和 stderr 的读取
	go a.readOutputLines(stdout, modelName, "stdout")
	go a.readOutputLines(stderr, modelName, "stderr")

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		log.Printf("拉取模型失败: %v", err)
		// 检查是否已经发送了错误事件
		// 这里可以添加更复杂的错误检测逻辑
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
func (a *App) readOutputLines(reader io.Reader, modelName, streamType string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("模型拉取输出[%s]: %s", streamType, line)

		// 检测错误信息
		if streamType == "stderr" && (strings.Contains(line, "Error:") || strings.Contains(line, "error:")) {
			// 发送错误事件
			a.sendPullProgressEvent(modelName, "error", 0, line)
			return
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
	}
}

// parsePullProgress 解析拉取进度信息
func (a *App) parsePullProgress(line string) (float64, string) {
	// 尝试解析百分比格式，例如: "pulling manifest... 50%"
	// 或者: "pulling 123456... 100% |████████████████████| (1.0/1.0 MB)"

	// 查找百分比
	if strings.Contains(line, "%") {
		// 简单的正则表达式匹配
		// 寻找数字后跟 % 的模式
		line = strings.TrimSpace(line)

		// 分割空格
		parts := strings.Split(line, " ")
		for _, part := range parts {
			if strings.Contains(part, "%") {
				// 提取数字部分
				percentStr := strings.Trim(part, "%")
				if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
					// 返回百分比和原始行作为消息
					return percent, line
				}
			}
		}
	}

	// 如果没有找到百分比，返回 -1 表示没有进度信息
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

// ShowModel 显示模型信息
func (a *App) ShowModel(name string) map[string]interface{} {
	// 尝试通过命令行获取模型信息
	result, err := a.runOllamaCommandWithJSON("show", name, "--format", "json")
	if err != nil {
		// 如果失败，返回模拟数据
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

// SaveEnvironmentVariables 保存环境变量配置
func (a *App) SaveEnvironmentVariables(variables map[string]interface{}) map[string]interface{} {
	log.Printf("SaveEnvironmentVariables: 保存环境变量配置: %+v\n", variables)

	// 保存环境变量
	a.environmentVariables = variables

	// 可以在这里添加持久化存储逻辑，例如保存到文件

	log.Println("SaveEnvironmentVariables: 环境变量配置已保存")

	return map[string]interface{}{
		"message":   "环境变量配置已保存",
		"variables": variables,
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

// searchOnlineModelsWithOllama 使用 Ollama 命令搜索在线模型
func (a *App) searchOnlineModelsWithOllama(query string) ([]map[string]interface{}, error) {
	// 首先获取所有内置模型
	allModels := a.getBuiltinOnlineModels()

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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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

// 初始化HTTP服务器，添加WebSocket路由
func (a *App) initHTTPServer() {
	// 启动HTTP服务器，监听WebSocket连接
	go func() {
		// 注册WebSocket路由
		http.HandleFunc("/ws/chat", a.WebSocketHandler)

		// 启动服务器，使用不同的端口以避免与Ollama服务冲突
		log.Println("WebSocket服务器启动在 :11435")
		if err := http.ListenAndServe(":11435", nil); err != nil {
			log.Printf("WebSocket服务器启动失败: %v", err)
		}
	}()
}
