<template>
  <div class="dashboard">
    <!-- 顶部状态栏 -->
    <div class="status-bar">
      <StatusCard title="Ollama 版本" :value="environmentInfo.ollama_version || 'unknown'" icon="Monitor" />
      <StatusCard title="GPU 状态" :value="gpuStatusText" :status="gpuStatusOk ? 'ok' : 'warning'" icon="Cpu" />
      <StatusCard title="内存使用" :value="environmentInfo.memory_usage || 'unknown'" icon="Memo" />
      <StatusCard title="服务状态" :value="serviceStatusText" :status="serviceStatusOk ? 'ok' : 'error'" icon="Connection" />
    </div>

    <!-- 主区域 -->
    <div class="main-grid">
      <div class="left-panel">
        <!-- Ollama 功能卡片 -->
        <div class="feature-cards">
          <div class="feature-card" @click="navigateTo('/chat')">
            <div class="feature-icon chat">
              <el-icon size="28"><ChatDotRound /></el-icon>
            </div>
            <div class="feature-info">
              <h4>AI 对话</h4>
              <p>与AI模型进行智能对话</p>
            </div>
          </div>
          <div class="feature-card" @click="navigateTo('/models')">
            <div class="feature-icon models">
              <el-icon size="28"><Box /></el-icon>
            </div>
            <div class="feature-info">
              <h4>模型管理</h4>
              <p>管理本地已下载的模型</p>
            </div>
          </div>
          <div class="feature-card" @click="navigateTo('/online')">
            <div class="feature-icon online">
              <el-icon size="28"><Download /></el-icon>
            </div>
            <div class="feature-info">
              <h4>在线模型</h4>
              <p>浏览和下载新模型</p>
            </div>
          </div>
          <div class="feature-card" @click="navigateTo('/settings')">
            <div class="feature-icon settings">
              <el-icon size="28"><Setting /></el-icon>
            </div>
            <div class="feature-info">
              <h4>系统设置</h4>
              <p>配置环境和参数</p>
            </div>
          </div>
        </div>

        <!-- 控制面板 -->
        <ControlPanel />
        
        <!-- 活动面板 -->
        <ActivityPanel :recentChats="recentChats" />
      </div>
      
      <div class="right-panel">
        <RealTimeMonitor />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ChatDotRound, Box, Download, Setting } from '@element-plus/icons-vue'
import StatusCard from '@/components/StatusCard.vue'
import ControlPanel from '@/components/ControlPanel.vue'
import ActivityPanel from '@/components/ActivityPanel.vue'
import RealTimeMonitor from '@/components/RealTimeMonitor.vue'
import { GetEnvironmentInfo, GetStats, GetServiceStatus } from '../../wailsjs/go/main/App'

const router = useRouter()

const environmentInfo = ref({})
const stats = ref({})
const serviceStatus = ref({})
const recentChats = ref([])

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
 * 加载统计信息
 */
const loadStats = async () => {
  try {
    const s = await GetStats()
    stats.value = s
  } catch (error) {
    console.error('加载统计信息失败:', error)
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

onMounted(() => {
  loadEnvironmentInfo()
  loadStats()
  loadServiceStatus()
  loadRecentChats()
  
  // 定期刷新
  refreshInterval = setInterval(() => {
    loadEnvironmentInfo()
    loadServiceStatus()
  }, 5000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped>
.dashboard {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
  height: 100%;
  overflow-y: auto;
}

.status-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-lg);
}

.main-grid {
  display: grid;
  grid-template-columns: 1fr 380px;
  gap: var(--spacing-xl);
  flex: 1;
  min-height: 0;
}

.left-panel {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
  flex: 1;
  min-height: 0;
}

.right-panel {
  height: 100%;
  min-height: 500px;
}

/* 功能卡片 */
.feature-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-lg);
}

.feature-card {
  background: var(--bg-elevated);
  border-radius: var(--radius-xl);
  padding: var(--spacing-xl);
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  cursor: pointer;
  transition: all var(--transition-base);
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-sm);
}

.feature-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
  border-color: var(--color-primary);
}

.feature-icon {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.feature-icon.chat {
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
}

.feature-icon.models {
  background: linear-gradient(135deg, var(--color-secondary), var(--color-secondary-dark));
}

.feature-icon.online {
  background: linear-gradient(135deg, var(--color-success), #059669);
}

.feature-icon.settings {
  background: linear-gradient(135deg, var(--color-warning), #d97706);
}

.feature-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.feature-info p {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

/* 响应式布局 */
@media (max-width: 1600px) {
  .main-grid {
    grid-template-columns: 1fr;
  }
  
  .right-panel {
    height: auto;
    min-height: 400px;
  }
}

@media (max-width: 1200px) {
  .feature-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 1400px) {
  .status-bar {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .status-bar {
    grid-template-columns: 1fr;
  }
  
  .feature-cards {
    grid-template-columns: 1fr;
  }
}
</style>
