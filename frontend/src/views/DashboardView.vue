<template>
  <div class="dashboard-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-title">
        <div class="title-icon">
          <el-icon size="24"><Odometer /></el-icon>
        </div>
        <div class="title-content">
          <h2>系统仪表盘</h2>
          <p class="subtitle">实时监控和管理您的AI模型服务</p>
        </div>
      </div>
      <div class="header-time">
        <span class="time-value">{{ currentTime }}</span>
        <span class="date-value">{{ currentDate }}</span>
      </div>
    </div>

    <!-- 状态卡片 -->
    <div class="status-grid">
      <div class="status-card" :class="{ active: serviceStatusOk }">
        <div class="status-icon">
          <el-icon size="24"><Connection /></el-icon>
        </div>
        <div class="status-info">
          <span class="status-label">服务状态</span>
          <span class="status-value">{{ serviceStatusText }}</span>
        </div>
        <div class="status-indicator" :class="serviceStatusOk ? 'online' : 'offline'"></div>
      </div>
      
      <div class="status-card" :class="{ active: gpuStatusOk }">
        <div class="status-icon">
          <el-icon size="24"><Cpu /></el-icon>
        </div>
        <div class="status-info">
          <span class="status-label">GPU 状态</span>
          <span class="status-value">{{ gpuStatusText }}</span>
        </div>
        <div class="status-indicator" :class="gpuStatusOk ? 'online' : 'warning'"></div>
      </div>
      
      <div class="status-card">
        <div class="status-icon">
          <el-icon size="24"><Memo /></el-icon>
        </div>
        <div class="status-info">
          <span class="status-label">内存使用</span>
          <span class="status-value">{{ environmentInfo.memory_usage || '检测中...' }}</span>
        </div>
        <div class="status-indicator"></div>
      </div>
      
      <div class="status-card">
        <div class="status-icon">
          <el-icon size="24"><InfoFilled /></el-icon>
        </div>
        <div class="status-info">
          <span class="status-label">Ollama 版本</span>
          <span class="status-value">{{ environmentInfo.ollama_version || 'unknown' }}</span>
        </div>
        <div class="status-indicator"></div>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="main-content">
      <!-- 左侧面板 -->
      <div class="left-panel">
        <!-- 功能快捷入口 -->
        <div class="section-card">
          <div class="section-header">
            <h3><el-icon><Grid /></el-icon> 快捷功能</h3>
          </div>
          <div class="feature-grid">
            <div class="feature-item" @click="navigateTo('/chat')">
              <div class="feature-icon chat">
                <el-icon size="24"><ChatDotRound /></el-icon>
              </div>
              <div class="feature-text">
                <span class="feature-title">AI 对话</span>
                <span class="feature-desc">智能对话交互</span>
              </div>
              <el-icon class="feature-arrow"><ArrowRight /></el-icon>
            </div>
            
            <div class="feature-item" @click="navigateTo('/models')">
              <div class="feature-icon models">
                <el-icon size="24"><Box /></el-icon>
              </div>
              <div class="feature-text">
                <span class="feature-title">模型管理</span>
                <span class="feature-desc">本地模型管理</span>
              </div>
              <el-icon class="feature-arrow"><ArrowRight /></el-icon>
            </div>
            
            <div class="feature-item" @click="navigateTo('/online')">
              <div class="feature-icon online">
                <el-icon size="24"><Download /></el-icon>
              </div>
              <div class="feature-text">
                <span class="feature-title">在线模型</span>
                <span class="feature-desc">浏览下载模型</span>
              </div>
              <el-icon class="feature-arrow"><ArrowRight /></el-icon>
            </div>
            
            <div class="feature-item" @click="navigateTo('/settings')">
              <div class="feature-icon settings">
                <el-icon size="24"><Setting /></el-icon>
              </div>
              <div class="feature-text">
                <span class="feature-title">系统设置</span>
                <span class="feature-desc">环境参数配置</span>
              </div>
              <el-icon class="feature-arrow"><ArrowRight /></el-icon>
            </div>
          </div>
        </div>

        <!-- 控制面板 -->
        <div class="section-card">
          <div class="section-header">
            <h3><el-icon><SwitchButton /></el-icon> 服务控制</h3>
          </div>
          <div class="control-panel">
            <div class="control-buttons">
              <button class="control-btn start" @click="startService" :disabled="serviceStatus.running">
                <el-icon><VideoPlay /></el-icon>
                启动服务
              </button>
              <button class="control-btn stop" @click="stopService" :disabled="!serviceStatus.running">
                <el-icon><VideoPause /></el-icon>
                停止服务
              </button>
              <button class="control-btn refresh" @click="refreshStatus">
                <el-icon><Refresh /></el-icon>
                刷新状态
              </button>
            </div>
            <div class="service-info" v-if="serviceStatus.running">
              <div class="info-item">
                <span class="info-label">服务地址</span>
                <span class="info-value">{{ serviceStatus.host }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">运行版本</span>
                <span class="info-value">{{ serviceStatus.version }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 最近活动 -->
        <div class="section-card">
          <div class="section-header">
            <h3><el-icon><Clock /></el-icon> 最近活动</h3>
          </div>
          <div class="activity-list">
            <div v-if="recentChats.length === 0" class="empty-activity">
              <el-icon size="32"><Document /></el-icon>
              <span>暂无最近活动</span>
            </div>
            <div v-for="(chat, index) in recentChats" :key="chat.id" class="activity-item">
              <div class="activity-index">{{ index + 1 }}</div>
              <div class="activity-content">
                <span class="activity-text">{{ chat.content }}</span>
                <span class="activity-time">{{ formatTime(chat.time) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧面板 -->
      <div class="right-panel">
        <div class="section-card monitor-card">
          <div class="section-header">
            <h3><el-icon><DataLine /></el-icon> 实时监控</h3>
            <span class="monitor-status" :class="monitorActive ? 'active' : ''">
              <span class="status-dot"></span>
              {{ monitorActive ? '监控中' : '已暂停' }}
            </span>
          </div>
          <div class="monitor-content">
            <!-- CPU 使用率 -->
            <div class="monitor-item">
              <div class="monitor-header">
                <span class="monitor-label">CPU 使用率</span>
                <span class="monitor-value">{{ cpuUsage }}%</span>
              </div>
              <div class="monitor-bar">
                <div class="bar-fill cpu" :style="{ width: cpuUsage + '%' }"></div>
              </div>
            </div>
            
            <!-- 内存使用率 -->
            <div class="monitor-item">
              <div class="monitor-header">
                <span class="monitor-label">内存使用</span>
                <span class="monitor-value">{{ memoryUsage }} GB / {{ memoryTotal }} GB</span>
              </div>
              <div class="monitor-bar">
                <div class="bar-fill memory" :style="{ width: memoryPercent + '%' }"></div>
              </div>
            </div>
            
            <!-- GPU 状态 -->
            <div class="monitor-item" v-if="gpuInfo.available">
              <div class="monitor-header">
                <span class="monitor-label">GPU 显存</span>
                <span class="monitor-value">{{ gpuInfo.used }} / {{ gpuInfo.total }}</span>
              </div>
              <div class="monitor-bar">
                <div class="bar-fill gpu" :style="{ width: gpuInfo.percent + '%' }"></div>
              </div>
            </div>
            
            <!-- 服务状态列表 -->
            <div class="service-status-list">
              <div class="service-status-item">
                <span class="service-name">Ollama API</span>
                <span class="service-badge" :class="serviceStatus.running ? 'online' : 'offline'">
                  {{ serviceStatus.running ? '在线' : '离线' }}
                </span>
              </div>
              <div class="service-status-item">
                <span class="service-name">OpenAI 兼容</span>
                <span class="service-badge online">可用</span>
              </div>
              <div class="service-status-item">
                <span class="service-name">WebSocket</span>
                <span class="service-badge" :class="wsConnected ? 'online' : 'offline'">
                  {{ wsConnected ? '已连接' : '未连接' }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  Odometer, Connection, Cpu, Memo, InfoFilled, Grid, ArrowRight,
  ChatDotRound, Box, Download, Setting, SwitchButton, VideoPlay, 
  VideoPause, Refresh, Clock, Document, DataLine
} from '@element-plus/icons-vue'
import { GetEnvironmentInfo, GetServiceStatus, StartService, StopService, GetRealTimeStats } from '../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'

const router = useRouter()

const environmentInfo = ref({})
const serviceStatus = ref({})
const recentChats = ref([])
const currentTime = ref('')
const currentDate = ref('')
const monitorActive = ref(true)
const wsConnected = ref(false)

// 监控数据
const cpuUsage = ref(0)
const memoryUsage = ref(0)
const memoryTotal = ref(0)
const memoryPercent = ref(0)
const gpuInfo = ref({ available: false, used: '0 GB', total: '0 GB', percent: 0 })

/**
 * GPU状态文本
 */
const gpuStatusText = computed(() => {
  const status = environmentInfo.value.gpu_status || ''
  if (status.includes('Available') || status.includes('Running')) {
    return 'GPU 可用'
  }
  return status || '检测中...'
})

/**
 * GPU状态是否正常
 */
const gpuStatusOk = computed(() => {
  const status = environmentInfo.value.gpu_status || ''
  return status.includes('Available') || status.includes('Running') || status.includes('Detected')
})

/**
 * 服务状态文本
 */
const serviceStatusText = computed(() => {
  return serviceStatus.value.running ? '运行中' : '已停止'
})

/**
 * 服务状态是否正常
 */
const serviceStatusOk = computed(() => {
  return serviceStatus.value.running === true
})

/**
 * 导航到指定路径
 */
const navigateTo = (path) => {
  router.push(path)
}

/**
 * 更新时间
 */
const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  currentDate.value = now.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' })
}

/**
 * 格式化时间
 */
const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

/**
 * 加载环境信息
 */
const loadEnvironmentInfo = async () => {
  try {
    const info = await GetEnvironmentInfo()
    environmentInfo.value = info
  } catch (error) {
    console.error('加载环境信息失败:', error)
  }
}

/**
 * 加载服务状态
 */
const loadServiceStatus = async () => {
  try {
    const status = await GetServiceStatus()
    serviceStatus.value = status
  } catch (error) {
    console.error('加载服务状态失败:', error)
  }
}

/**
 * 加载监控数据
 */
const loadMonitorData = async () => {
  try {
    const stats = await GetRealTimeStats()
    if (stats.cpu) {
      cpuUsage.value = Math.round(stats.cpu.usage_percent || 0)
    }
    if (stats.memory) {
      memoryUsage.value = (stats.memory.used_gb || 0).toFixed(1)
      memoryTotal.value = (stats.memory.total_gb || 0).toFixed(1)
      memoryPercent.value = Math.round(stats.memory.used_percent || 0)
    }
    if (stats.gpu) {
      gpuInfo.value = {
        available: stats.gpu.available || false,
        used: stats.gpu.used_memory || '0 GB',
        total: stats.gpu.total_memory || '0 GB',
        percent: stats.gpu.memory_percent || 0
      }
    }
  } catch (error) {
    console.error('加载监控数据失败:', error)
  }
}

/**
 * 启动服务
 */
const startService = async () => {
  try {
    await StartService()
    ElMessage.success('服务启动中...')
    setTimeout(loadServiceStatus, 2000)
  } catch (error) {
    ElMessage.error('启动服务失败: ' + error.message)
  }
}

/**
 * 停止服务
 */
const stopService = async () => {
  try {
    await StopService()
    ElMessage.success('服务停止中...')
    setTimeout(loadServiceStatus, 1000)
  } catch (error) {
    ElMessage.error('停止服务失败: ' + error.message)
  }
}

/**
 * 刷新状态
 */
const refreshStatus = async () => {
  await Promise.all([
    loadEnvironmentInfo(),
    loadServiceStatus(),
    loadMonitorData()
  ])
  ElMessage.success('状态已刷新')
}

/**
 * 加载最近会话
 */
const loadRecentChats = () => {
  try {
    const stored = localStorage.getItem('ollama-chat-sessions')
    if (stored) {
      const sessions = JSON.parse(stored)
      recentChats.value = sessions.slice(0, 5).map(s => ({
        id: s.id,
        content: s.preview || s.name,
        time: s.updatedAt
      }))
    }
  } catch (error) {
    console.error('加载最近会话失败:', error)
  }
}

let refreshInterval = null
let timeInterval = null

onMounted(() => {
  updateTime()
  loadEnvironmentInfo()
  loadServiceStatus()
  loadMonitorData()
  loadRecentChats()
  
  timeInterval = setInterval(updateTime, 1000)
  refreshInterval = setInterval(() => {
    loadEnvironmentInfo()
    loadServiceStatus()
    loadMonitorData()
  }, 5000)
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
  if (timeInterval) clearInterval(timeInterval)
})
</script>

<style scoped>
.dashboard-page {
  padding: 24px;
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0f1a 0%, #1a1f2e 100%);
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 16px;
}

.title-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #06b6d4 0%, #3b82f6 100%);
  border-radius: 12px;
  color: white;
}

.title-content h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #fff;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
}

.header-time {
  text-align: right;
}

.time-value {
  display: block;
  font-size: 28px;
  font-weight: 700;
  color: #06b6d4;
  font-family: monospace;
}

.date-value {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
}

/* 状态卡片 */
.status-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.status-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.status-card:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(6, 182, 212, 0.3);
}

