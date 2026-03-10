<template>
  <div class="realtime-monitor-container">
    <div class="monitor-header">
      <h3>实时监控</h3>
      <div class="monitor-status">
        <span class="status-dot" :class="{ active: isMonitoring }"></span>
        <span class="status-text">{{ isMonitoring ? '监控中' : '已停止' }}</span>
      </div>
    </div>

    <div class="monitor-grid">
      <!-- CPU 使用率 -->
      <div class="monitor-card">
        <div class="card-header">
          <el-icon><Cpu /></el-icon>
          <span>CPU 使用率</span>
        </div>
        <div class="card-content">
          <div class="metric-value">{{ cpuUsage }}%</div>
          <div class="metric-chart">
            <svg viewBox="0 0 100 100" class="progress-ring">
              <circle
                class="progress-ring-bg"
                cx="50"
                cy="50"
                r="40"
                fill="none"
                stroke-width="8"
              />
              <circle
                class="progress-ring-circle"
                cx="50"
                cy="50"
                r="40"
                fill="none"
                stroke-width="8"
                :stroke-dasharray="circumference"
                :stroke-dashoffset="cpuOffset"
                :style="{ stroke: getUsageColor(cpuUsage) }"
              />
            </svg>
          </div>
        </div>
        <div class="card-footer">
          <span class="trend" :class="cpuTrend">
            <el-icon><component :is="cpuTrendIcon" /></el-icon>
            {{ cpuTrendText }}
          </span>
        </div>
      </div>

      <!-- 内存使用情况 -->
      <div class="monitor-card">
        <div class="card-header">
          <el-icon><Memo /></el-icon>
          <span>内存使用</span>
        </div>
        <div class="card-content">
          <div class="metric-value">{{ memoryUsage }} GB</div>
          <div class="metric-sub">{{ memoryPercent }}%</div>
          <div class="metric-chart">
            <svg viewBox="0 0 100 100" class="progress-ring">
              <circle
                class="progress-ring-bg"
                cx="50"
                cy="50"
                r="40"
                fill="none"
                stroke-width="8"
              />
              <circle
                class="progress-ring-circle"
                cx="50"
                cy="50"
                r="40"
                fill="none"
                stroke-width="8"
                :stroke-dasharray="circumference"
                :stroke-dashoffset="memoryOffset"
                :style="{ stroke: getUsageColor(memoryPercent) }"
              />
            </svg>
          </div>
        </div>
        <div class="card-footer">
          <span class="trend" :class="memoryTrend">
            <el-icon><component :is="memoryTrendIcon" /></el-icon>
            {{ memoryTrendText }}
          </span>
        </div>
      </div>

      <!-- GPU 状态 -->
      <div class="monitor-card">
        <div class="card-header">
          <el-icon><Monitor /></el-icon>
          <span>GPU 状态</span>
        </div>
        <div class="card-content">
          <div class="metric-value">{{ gpuUsage }}%</div>
          <div class="metric-sub">{{ gpuMemory }} GB</div>
          <div class="metric-chart">
            <svg viewBox="0 0 100 100" class="progress-ring">
              <circle
                class="progress-ring-bg"
                cx="50"
                cy="50"
                r="40"
                fill="none"
                stroke-width="8"
              />
              <circle
                class="progress-ring-circle"
                cx="50"
                cy="50"
                r="40"
                fill="none"
                stroke-width="8"
                :stroke-dasharray="circumference"
                :stroke-dashoffset="gpuOffset"
                :style="{ stroke: getUsageColor(gpuUsage) }"
              />
            </svg>
          </div>
        </div>
        <div class="card-footer">
          <span class="trend" :class="gpuTrend">
            <el-icon><component :is="gpuTrendIcon" /></el-icon>
            {{ gpuTrendText }}
          </span>
        </div>
      </div>

      <!-- 服务运行状态 -->
      <div class="monitor-card service-status">
        <div class="card-header">
          <el-icon><Connection /></el-icon>
          <span>服务状态</span>
        </div>
        <div class="card-content">
          <div class="service-list">
            <div class="service-item" v-for="service in services" :key="service.name">
              <span class="service-name">{{ service.name }}</span>
              <el-tag
                :type="service.status === 'running' ? 'success' : 'danger'"
                size="small"
              >
                {{ service.statusText }}
              </el-tag>
            </div>
          </div>
        </div>
        <div class="card-footer">
          <span class="last-update">更新于: {{ lastUpdateTime }}</span>
        </div>
      </div>
    </div>

    <!-- 控制按钮 -->
    <div class="monitor-controls">
      <el-button
        :type="isMonitoring ? 'danger' : 'primary'"
        @click="toggleMonitoring"
        :icon="isMonitoring ? VideoPause : VideoPlay"
      >
        {{ isMonitoring ? '停止监控' : '开始监控' }}
      </el-button>
      <el-button @click="refreshData" :icon="Refresh">
        立即刷新
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Cpu, Memo, Monitor, Connection, VideoPlay, VideoPause, Refresh, ArrowUp, ArrowDown, Minus } from '@element-plus/icons-vue'

