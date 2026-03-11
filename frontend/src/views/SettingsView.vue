<template>
  <div class="tech-settings-container">
    <div class="tech-page-header">
      <div class="header-left">
        <div class="header-title">
          <div class="title-icon">
            <el-icon><Setting /></el-icon>
          </div>
          <div class="title-text">
            <h2>系统设置</h2>
            <span class="subtitle">System Settings</span>
          </div>
        </div>
      </div>
    </div>

    <div class="settings-content">
      <div class="settings-sidebar">
        <div
          v-for="tab in tabs"
          :key="tab.name"
          class="sidebar-item"
          :class="{ active: activeTab === tab.name }"
          @click="activeTab = tab.name"
        >
          <el-icon><component :is="tab.icon" /></el-icon>
          <span>{{ tab.label }}</span>
        </div>
      </div>

      <div class="settings-main">
        <div v-show="activeTab === 'general'" class="settings-panel">
          <div class="panel-header">
            <h3>通用设置</h3>
            <span>界面与基础配置</span>
          </div>
          <div class="panel-body">
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">主题模式</span>
                <span class="setting-desc">切换暗色/亮色主题</span>
              </div>
              <div class="setting-control">
                <el-switch
                  v-model="themeMode"
                  inline-prompt
                  active-text="暗色"
                  inactive-text="亮色"
                  @change="toggleTheme"
                  class="tech-switch"
                />
              </div>
            </div>
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">语言</span>
                <span class="setting-desc">选择界面显示语言</span>
              </div>
              <div class="setting-control">
                <el-select v-model="language" placeholder="选择语言" @change="changeLanguage" class="tech-select">
                  <el-option label="简体中文" value="zh-CN" />
                  <el-option label="English" value="en-US" />
                  <el-option label="日本語" value="ja-JP" />
                </el-select>
              </div>
            </div>
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">自动检查更新</span>
                <span class="setting-desc">启动时自动检查新版本</span>
              </div>
              <div class="setting-control">
                <el-switch v-model="autoUpdate" class="tech-switch" />
              </div>
            </div>
          </div>
        </div>

        <div v-show="activeTab === 'service'" class="settings-panel">
          <div class="panel-header">
            <h3>服务管理</h3>
            <span>Ollama 服务配置与控制</span>
          </div>
          <div class="panel-body">
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">服务状态</span>
                <span class="setting-desc">当前 Ollama 服务运行状态</span>
              </div>
              <div class="setting-control">
                <span class="status-badge" :class="serviceStatus.running ? 'running' : 'stopped'">
                  {{ serviceStatus.running ? '运行中' : '已停止' }}
                </span>
                <button
                  class="tech-btn"
                  :class="serviceStatus.running ? 'tech-btn-danger' : 'tech-btn-success'"
                  @click="toggleService"
                >
                  {{ serviceStatus.running ? '停止服务' : '启动服务' }}
                </button>
              </div>
            </div>
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">服务地址</span>
                <span class="setting-desc">Ollama API 监听地址 (可修改)</span>
              </div>
              <div class="setting-control">
                <el-input v-model="serviceAddress" class="tech-input" placeholder="例如: 127.0.0.1:11434" />
                <button class="tech-btn tech-btn-primary tech-btn-sm" @click="saveServiceAddress">
                  <el-icon><Check /></el-icon>
                  保存
                </button>
              </div>
            </div>
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">服务版本</span>
                <span class="setting-desc">当前安装的 Ollama 版本</span>
              </div>
              <div class="setting-control">
                <el-input v-model="serviceVersion" disabled class="tech-input" />
              </div>
            </div>
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">自动启动</span>
                <span class="setting-desc">应用启动时自动启动 Ollama 服务</span>
              </div>
              <div class="setting-control">
                <el-switch v-model="autoStartService" class="tech-switch" />
              </div>
            </div>
          </div>

          <div class="panel-section">
            <div class="section-header">
              <h4>环境变量配置</h4>
              <button class="tech-btn tech-btn-primary tech-btn-sm" @click="saveEnvironmentVariables">
                <el-icon><Check /></el-icon>
                保存配置
              </button>
            </div>
            <div class="panel-body">
              <div class="setting-item">
                <div class="setting-info">
                  <span class="setting-label">模型下载源</span>
                  <span class="setting-desc">选择镜像源加速模型下载</span>
                </div>
                <div class="setting-control">
                  <el-select v-model="environmentVariables.OLLAMA_MODEL_SOURCE" placeholder="选择模型下载源" class="tech-select">
                    <el-option label="Ollama 官方 (默认)" value="" />
                    <el-option label="ModelScope 镜像 (国内推荐)" value="modelscope" />
                    <el-option label="阿里云镜像" value="aliyun" />
                  </el-select>
                </div>
              </div>
              <div class="setting-item">
                <div class="setting-info">
                  <span class="setting-label">允许跨域请求</span>
                  <span class="setting-desc">设置允许跨域请求的源，* 表示允许所有</span>
                </div>
                <div class="setting-control">
                  <el-input v-model="environmentVariables.OLLAMA_ORIGINS" placeholder="例如: * 或 http://localhost:3000" class="tech-input" />
                </div>
              </div>
              <div class="setting-item">
                <div class="setting-info">
                  <span class="setting-label">Ollama 服务地址</span>
                  <span class="setting-desc">设置 Ollama 服务监听地址</span>
                </div>
                <div class="setting-control">
                  <el-input v-model="environmentVariables.OLLAMA_HOST" placeholder="例如: 0.0.0.0:11434" class="tech-input" />
                </div>
              </div>
              <div class="setting-item">
                <div class="setting-info">
                  <span class="setting-label">上下文长度</span>
                  <span class="setting-desc">设置模型的上下文窗口大小</span>
                </div>
                <div class="setting-control">
                  <el-input-number v-model="environmentVariables.OLLAMA_NUM_CTX" :min="2048" :max="1048576" :step="2048" class="tech-number" />
                </div>
              </div>
              <div class="setting-item">
                <div class="setting-info">
                  <span class="setting-label">GPU 选择</span>
                  <span class="setting-desc">在多 GPU 环境中选择特定 GPU</span>
                </div>
                <div class="setting-control">
                  <el-input v-model="environmentVariables.ONEAPI_DEVICE_SELECTOR" placeholder="例如: gpu:0" class="tech-input" />
                </div>
              </div>
              <div class="setting-item">
                <div class="setting-info">
                  <span class="setting-label">启用 GPU 加速</span>
                  <span class="setting-desc">启用 Intel GPU 加速</span>
                </div>
                <div class="setting-control">
                  <el-switch v-model="environmentVariables.OLLAMA_INTEL_GPU" class="tech-switch" />
                </div>
              </div>
              <div class="setting-item">
                <div class="setting-info">
                  <span class="setting-label">Ollama 执行程序路径</span>
                  <span class="setting-desc">自定义 Ollama 执行程序位置</span>
                </div>
                <div class="setting-control full-width">
                  <el-input v-model="environmentVariables.OLLAMA_EXECUTABLE_PATH" placeholder="例如: C:\ollama\ollama.exe" class="tech-input" />
                  <span class="path-hint">当前路径: {{ currentOllamaPath }}</span>
                </div>
              </div>
              <div class="setting-item">
                <div class="setting-info">
                  <span class="setting-label">调试模式</span>
                  <span class="setting-desc">启用 Ollama 调试日志</span>
                </div>
                <div class="setting-control">
                  <el-switch v-model="environmentVariables.OLLAMA_DEBUG" class="tech-switch" />
                </div>
              </div>
            </div>
          </div>

          <div class="panel-section">
            <div class="section-header">
              <h4>OpenAI 兼容 API</h4>
            </div>
            <div class="panel-body">
              <div class="setting-item">
                <div class="setting-info">
                  <span class="setting-label">启用 OpenAI 兼容 API</span>
                  <span class="setting-desc">允许其他应用像调用 OpenAI 一样调用本地模型</span>
                </div>
                <div class="setting-control">
                  <el-switch v-model="environmentVariables.OLLAMA_OPENAI_COMPATIBLE" class="tech-switch" />
                </div>
              </div>
              <template v-if="environmentVariables.OLLAMA_OPENAI_COMPATIBLE">
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">API 密钥</span>
                    <span class="setting-desc">设置用于验证 API 调用的密钥（可选）</span>
                  </div>
                  <div class="setting-control">
                    <el-input v-model="environmentVariables.OLLAMA_OPENAI_API_KEY" placeholder="设置 API 密钥（可选）" class="tech-input" />
                  </div>
                </div>
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">API 端口</span>
                    <span class="setting-desc">设置 OpenAI 兼容 API 服务端口</span>
                  </div>
                  <div class="setting-control">
                    <el-input-number v-model="environmentVariables.OLLAMA_OPENAI_PORT" :min="1024" :max="65535" :step="1" class="tech-number" />
                  </div>
                </div>
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">API 调用地址</span>
                    <span class="setting-desc">Ollama 内置的 OpenAI 兼容 API 地址</span>
                  </div>
                  <div class="setting-control">
                    <div class="api-url-box">
                      <span class="api-url">{{ openaiApiUrl }}</span>
                      <button class="tech-btn tech-btn-sm" @click="copyApiUrl">
                        <el-icon><CopyDocument /></el-icon>
                        复制
                      </button>
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div v-show="activeTab === 'intel'" class="settings-panel">
          <div class="panel-header">
            <h3>英特尔优化</h3>
            <span>Intel 硬件加速配置</span>
          </div>
          <div class="panel-body">
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">GPU 加速</span>
                <span class="setting-desc">启用 Intel Arc GPU 加速</span>
              </div>
              <div class="setting-control">
                <el-switch v-model="gpuAcceleration" class="tech-switch" />
              </div>
            </div>
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">OneAPI 优化</span>
                <span class="setting-desc">启用 Intel OneAPI 优化</span>
              </div>
              <div class="setting-control">
                <el-switch v-model="oneApiOptimization" class="tech-switch" />
              </div>
            </div>
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">Intel MKL</span>
                <span class="setting-desc">启用 Intel Math Kernel Library</span>
              </div>
              <div class="setting-control">
                <el-switch v-model="intelMKL" class="tech-switch" />
              </div>
            </div>
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-label">硬件检测结果</span>
                <span class="setting-desc">系统硬件信息</span>
              </div>
              <div class="setting-control full-width">
                <div class="hardware-info">
                  <div class="hw-item">
                    <span class="hw-label">CPU</span>
                    <span class="hw-value">{{ hardwareInfo.cpu }}</span>
                  </div>
                  <div class="hw-item">
                    <span class="hw-label">GPU</span>
                    <span class="hw-value">{{ hardwareInfo.gpu }}</span>
                  </div>
                  <div class="hw-item">
                    <span class="hw-label">内存</span>
                    <span class="hw-value">{{ hardwareInfo.memory }}</span>
                  </div>
                  <div class="hw-item">
                    <span class="hw-label">指令集</span>
                    <span class="hw-value">{{ hardwareInfo.instructions }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-show="activeTab === 'about'" class="settings-panel">
          <div class="panel-header">
            <h3>关于</h3>
            <span>应用程序信息</span>
          </div>
          <div class="panel-body">
            <div class="about-section">
              <div class="about-logo">
                <div class="logo-icon">
                  <el-icon><Cpu /></el-icon>
                </div>
                <div class="logo-info">
                  <h4>Ollama 英特尔优化版</h4>
                  <span>专为英特尔硬件优化的桌面应用</span>
                </div>
              </div>
              <div class="about-info">
                <div class="info-row">
                  <span class="info-label">版本</span>
                  <span class="info-value">1.0.0</span>
                </div>
                <div class="info-row">
                  <span class="info-label">构建时间</span>
                  <span class="info-value">{{ buildTime }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">技术栈</span>
                  <span class="info-value">Go + Vue 3 + Element Plus</span>
                </div>
                <div class="info-row">
                  <span class="info-label">优化</span>
                  <span class="info-value">专为英特尔硬件优化</span>
                </div>
              </div>
              <div class="about-desc">
                <p>Ollama 英特尔优化版是一个专为英特尔硬件优化的桌面应用解决方案，用于在 Windows、macOS 和 Linux 操作系统上运行和管理 Ollama 模型的 GUI 工具。</p>
                <p>该版本特别针对英特尔处理器、Intel Arc GPU 和相关技术栈进行了深度优化，提供最佳性能体验。</p>
              </div>
              <div class="about-actions">
                <button class="tech-btn tech-btn-primary" @click="openWebsite">
                  <el-icon><Link /></el-icon>
                  官方网站
                </button>
                <button class="tech-btn" @click="openDocumentation">
                  <el-icon><Document /></el-icon>
                  使用文档
                </button>
                <button class="tech-btn" @click="checkForUpdates">
                  <el-icon><Refresh /></el-icon>
                  检查更新
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Setting, Monitor, Cpu, InfoFilled, Check, CopyDocument, Link, Document, Refresh } from '@element-plus/icons-vue'
import { GetServiceStatus, StartService, StopService, GetEnvironmentInfo, GetIntelOptimizationInfo, GetEnvironmentVariables, SaveEnvironmentVariables, GetOllamaPath } from '../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'
import { useConfigStore } from '@/stores/configStore'

const configStore = useConfigStore()

const tabs = [
  { name: 'general', label: '通用', icon: Setting },
  { name: 'service', label: '服务', icon: Monitor },
  { name: 'intel', label: '英特尔优化', icon: Cpu },
  { name: 'about', label: '关于', icon: InfoFilled }
]

const activeTab = ref('general')
const themeMode = ref(false)
const language = ref('zh-CN')
const autoUpdate = ref(true)
const serviceStatus = ref({
  running: false,
  host: '127.0.0.1:11434',
  version: 'unknown'
})
const serviceAddress = ref('127.0.0.1:11434')
const serviceVersion = ref('unknown')
const autoStartService = ref(false)
const gpuAcceleration = ref(true)
const oneApiOptimization = ref(true)
const intelMKL = ref(true)
const hardwareInfo = ref({
  cpu: '检测中...',
  gpu: '检测中...',
  memory: '检测中...',
  instructions: '检测中...'
})
const buildTime = ref(new Date().toLocaleDateString())
const environmentVariables = ref({
  OLLAMA_MODEL_SOURCE: '',
  OLLAMA_ORIGINS: '*',
  OLLAMA_HOST: '',
  OLLAMA_NUM_CTX: 2048,
  ONEAPI_DEVICE_SELECTOR: '',
  OLLAMA_INTEL_GPU: true,
  OLLAMA_EXECUTABLE_PATH: '',
  OLLAMA_DEBUG: false,
  OLLAMA_OPENAI_COMPATIBLE: false,
  OLLAMA_OPENAI_API_KEY: '',
  OLLAMA_OPENAI_PORT: 8080
})
const currentOllamaPath = ref('检测中...')

const toggleTheme = (val) => {
  document.body.classList.toggle('dark-theme', val)
  localStorage.setItem('theme', val ? 'dark' : 'light')
}

const changeLanguage = (lang) => {
  language.value = lang
  localStorage.setItem('language', lang)
  ElMessage.success(`语言已切换为 ${lang === 'zh-CN' ? '简体中文' : lang === 'en-US' ? 'English' : '日本語'}`)
}

const toggleService = async () => {
  try {
    if (serviceStatus.value.running) {
      await StopService()
      ElMessage.success('服务已停止')
    } else {
      await StartService()
      ElMessage.success('服务启动中...')
    }
    setTimeout(() => {
      loadServiceStatus()
    }, 1000)
  } catch (error) {
    console.error('服务控制失败:', error)
    ElMessage.error('服务控制失败')
  }
}

const loadServiceStatus = async () => {
  try {
    const status = await GetServiceStatus()
    serviceStatus.value = status
    serviceAddress.value = configStore.config.value.ollamaHost || status.host
    serviceVersion.value = status.version
  } catch (error) {
    console.error('获取服务状态失败:', error)
    serviceAddress.value = configStore.config.value.ollamaHost
  }
}

const saveServiceAddress = () => {
  if (!serviceAddress.value) {
    ElMessage.warning('请输入服务地址')
    return
  }
  
  const addressPattern = /^[\w.-]+:\d+$/
  if (!addressPattern.test(serviceAddress.value)) {
    ElMessage.warning('服务地址格式不正确，例如: 127.0.0.1:11434')
    return
  }
  
  configStore.updateConfig({ ollamaHost: serviceAddress.value })
  ElMessage.success('服务地址已保存，刷新页面后生效')
}

const loadHardwareInfo = async () => {
  try {
    const envInfo = await GetEnvironmentInfo()
    const intelInfo = await GetIntelOptimizationInfo()
    
    hardwareInfo.value = {
      cpu: envInfo.cpu_info || 'Intel CPU',
      gpu: envInfo.gpu_status || 'Intel GPU',
      memory: envInfo.memory_usage || 'System Memory',
      instructions: intelInfo.supported_devices ? intelInfo.supported_devices.join(', ') : 'Intel Hardware'
    }
  } catch (error) {
    console.error('加载硬件信息失败:', error)
  }
}

const openWebsite = () => {
  window.open('https://github.com/jianggujin/ollama-desktop', '_blank')
}

const openDocumentation = () => {
  window.open('https://www.modelscope.cn/models/Intel/ollama/summary', '_blank')
}

const checkForUpdates = () => {
  ElMessage.info('正在检查更新...')
  setTimeout(() => {
    ElMessage.success('已是最新版本')
  }, 1000)
}

const saveEnvironmentVariables = async () => {
  try {
    const result = await SaveEnvironmentVariables(environmentVariables.value)
    ElMessage.success('环境变量配置已保存')
    if (result && result.path) {
      currentOllamaPath.value = result.path
    }
    ElMessage.warning('配置已保存，需要重启服务才能生效')
  } catch (error) {
    console.error('保存环境变量失败:', error)
    ElMessage.error('保存环境变量失败')
  }
}

const loadEnvironmentVariables = async () => {
  try {
    const envVars = await GetEnvironmentVariables()
    if (envVars) {
      environmentVariables.value = { ...environmentVariables.value, ...envVars }
    }
  } catch (error) {
    console.error('加载环境变量失败:', error)
  }
}

const loadCurrentOllamaPath = async () => {
  try {
    const result = await GetOllamaPath()
    if (result && result.path) {
      currentOllamaPath.value = result.path
    }
  } catch (error) {
    console.error('加载 Ollama 路径失败:', error)
    currentOllamaPath.value = '无法获取路径'
  }
}

const openaiApiUrl = computed(() => {
  return `http://localhost:11434/v1`
})

const copyApiUrl = async () => {
  try {
    await navigator.clipboard.writeText(openaiApiUrl.value)
    ElMessage.success('API 地址已复制到剪贴板')
  } catch (err) {
    console.error('复制失败:', err)
    ElMessage.error('复制失败')
  }
}

onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark') {
    themeMode.value = true
    document.body.classList.add('dark-theme')
  }
  
  const savedLang = localStorage.getItem('language')
  if (savedLang) {
    language.value = savedLang
  }
  
  loadServiceStatus()
  loadHardwareInfo()
  loadEnvironmentVariables()
  loadCurrentOllamaPath()
})
</script>

