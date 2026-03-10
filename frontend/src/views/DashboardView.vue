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
      <div class="left-panel">
        <ControlPanel />
        <ActivityPanel :recentChats="recentChats" />
      </div>
      <div class="right-panel">
        <RealTimeMonitor />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import StatusCard from '@/components/StatusCard.vue'
import ControlPanel from '@/components/ControlPanel.vue'
import ActivityPanel from '@/components/ActivityPanel.vue'
import RealTimeMonitor from '@/components/RealTimeMonitor.vue'
import { GetEnvironmentInfo, GetStats } from '../../wailsjs/go/main/App'

// 响应式数据
const environmentInfo = ref({})
const stats = ref({})
const recentChats = ref([
  { id: 1, content: '关于 Python 编程的讨论' },
  { id: 2, content: '如何使用 Vue.js 构建应用' },
  { id: 3, content: '机器学习模型训练技巧' }
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
    const s = await GetStats()
    stats.value = s
  } catch (error) {
    console.error('加载统计信息失败:', error)
  }
}

onMounted(() => {
  loadEnvironmentInfo()
  loadStats()
  
  // 定期刷新环境信息
  setInterval(() => {
    loadEnvironmentInfo()
  }, 5000)
})
</script>

<style scoped>
.dashboard {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 24px;
  height: 100%;
}

.status-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.main-grid {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 24px;
  flex: 1;
  min-height: 0;
}

.left-panel {
  display: flex;
  flex-direction: column;
  gap: 24px;
  flex: 1;
  min-height: 0;
}

.right-panel {
  height: 100%;
}

/* 响应式布局 */
@media (max-width: 1600px) {
  .main-grid {
    grid-template-columns: 1fr;
  }
  
  .right-panel {
    height: 400px;
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
}
</style>