// 响应式数据
const isMonitoring = ref(true)
const cpuUsage = ref(0)
const memoryUsage = ref(0)
const memoryPercent = ref(0)
const gpuUsage = ref(0)
const gpuMemory = ref(0)
const lastUpdateTime = ref('')
const refreshInterval = ref(null)

// 趋势数据
const cpuTrend = ref('stable')
const memoryTrend = ref('stable')
const gpuTrend = ref('stable')
const previousCpu = ref(0)
const previousMemory = ref(0)
const previousGpu = ref(0)

// 服务状态
const services = ref([
  { name: 'Ollama 服务', status: 'running', statusText: '运行中' },
  { name: 'WebSocket 服务', status: 'running', statusText: '运行中' },
  { name: '模型加载器', status: 'running', statusText: '运行中' }
])

// 圆环周长
const circumference = 2 * Math.PI * 40

/**
 * 计算 CPU 偏移量
 */
const cpuOffset = computed(() => {
  return circumference - (cpuUsage.value / 100) * circumference
})

/**
 * 计算内存偏移量
 */
const memoryOffset = computed(() => {
  return circumference - (memoryPercent.value / 100) * circumference
})

/**
 * 计算 GPU 偏移量
 */
const gpuOffset = computed(() => {
  return circumference - (gpuUsage.value / 100) * circumference
})

/**
 * 获取使用率颜色
 */
const getUsageColor = (usage) => {
  if (usage >= 80) return '#f56c6c' // 红色
  if (usage >= 60) return '#e6a23c' // 橙色
  if (usage >= 40) return '#409eff' // 蓝色
  return '#67c23a' // 绿色
}

/**
 * 获取趋势图标
 */
const cpuTrendIcon = computed(() => {
  if (cpuTrend.value === 'up') return ArrowUp
  if (cpuTrend.value === 'down') return ArrowDown
  return Minus
})

const memoryTrendIcon = computed(() => {
  if (memoryTrend.value === 'up') return ArrowUp
  if (memoryTrend.value === 'down') return ArrowDown
  return Minus
})

const gpuTrendIcon = computed(() => {
  if (gpuTrend.value === 'up') return ArrowUp
  if (gpuTrend.value === 'down') return ArrowDown
  return Minus
})

/**
 * 获取趋势文本
 */
const cpuTrendText = computed(() => {
  if (cpuTrend.value === 'up') return '上升'
  if (cpuTrend.value === 'down') return '下降'
  return '稳定'
})

const memoryTrendText = computed(() => {
  if (memoryTrend.value === 'up') return '上升'
  if (memoryTrend.value === 'down') return '下降'
  return '稳定'
})

const gpuTrendText = computed(() => {
  if (gpuTrend.value === 'up') return '上升'
  if (gpuTrend.value === 'down') return '下降'
  return '稳定'
})

/**
 * 计算趋势
 */
const calculateTrend = (current, previous) => {
  const diff = current - previous
  if (diff > 5) return 'up'
  if (diff < -5) return 'down'
  return 'stable'
}

/**
 * 模拟获取系统监控数据
 * 注意：实际应用中应该从后端 API 获取真实数据
 */