<style scoped>
.tech-settings-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(135deg, #0a0a0f 0%, #1a1a2e 50%, #0f0f1a 100%);
  position: relative;
  overflow: hidden;
}

.tech-settings-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: 
    radial-gradient(circle at 20% 80%, rgba(6, 182, 212, 0.05) 0%, transparent 50%),
    radial-gradient(circle at 80% 20%, rgba(139, 92, 246, 0.05) 0%, transparent 50%);
  pointer-events: none;
}

.tech-page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: rgba(15, 15, 25, 0.8);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(6, 182, 212, 0.2);
  position: relative;
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-icon {
  width: 42px;
  height: 42px;
  background: linear-gradient(135deg, #06b6d4 0%, #8b5cf6 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
  box-shadow: 0 4px 15px rgba(6, 182, 212, 0.3);
}

.title-text h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #e2e8f0;
  letter-spacing: 0.5px;
}

.title-text .subtitle {
  font-size: 11px;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.settings-content {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
  z-index: 1;
}

.settings-sidebar {
  width: 200px;
  padding: 20px 12px;
  background: rgba(15, 15, 25, 0.6);
  border-right: 1px solid rgba(6, 182, 212, 0.1);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  color: #94a3b8;
  font-size: 14px;
}

.sidebar-item:hover {
  background: rgba(30, 41, 59, 0.6);
  color: #e2e8f0;
}

.sidebar-item.active {
  background: rgba(6, 182, 212, 0.15);
  color: #06b6d4;
}

.sidebar-item .el-icon {
  font-size: 18px;
}

.settings-main {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.settings-main::-webkit-scrollbar {
  width: 6px;
}

.settings-main::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.5);
}