.status-card.active {
  border-color: rgba(16, 185, 129, 0.3);
}

.status-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(6, 182, 212, 0.1);
  border-radius: 12px;
  color: #06b6d4;
}

.status-info {
  flex: 1;
}

.status-label {
  display: block;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 4px;
}

.status-value {
  font-size: 16px;
  font-weight: 600;
  color: #fff;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
}

.status-indicator.online {
  background: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
}

.status-indicator.offline {
  background: #ef4444;
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
}

.status-indicator.warning {
  background: #f59e0b;
  box-shadow: 0 0 8px rgba(245, 158, 11, 0.5);
}

/* 主内容区 */
.main-content {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 24px;
}

.left-panel {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.right-panel {
  display: flex;
  flex-direction: column;
}

/* 区块卡片 */
.section-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.section-header h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-header h3 .el-icon {
  color: #06b6d4;
}

/* 功能网格 */
.feature-grid {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.feature-item:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(6, 182, 212, 0.3);
  transform: translateX(4px);
}

.feature-icon {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  color: white;
}

.feature-icon.chat { background: linear-gradient(135deg, #06b6d4, #0891b2); }
.feature-icon.models { background: linear-gradient(135deg, #8b5cf6, #7c3aed); }
.feature-icon.online { background: linear-gradient(135deg, #10b981, #059669); }
.feature-icon.settings { background: linear-gradient(135deg, #f59e0b, #d97706); }

.feature-text {
  flex: 1;
}

.feature-title {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.feature-desc {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
}

.feature-arrow {
  color: rgba(255, 255, 255, 0.3);
}

/* 控制面板 */
.control-panel {
  padding: 20px;
}

.control-buttons {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.control-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 16px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.control-btn.start {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
}

.control-btn.stop {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: white;
}

.control-btn.refresh {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.control-btn:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

.control-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.service-info {
  display: flex;
  gap: 24px;
  padding: 16px;
  background: rgba(6, 182, 212, 0.05);
  border-radius: 10px;
  border: 1px solid rgba(6, 182, 212, 0.1);
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
}

.info-value {
  font-size: 13px;
  color: #06b6d4;
  font-family: monospace;
}

/* 活动列表 */
.activity-list {
  padding: 16px;
  max-height: 200px;
  overflow-y: auto;
}

.empty-activity {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 32px;
  color: rgba(255, 255, 255, 0.3);
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 10px;
  margin-bottom: 8px;
}

.activity-index {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(6, 182, 212, 0.1);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  color: #06b6d4;
}

.activity-content {
  flex: 1;
  min-width: 0;
}

.activity-text {
  display: block;
  font-size: 13px;
  color: #fff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.activity-time {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.4);
}

/* 监控卡片 */
.monitor-card {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.monitor-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
}

.monitor-status.active {
  color: #10b981;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.monitor-status.active .status-dot {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.monitor-content {
  flex: 1;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.monitor-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.monitor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.monitor-label {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.6);
}

.monitor-value {
  font-size: 13px;
  font-weight: 600;
  color: #06b6d4;
  font-family: monospace;
}

.monitor-bar {
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.5s ease;
}

.bar-fill.cpu { background: linear-gradient(90deg, #06b6d4, #0891b2); }
.bar-fill.memory { background: linear-gradient(90deg, #8b5cf6, #7c3aed); }
.bar-fill.gpu { background: linear-gradient(90deg, #10b981, #059669); }

/* 服务状态列表 */
.service-status-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  margin-top: auto;
}

.service-status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
}

.service-name {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
}

.service-badge {
  font-size: 11px;
  padding: 4px 10px;
  border-radius: 12px;
  font-weight: 500;
}

.service-badge.online {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.service-badge.offline {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

/* 响应式 */
@media (max-width: 1400px) {
  .main-content {
    grid-template-columns: 1fr;
  }
  
  .right-panel {
    min-height: 400px;
  }
}

@media (max-width: 1200px) {
  .status-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .status-grid {
    grid-template-columns: 1fr;
  }
  
  .page-header {
    flex-direction: column;
    text-align: center;
    gap: 16px;
  }
  
  .header-time {
    text-align: center;
  }
}
</style>
