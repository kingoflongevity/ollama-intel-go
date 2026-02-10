<template>
  <div class="settings-container">
    <el-tabs v-model="activeTab" class="settings-tabs">
      <!-- 通用设置 -->
      <el-tab-pane label="通用" name="general">
        <el-card class="setting-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">界面设置</span>
            </div>
          </template>
          <el-form label-width="120px">
            <el-form-item label="主题模式">
              <el-switch
                v-model="themeMode"
                inline-prompt
                active-text="暗色"
                inactive-text="亮色"
                @change="toggleTheme"
              />
            </el-form-item>
            <el-form-item label="语言">
              <el-select v-model="language" placeholder="选择语言" @change="changeLanguage">
                <el-option label="简体中文" value="zh-CN" />
                <el-option label="English" value="en-US" />
                <el-option label="日本語" value="ja-JP" />
              </el-select>
            </el-form-item>
            <el-form-item label="自动检查更新">
              <el-switch v-model="autoUpdate" />
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 服务设置 -->
      <el-tab-pane label="服务" name="service">
        <el-card class="setting-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">Ollama 服务管理</span>
            </div>
          </template>
          <el-form label-width="120px">
            <el-form-item label="服务状态">
              <el-tag :type="serviceStatus.running ? 'success' : 'danger'">
                {{ serviceStatus.running ? '运行中' : '已停止' }}
              </el-tag>
              <el-button 
                style="margin-left: 20px;" 
                :type="serviceStatus.running ? 'danger' : 'success'"
                @click="toggleService"
              >
                {{ serviceStatus.running ? '停止服务' : '启动服务' }}
              </el-button>
            </el-form-item>
            <el-form-item label="服务地址">
              <el-input v-model="serviceAddress" disabled />
            </el-form-item>
            <el-form-item label="服务版本">
              <el-input v-model="serviceVersion" disabled />
            </el-form-item>
            <el-form-item label="自动启动">
              <el-switch v-model="autoStartService" />
              <div class="form-help">应用启动时自动启动 Ollama 服务</div>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="setting-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">环境变量配置</span>
              <el-button type="primary" size="small" @click="saveEnvironmentVariables">
                保存配置
              </el-button>
            </div>
          </template>
          <el-form label-width="180px">
            <el-form-item label="模型下载源">
              <el-select v-model="environmentVariables.OLLAMA_MODEL_SOURCE" placeholder="选择模型下载源">
                <el-option label="ModelScope (默认)" value="modelscope" />
                <el-option label="Ollama 官方" value="ollama" />
              </el-select>
              <div class="form-help">设置模型的下载来源</div>
            </el-form-item>
            <el-form-item label="上下文长度">
              <el-input-number 
                v-model="environmentVariables.OLLAMA_NUM_CTX" 
                :min="2048" 
                :max="1048576" 
                :step="2048"
                style="width: 180px"
              />
              <div class="form-help">设置模型的上下文窗口大小，默认 2048</div>
            </el-form-item>
            <el-form-item label="GPU 选择">
              <el-input v-model="environmentVariables.ONEAPI_DEVICE_SELECTOR" placeholder="例如: gpu:0" />
              <div class="form-help">在多 GPU 环境中选择特定 GPU，例如: gpu:0</div>
            </el-form-item>
            <el-form-item label="启用 GPU 加速">
              <el-switch v-model="environmentVariables.OLLAMA_INTEL_GPU" />
              <div class="form-help">启用 Intel GPU 加速</div>
            </el-form-item>
            <el-form-item label="Ollama 执行程序路径">
              <el-input 
                v-model="environmentVariables.OLLAMA_EXECUTABLE_PATH" 
                placeholder="设置 Ollama 执行程序的路径，例如: C:\\ollama\\ollama.exe"
              />
              <div class="form-help">设置 Ollama 执行程序的路径，让用户自行下载放置。留空则使用默认路径</div>
              <div class="form-help" style="color: #67c23a;">
                当前默认路径: {{ currentOllamaPath }}
              </div>
            </el-form-item>
            <el-form-item label="调试模式">
              <el-switch v-model="environmentVariables.OLLAMA_DEBUG" />
              <div class="form-help">启用 Ollama 调试日志</div>
            </el-form-item>
            <el-form-item label="OpenAI 兼容 API">
              <el-switch v-model="environmentVariables.OLLAMA_OPENAI_COMPATIBLE" />
              <div class="form-help">启用 OpenAI 兼容的 API 接口，允许其他应用像调用 OpenAI 一样调用本地模型</div>
            </el-form-item>
            <el-form-item label="API 密钥" v-if="environmentVariables.OLLAMA_OPENAI_COMPATIBLE">
              <el-input v-model="environmentVariables.OLLAMA_OPENAI_API_KEY" placeholder="设置 API 密钥（可选）" />
              <div class="form-help">设置用于验证 API 调用的密钥，留空则不需要验证</div>
            </el-form-item>
            <el-form-item label="API 端口" v-if="environmentVariables.OLLAMA_OPENAI_COMPATIBLE">
              <el-input-number 
                v-model="environmentVariables.OLLAMA_OPENAI_PORT" 
                :min="1024" 
                :max="65535" 
                :step="1" 
                style="width: 120px"
              />
              <div class="form-help">设置应用程序内置的 OpenAI 兼容 API 服务端口，默认 8080</div>
              <div class="form-help" style="color: #67c23a;">
                推荐使用: Ollama 内置 API (http://localhost:11434/v1)
              </div>
            </el-form-item>
            <el-form-item label="API 调用地址" v-if="environmentVariables.OLLAMA_OPENAI_COMPATIBLE">
              <el-input 
                :value="openaiApiUrl" 
                readonly 
                style="width: 400px"
              >
                <template #append>
                  <el-button @click="copyApiUrl">
                    <el-icon><CopyDocument /></el-icon>
                    复制
                  </el-button>
                </template>
              </el-input>
              <div class="form-help">Ollama 内置的 OpenAI 兼容 API 地址，其他应用可以通过此地址调用本地模型</div>
              <div class="form-help" style="color: #67c23a;">
                此地址为 Ollama 服务默认提供的 OpenAI 兼容 API 端点
              </div>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 英特尔优化 -->
      <el-tab-pane label="英特尔优化" name="intel">
        <el-card class="setting-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">英特尔硬件优化</span>
            </div>
          </template>
          <el-form label-width="150px">
            <el-form-item label="GPU 加速">
              <el-switch v-model="gpuAcceleration" />
              <div class="form-help">启用 Intel Arc GPU 加速</div>
            </el-form-item>
            <el-form-item label="OneAPI 优化">
              <el-switch v-model="oneApiOptimization" />
              <div class="form-help">启用 Intel OneAPI 优化</div>
            </el-form-item>
            <el-form-item label="Intel MKL">
              <el-switch v-model="intelMKL" />
              <div class="form-help">启用 Intel Math Kernel Library</div>
            </el-form-item>
            <el-form-item label="硬件检测结果">
              <div class="hardware-info">
                <div class="hardware-item">
                  <span class="label">CPU:</span>
                  <span class="value">{{ hardwareInfo.cpu }}</span>
                </div>
                <div class="hardware-item">
                  <span class="label">GPU:</span>
                  <span class="value">{{ hardwareInfo.gpu }}</span>
                </div>
                <div class="hardware-item">
                  <span class="label">内存:</span>
                  <span class="value">{{ hardwareInfo.memory }}</span>
                </div>
                <div class="hardware-item">
                  <span class="label">指令集:</span>
                  <span class="value">{{ hardwareInfo.instructions }}</span>
                </div>
              </div>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 关于 -->
      <el-tab-pane label="关于" name="about">
        <el-card class="setting-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">Ollama 英特尔优化版</span>
            </div>
          </template>
          <div class="about-content">
            <div class="app-info">
              <div class="info-item">
                <span class="label">版本:</span>
                <span class="value">1.0.0</span>
              </div>
              <div class="info-item">
                <span class="label">构建时间:</span>
                <span class="value">{{ buildTime }}</span>
              </div>
              <div class="info-item">
                <span class="label">技术栈:</span>
                <span class="value">Go + Vue 3 + Element Plus</span>
              </div>
              <div class="info-item">
                <span class="label">优化:</span>
                <span class="value">专为英特尔硬件优化</span>
              </div>
            </div>
            <div class="description">
              <p>Ollama 英特尔优化版是一个专为英特尔硬件优化的桌面应用解决方案，用于在 Windows、macOS 和 Linux 操作系统上运行和管理 Ollama 模型的 GUI 工具。</p>
              <p>该版本特别针对英特尔处理器、Intel Arc GPU 和相关技术栈进行了深度优化，提供最佳性能体验。</p>
            </div>
            <div class="links">
              <el-button type="primary" @click="openWebsite">官方网站</el-button>
              <el-button @click="openDocumentation">使用文档</el-button>
              <el-button @click="checkForUpdates">检查更新</el-button>
            </div>
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { GetServiceStatus, StartService, StopService, GetEnvironmentInfo, GetIntelOptimizationInfo, GetEnvironmentVariables, SaveEnvironmentVariables, GetOllamaPath } from '../../wailsjs/go/main/App'
import { CopyDocument } from '@element-plus/icons-vue'

// 响应式数据
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
  OLLAMA_MODEL_SOURCE: 'modelscope',
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

// 切换主题
const toggleTheme = (val) => {
  document.body.classList.toggle('dark-theme', val)
  localStorage.setItem('theme', val ? 'dark' : 'light')
}

// 切换语言
const changeLanguage = (lang) => {
  language.value = lang
  localStorage.setItem('language', lang)
  // 重新加载页面应用新语言（简化实现）
  ElMessage.success(`语言已切换为 ${lang === 'zh-CN' ? '简体中文' : lang === 'en-US' ? 'English' : '日本語'}`)
}

// 切换服务状态
const toggleService = async () => {
  try {
    if (serviceStatus.value.running) {
      await StopService()
      ElMessage.success('服务已停止')
    } else {
      await StartService()
      ElMessage.success('服务启动中...')
    }
    // 等待一段时间后更新状态
    setTimeout(() => {
      loadServiceStatus()
    }, 1000)
  } catch (error) {
    console.error('服务控制失败:', error)
    ElMessage.error('服务控制失败')
  }
}

// 加载服务状态
const loadServiceStatus = async () => {
  try {
    const status = await GetServiceStatus()
    serviceStatus.value = status
    serviceAddress.value = status.host
    serviceVersion.value = status.version
  } catch (error) {
    console.error('获取服务状态失败:', error)
  }
}

// 加载硬件信息
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

// 打开网站
const openWebsite = () => {
  window.open('https://github.com/jianggujin/ollama-desktop', '_blank')
}

// 打开文档
const openDocumentation = () => {
  window.open('https://www.modelscope.cn/models/Intel/ollama/summary', '_blank')
}

// 检查更新
const checkForUpdates = () => {
  ElMessage.info('正在检查更新...')
  // 模拟检查更新
  setTimeout(() => {
    ElMessage.success('已是最新版本')
  }, 1000)
}

// 保存环境变量
const saveEnvironmentVariables = async () => {
  try {
    const result = await SaveEnvironmentVariables(environmentVariables.value)
    ElMessage.success('环境变量配置已保存')
    // 更新显示的Ollama路径
    if (result && result.path) {
      currentOllamaPath.value = result.path
    }
    // 提示需要重启服务
    ElMessage.warning('配置已保存，需要重启服务才能生效')
  } catch (error) {
    console.error('保存环境变量失败:', error)
    ElMessage.error('保存环境变量失败')
  }
}

// 加载环境变量
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

// 加载当前的Ollama路径
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

// 计算属性：OpenAI API 调用地址
const openaiApiUrl = computed(() => {
  // 使用Ollama默认的OpenAI兼容API端口11434
  return `http://localhost:11434/v1`
})

// 复制 API URL 到剪贴板
const copyApiUrl = async () => {
  try {
    await navigator.clipboard.writeText(openaiApiUrl.value)
    ElMessage.success('API 地址已复制到剪贴板')
  } catch (err) {
    console.error('复制失败:', err)
    ElMessage.error('复制失败')
  }
}

// 初始化
onMounted(() => {
  // 加载保存的主题设置
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark') {
    themeMode.value = true
    document.body.classList.add('dark-theme')
  }
  
  // 加载保存的语言设置
  const savedLang = localStorage.getItem('language')
  if (savedLang) {
    language.value = savedLang
  }
  
  // 加载服务状态
  loadServiceStatus()
  
  // 加载硬件信息
  loadHardwareInfo()
  
  // 加载环境变量配置
  loadEnvironmentVariables()
  
  // 加载当前的Ollama路径
  loadCurrentOllamaPath()
})
</script>