.settings-main::-webkit-scrollbar-thumb {
  background: rgba(6, 182, 212, 0.3);
  border-radius: 3px;
}

.settings-panel {
  max-width: 800px;
}

.panel-header {
  margin-bottom: 24px;
}

.panel-header h3 {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
  color: #e2e8f0;
}

.panel-header span {
  font-size: 12px;
  color: #64748b;
}

.panel-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.panel-section {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid rgba(51, 65, 85, 0.5);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.section-header h4 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #e2e8f0;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: rgba(30, 41, 59, 0.4);
  border-radius: 12px;
  border: 1px solid rgba(51, 65, 85, 0.3);
}

.setting-info {
  flex: 1;
  min-width: 0;
}

.setting-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #e2e8f0;
  margin-bottom: 4px;
}

.setting-desc {
  font-size: 12px;
  color: #64748b;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 12px;
}

.setting-control.full-width {
  flex-direction: column;
  align-items: flex-start;
  flex: 1;
  margin-left: 20px;
}

.tech-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  border: none;
  background: rgba(30, 41, 59, 0.8);
  color: #94a3b8;
  border: 1px solid rgba(100, 116, 139, 0.3);
}

.tech-btn:hover {
  background: rgba(51, 65, 85, 0.8);
  border-color: rgba(6, 182, 212, 0.5);
  color: #e2e8f0;
}