const fetchMonitorData = async () => {
  try {
    // 模拟数据 - 实际应用中应该调用后端 API
    const newCpuUsage = Math.floor(Math.random() * 60) + 20 // 20-80%
    const newMemoryUsage = (Math.random() * 8 + 4).toFixed(1) // 4-12 GB
    const newMemoryPercent = Math.floor((newMemoryUsage / 16) * 100) // 假设总内存 16GB
    const newGpuUsage = Math.floor(Math.random() * 70) + 10 // 10-80%
    const newGpuMemory = (Math.random() * 6 + 2).toFixed(1) // 2-8 GB

    // 更新趋势
    cpuTrend.value = calculateTrend(newCpuUsage, previousCpu.value)
    memoryTrend.value = calculateTrend(newMemoryPercent, previousMemory.value)
    gpuTrend.value = calculateTrend(newGpuUsage, previousGpu.value)

    // 保存旧值
    previousCpu.value = cpuUsage.value
    previousMemory.value = memoryPercent.value
    previousGpu.value = gpuUsage.value

    // 更新新值
    cpuUsage.value = newCpuUsage
    memoryUsage.value = newMemoryUsage
    memoryPercent.value = newMemoryPercent
    gpuUsage.value = newGpuUsage
    gpuMemory.value = newGpuMemory

    // 更新时间
    lastUpdateTime.value = new Date().toLocaleTimeString('zh-CN')
  } catch (error) {
    console.error('获取监控数据失败:', error)
  }
}

/**
 * 切换监控状态
 */
const toggleMonitoring = () => {
  isMonitoring.value = !isMonitoring.value
  if (isMonitoring.value) {
    startMonitoring()
  } else {
    stopMonitoring()
  }
}

/**
 * 开始监控
 */
const startMonitoring = () => {
  // 每 2 秒刷新一次数据
  refreshInterval.value = setInterval(() => {
    fetchMonitorData()
  }, 2000)
  
  // 立即获取一次数据
  fetchMonitorData()
}

/**
 * 停止监控
 */
const stopMonitoring = () => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
    refreshInterval.value = null
  }
}

/**
 * 手动刷新数据
 */
const refreshData = () => {
  fetchMonitorData()
}

// 组件挂载时开始监控
onMounted(() => {
  startMonitoring()
})

// 组件卸载时清理定时器
onUnmounted(() => {
  stopMonitoring()
})
</script>

<style scoped>
.realtime-monitor-container {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  height: 100%;
  display: flex;
  flex-direction: column;
}

body.dark-theme .realtime-monitor-container {
  background: #1e1e1e;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.monitor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e4e7ed;
}

body.dark-theme .monitor-header {
  border-bottom-color: #3c3c3c;
}

.monitor-header h3 {
  margin: 0;
  color: #303133;
  font-size: 18px;
}

body.dark-theme .monitor-header h3 {
  color: #e4e6eb;
}

.monitor-status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #606266;
}

body.dark-theme .monitor-status {
  color: #a0aec0;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #f56c6c;
  animation: pulse 2s infinite;
}

.status-dot.active {
  background: #67c23a;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.monitor-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  flex: 1;
  overflow-y: auto;
}

.monitor-card {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
}

body.dark-theme .monitor-card {
  background: #2d2d2d;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
  font-size: 14px;
}

body.dark-theme .card-header {
  color: #e4e6eb;
}

.card-header .el-icon {
  color: var(--accent-primary, #06b6d4);
}

.card-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
}

.metric-value {
  font-size: 32px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 8px;
}

body.dark-theme .metric-value {
  color: #e4e6eb;
}

.metric-sub {
  font-size: 14px;
  color: #909399;
  margin-bottom: 12px;
}

body.dark-theme .metric-sub {
  color: #718096;
}

.metric-chart {
  width: 80px;
  height: 80px;
}

.progress-ring {
  transform: rotate(-90deg);
}

.progress-ring-bg {
  stroke: #e4e7ed;
}

body.dark-theme .progress-ring-bg {
  stroke: #3c3c3c;
}

.progress-ring-circle {
  transition: stroke-dashoffset 0.5s ease;
  stroke-linecap: round;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e4e7ed;
}

body.dark-theme .card-footer {
  border-top-color: #3c3c3c;
}

.trend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #909399;
}

.trend.up {
  color: #f56c6c;
}

.trend.down {
  color: #67c23a;
}

.trend.stable {
  color: #909399;
}

.service-status .card-content {
  align-items: flex-start;
}

.service-list {
  width: 100%;
}

.service-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #e4e7ed;
}

body.dark-theme .service-item {
  border-bottom-color: #3c3c3c;
}

.service-item:last-child {
  border-bottom: none;
}

.service-name {
  font-size: 14px;
  color: #606266;
}

body.dark-theme .service-name {
  color: #a0aec0;
}

.last-update {
  font-size: 12px;
  color: #909399;
}

body.dark-theme .last-update {
  color: #718096;
}

.monitor-controls {
  display: flex;
  gap: 12px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #e4e7ed;
}

body.dark-theme .monitor-controls {
  border-top-color: #3c3c3c;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .monitor-grid {
    grid-template-columns: 1fr;
  }
}
</style>