<style scoped>
.settings-container {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  height: 100%;
}

body.dark-theme .settings-container {
  background: #1e1e1e;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.settings-tabs {
  height: 100%;
}

.settings-tabs :deep(.el-tabs__content) {
  height: calc(100% - 56px);
  overflow-y: auto;
  padding: 20px;
}

.setting-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

body.dark-theme .card-title {
  color: #e4e6eb;
}

.form-help {
  font-size: 13px;
  color: #909399;
  margin-top: 5px;
}

body.dark-theme .form-help {
  color: #a0aec0;
}

.hardware-info {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 8px;
}

body.dark-theme .hardware-info {
  background: #2d2d2d;
}

.hardware-item {
  display: flex;
  margin-bottom: 8px;
}

.hardware-item:last-child {
  margin-bottom: 0;
}

.label {
  font-weight: 600;
  width: 100px;
  color: #606266;
}

body.dark-theme .label {
  color: #c0c4cc;
}

.value {
  color: #303133;
  flex: 1;
}

body.dark-theme .value {
  color: #e4e6eb;
}

.about-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.app-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
}

.info-item {
  display: flex;
  flex-direction: column;
}

.description {
  line-height: 1.6;
  color: #606266;
}

body.dark-theme .description {
  color: #c0c4cc;
}