.tech-btn-primary {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  border: none;
  color: white;
}

.tech-btn-primary:hover {
  background: linear-gradient(135deg, #22d3ee 0%, #06b6d4 100%);
  box-shadow: 0 4px 15px rgba(6, 182, 212, 0.4);
}

.tech-btn-danger {
  background: transparent;
  border: 1px solid rgba(239, 68, 68, 0.5);
  color: #ef4444;
}

.tech-btn-danger:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: #ef4444;
}

.tech-btn-success {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  border: none;
  color: white;
}

.tech-btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.tech-switch :deep(.el-switch__core) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(100, 116, 139, 0.3);
}

.tech-switch.is-checked :deep(.el-switch__core) {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  border-color: #06b6d4;
}

.tech-select {
  width: 200px;
}

.tech-select :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 8px;
  box-shadow: none;
}

.tech-select :deep(.el-input__inner) {
  color: #e2e8f0;
}

.tech-input {
  width: 200px;
}

.tech-input :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 8px;
  box-shadow: none;
}

.tech-input :deep(.el-input__inner) {
  color: #e2e8f0;
}

.setting-control.full-width .tech-input {
  width: 100%;
}

.tech-number :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 8px;
  box-shadow: none;
}

.tech-number :deep(.el-input__inner) {
  color: #e2e8f0;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.running {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.status-badge.stopped {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.path-hint {
  font-size: 11px;
  color: #10b981;
  margin-top: 4px;
}

