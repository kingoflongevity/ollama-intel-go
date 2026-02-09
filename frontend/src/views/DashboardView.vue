<template>
  <div class="dashboard">
    <!-- 顶部状态栏 -->
    <div class="status-bar">
      <StatusCard title="Ollama" :value="environmentInfo.ollama_version || 'unknown'" />
      <StatusCard title="GPU" :value="environmentInfo.gpu_status || 'unknown'" :status="environmentInfo.gpu_status && environmentInfo.gpu_status.includes('Available') ? 'ok' : ''" />
      <StatusCard title="Intel GPU" :value="environmentInfo.memory_usage || 'unknown'" />
      <StatusCard title="Service" :value="environmentInfo.service_status || 'unknown'" :status="environmentInfo.service_status === 'running' ? 'ok' : ''" />
    </div>

    <!-- 主区域 -->
    <div class="main-grid">
      <ControlPanel />
      <ActivityPanel :recentChats="recentChats" />
    </div>

    <!-- 监控区域 -->
    <MonitorPanel :stats="stats" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import StatusCard from '../components/StatusCard.vue'
import ControlPanel from '../components/ControlPanel.vue'
import ActivityPanel from '../components/ActivityPanel.vue'
import MonitorPanel from '../components/MonitorPanel.vue'
import { GetEnvironmentInfo, GetStats } from '../../wailsjs/go/main/App'

// 响应式数据
const environmentInfo = ref({})
const stats = ref({})
const recentChats = ref([
  { id: 1, content: '关于 Python 编程的讨论' },
  { id: 2, content: '如何使用 Ollama 模型' },
  { id: 3, content: '系统监控配置' },
  { id: 4, content: '模型拉取问题' }
])

// 加载环境信息
const loadEnvironmentInfo = async () => {
  try {
    const info = await GetEnvironmentInfo()
    environmentInfo.value = info
  } catch (error) {
    console.error('加载环境信息失败:', error)
  }
}

// 加载统计信息
const loadStats = async () => {
  try {
    const data = await GetStats()
    stats.value = data
  } catch (error) {
    console.error('加载统计信息失败:', error)
  }
}

// 初始化
onMounted(() => {
  console.log('Dashboard mounted')
  loadEnvironmentInfo()
  loadStats()
  
  // 每5秒刷新一次数据
  setInterval(() => {
    loadEnvironmentInfo()
    loadStats()
  }, 5000)
})
</script>

<style scoped>
/* 主容器 */
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 24px;
  width: 100%;
  padding: 32px;
  max-width: 1440px;
  margin: 0 auto;
  background: linear-gradient(135deg, #f8f9fa, #e9ecef);
  min-height: 100vh;
}

/* 暗色主题 */
body.dark-theme .dashboard {
  background: linear-gradient(135deg, rgba(10, 17, 40, 0.95), rgba(18, 30, 58, 0.95));
}

/* 顶部状态栏 */
.status-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
  margin-bottom: 32px;
}

/* 主区域 */
.main-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 24px;
  margin-bottom: 32px;
}

/* 响应式布局 */
@media (max-width: 1600px) {
  .main-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 1400px) {
  .status-bar {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .main-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 1200px) {
  .status-bar {
    grid-template-columns: 1fr;
  }
  
  .main-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .dashboard {
    padding: 20px;
    gap: 20px;
  }
  
  .status-bar {
    gap: 16px;
    margin-bottom: 20px;
  }
  
  .main-grid {
    gap: 20px;
    margin-bottom: 20px;
  }
}
</style>