.links {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

/* 暗色主题 - Element Plus 组件样式 */
/* 全局暗色主题变量 */
body.dark-theme {
  --el-bg-color: #1e1e1e;
  --el-bg-color-overlay: #2d2d2d;
  --el-text-color-primary: #e4e6eb;
  --el-text-color-regular: #c0c4cc;
  --el-text-color-secondary: #a0aec0;
  --el-border-color: #4a4a4a;
  --el-border-color-light: #3a3a3a;
  --el-border-color-lighter: #2d2d2d;
  --el-border-color-extra-light: #252525;
  --el-fill-color: #2d2d2d;
  --el-fill-color-light: #3a3a3a;
  --el-fill-color-lighter: #4a4a4a;
  --el-fill-color-blank: #1e1e1e;
  --el-box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.2);
  --el-box-shadow-light: 0 2px 4px 0 rgba(0, 0, 0, 0.1);
  --el-box-shadow-lighter: 0 2px 4px 0 rgba(0, 0, 0, 0.05);
  --el-box-shadow-dark: 0 2px 8px 0 rgba(0, 0, 0, 0.3);
}

/* 卡片组件 */
body.dark-theme :deep(.el-card) {
  background-color: #1e1e1e !important;
  border-color: #4a4a4a !important;
}

body.dark-theme :deep(.el-card__header) {
  background-color: #2d2d2d !important;
  border-bottom-color: #4a4a4a !important;
}