.api-url-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 8px;
  border: 1px solid rgba(6, 182, 212, 0.2);
}

.api-url {
  font-size: 13px;
  color: #06b6d4;
  font-family: monospace;
}

.hardware-info {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  width: 100%;
}

.hw-item {
  display: flex;
  flex-direction: column;
  padding: 12px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 8px;
  border: 1px solid rgba(51, 65, 85, 0.3);
}

.hw-label {
  font-size: 11px;
  color: #64748b;
  margin-bottom: 4px;
  text-transform: uppercase;
}

.hw-value {
  font-size: 13px;
  color: #e2e8f0;
}

.about-section {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.about-logo {
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo-icon {
  width: 56px;
  height: 56px;
  background: linear-gradient(135deg, #06b6d4 0%, #8b5cf6 100%);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 28px;
}

.logo-info h4 {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
  color: #e2e8f0;
}

.logo-info span {
  font-size: 12px;
  color: #64748b;
}

.about-info {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 12px;
  background: rgba(30, 41, 59, 0.4);
  border-radius: 8px;
}

.info-label {
  font-size: 13px;
  color: #64748b;
}

.info-value {
  font-size: 13px;
  color: #e2e8f0;
  font-weight: 500;
}

.about-desc {
  line-height: 1.8;
}

.about-desc p {
  margin: 0 0 12px 0;
  font-size: 13px;
  color: #94a3b8;
}

.about-actions {
  display: flex;
  gap: 12px;
}
</style>