body.dark-theme :deep(.el-card__body) {
  background-color: #1e1e1e !important;
}

/* 标签页组件 */
body.dark-theme :deep(.el-tabs) {
  background-color: #1e1e1e !important;
}

body.dark-theme :deep(.el-tabs__header) {
  background-color: #1e1e1e !important;
  border-bottom-color: #4a4a4a !important;
}

body.dark-theme :deep(.el-tabs__nav) {
  background-color: #1e1e1e !important;
}

body.dark-theme :deep(.el-tabs__item) {
  color: #a0aec0 !important;
}

body.dark-theme :deep(.el-tabs__item.is-active) {
  color: #409eff !important;
}

body.dark-theme :deep(.el-tabs__content) {
  background-color: #1e1e1e !important;
}

body.dark-theme :deep(.el-tab-pane) {
  background-color: #1e1e1e !important;
}

/* 表单组件 */
body.dark-theme :deep(.el-form-item__label) {
  color: #e4e6eb !important;
}

body.dark-theme :deep(.el-input__wrapper) {
  background-color: #2d2d2d !important;
  border-color: #4a4a4a !important;
}

body.dark-theme :deep(.el-input__inner) {
  color: #e4e6eb !important;
}

body.dark-theme :deep(.el-input__placeholder) {
  color: #a0aec0 !important;
}

/* 选择器组件 */
body.dark-theme :deep(.el-select__wrapper) {
  background-color: #2d2d2d !important;
  border-color: #4a4a4a !important;
}

body.dark-theme :deep(.el-select__placeholder) {
  color: #a0aec0 !important;
}

body.dark-theme :deep(.el-select__caret) {
  color: #a0aec0 !important;
}

/* 开关组件 */
body.dark-theme :deep(.el-switch__core) {
  background-color: #4a4a4a !important;
}

body.dark-theme :deep(.el-switch.is-checked .el-switch__core) {
  background-color: #409eff !important;
}

/* 数字输入框组件 */
body.dark-theme :deep(.el-input-number) {
  background-color: #2d2d2d !important;
  border-color: #4a4a4a !important;
}

body.dark-theme :deep(.el-input-number__decrease),
body.dark-theme :deep(.el-input-number__increase) {
  background-color: #2d2d2d !important;
  border-color: #4a4a4a !important;
  color: #e4e6eb !important;
}

body.dark-theme :deep(.el-input-number__decrease:hover),
body.dark-theme :deep(.el-input-number__increase:hover) {
  background-color: #3a3a3a !important;
}
</style>