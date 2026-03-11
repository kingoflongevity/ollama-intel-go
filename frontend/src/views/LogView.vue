<template>
  <div class="tech-log-container">
    <div class="tech-page-header">
      <div class="header-left">
        <div class="header-title">
          <div class="title-icon">
            <el-icon><Document /></el-icon>
          </div>
          <div class="title-text">
            <h2>系统日志</h2>
            <span class="subtitle">System Logs</span>
          </div>
        </div>
        <div class="log-stats">
          <div class="stat-item">
            <span class="stat-value">{{ logs.length }}</span>
            <span class="stat-label">总日志</span>
          </div>
          <div class="stat-item error" v-if="errorCount > 0">
            <span class="stat-value">{{ errorCount }}</span>
            <span class="stat-label">错误</span>
          </div>
          <div class="stat-item warning" v-if="warningCount > 0">
            <span class="stat-value">{{ warningCount }}</span>
            <span class="stat-label">警告</span>
          </div>
        </div>
      </div>
      <div class="header-right">
        <div class="search-box">
          <el-icon class="search-icon"><Search /></el-icon>
          <input
            v-model="filterText"
            type="text"
            placeholder="过滤日志..."
            class="tech-search-input"
          />
        </div>
        <button class="tech-btn" :class="autoScroll ? 'tech-btn-active' : ''" @click="toggleAutoScroll">
          <el-icon><VideoPlay /></el-icon>
          {{ autoScroll ? '自动滚动' : '手动' }}
        </button>
        <button class="tech-btn" @click="exportLogs">
          <el-icon><Download /></el-icon>
          导出
        </button>
        <button class="tech-btn tech-btn-danger" @click="clearLogs" :disabled="logs.length === 0">
          <el-icon><Delete /></el-icon>
          清空
        </button>
      </div>
    </div>

    <div class="log-content" ref="logContentRef">
      <div v-if="filteredLogs.length === 0" class="empty-state">
        <div class="empty-icon">
          <el-icon><Document /></el-icon>
        </div>
        <h3>暂无日志</h3>
        <p>系统运行日志将在此处显示</p>
      </div>
      
      <div v-else class="log-entries">
        <div
          v-for="(log, index) in filteredLogs"
          :key="index"
          class="log-entry"
          :class="getLogLevelClass(log.level)"
        >
          <div class="log-time">{{ log.time }}</div>
          <div class="log-level">
            <span class="level-badge" :class="log.level.toLowerCase()">{{ log.level }}</span>
          </div>
          <div class="log-message">{{ log.message }}</div>
          <div class="log-actions">
            <button class="action-btn" @click="copyLog(log.message)" title="复制">
              <el-icon><CopyDocument /></el-icon>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { Document, Search, VideoPlay, Download, Delete, CopyDocument } from '@element-plus/icons-vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { ElMessage, ElMessageBox } from 'element-plus'

const logs = ref([])
const filterText = ref('')
const autoScroll = ref(true)
const logContentRef = ref(null)

const errorCount = computed(() => logs.value.filter(log => log.level === 'ERROR').length)
const warningCount = computed(() => logs.value.filter(log => log.level === 'WARNING').length)

onMounted(() => {
  addLog('INFO', '日志管理页面已加载')
  
  EventsOn('log', (logMessage) => {
    addLog('INFO', logMessage)
  })
  
  EventsOn('model_pull_progress', (eventData) => {
    const { model, status, progress, message } = eventData
    let level = 'INFO'
    if (status === 'error') level = 'ERROR'
    if (status === 'completed') level = 'SUCCESS'
    
    addLog(level, `[模型拉取] ${model}: ${message} (${progress}%)`)
  })
  
  EventsOn('service_status', (data) => {
    addLog('INFO', `[服务状态] ${data.status}`)
  })
})

onUnmounted(() => {
  EventsOff('log')
  EventsOff('model_pull_progress')
  EventsOff('service_status')
})

const addLog = (level, message) => {
  const timestamp = new Date()
  const timeStr = timestamp.toLocaleTimeString('zh-CN', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
  
  logs.value.push({
    time: timeStr,
    level,
    message,
    timestamp
  })
  
  if (logs.value.length > 1000) {
    logs.value = logs.value.slice(-500)
  }
  
  if (autoScroll.value) {
    nextTick(() => {
      const container = logContentRef.value
      if (container) {
        container.scrollTop = container.scrollHeight
      }
    })
  }
}

const filteredLogs = computed(() => {
  if (!filterText.value.trim()) {
    return logs.value
  }
  
  const filter = filterText.value.toLowerCase()
  return logs.value.filter(log => 
    log.message.toLowerCase().includes(filter) ||
    log.level.toLowerCase().includes(filter)
  )
})

const getLogLevelClass = (level) => {
  const levelMap = {
    'ERROR': 'log-error',
    'WARNING': 'log-warning',
    'INFO': 'log-info',
    'SUCCESS': 'log-success',
    'DEBUG': 'log-debug'
  }
  return levelMap[level] || 'log-info'
}

const toggleAutoScroll = () => {
  autoScroll.value = !autoScroll.value
  ElMessage.info(`自动滚动 ${autoScroll.value ? '已开启' : '已关闭'}`)
}

const clearLogs = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有日志吗？', '清空日志', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    logs.value = []
    ElMessage.success('日志已清空')
  } catch (error) {
    // 用户取消
  }
}

const copyLog = async (message) => {
  try {
    await navigator.clipboard.writeText(message)
    ElMessage.success('已复制到剪贴板')
  } catch (err) {
    console.error('复制失败:', err)
    ElMessage.error('复制失败')
  }
}

const exportLogs = () => {
  if (logs.value.length === 0) {
    ElMessage.warning('没有日志可导出')
    return
  }
  
  const logText = logs.value.map(log => 
    `[${log.time}] [${log.level}] ${log.message}`
  ).join('\n')
  
  const blob = new Blob([logText], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `ollama-logs-${new Date().toISOString().slice(0, 10)}.txt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  ElMessage.success(`已导出 ${logs.value.length} 条日志`)
}
</script>

<style scoped>
.tech-log-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(135deg, #0a0a0f 0%, #1a1a2e 50%, #0f0f1a 100%);
  position: relative;
  overflow: hidden;
}

.tech-log-container::before {
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
  gap: 24px;
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

.log-stats {
  display: flex;
  gap: 12px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6px 14px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 8px;
  border: 1px solid rgba(6, 182, 212, 0.2);
}

.stat-item.error {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.3);
}

.stat-item.error .stat-value {
  color: #ef4444;
}

.stat-item.warning{
  background: rgba(245, 158, 11, 0.15);
  border-color: rgba(245, 158, 11, 0.3);
}

.stat-item.warning .stat-value{
  color: #f59e0b;
}

.stat-value {
  font-size: 16px;
  font-weight: 700;
  color: #06b6d4;
}

.stat-label {
  font-size: 10px;
  color: #64748b;
  text-transform: uppercase;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: #64748b;
  font-size: 16px;
}

.tech-search-input {
  width: 200px;
  padding: 8px 12px 8px 36px;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 8px;
  color: #e2e8f0;
  font-size: 13px;
  outline: none;
  transition: all 0.3s ease;
}

.tech-search-input::placeholder {
  color: #475569;
}

.tech-search-input:focus {
  border-color: rgba(6, 182, 212, 0.6);
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
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

.tech-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.tech-btn-active {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  border: none;
  color: white;
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

.log-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  position: relative;
  z-index: 1;
}

.log-content::-webkit-scrollbar {
  width: 6px;
}

.log-content::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.5);
}

.log-content::-webkit-scrollbar-thumb {
  background: rgba(6, 182, 212, 0.3);
  border-radius: 3px;
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #64748b;
}

.empty-icon {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.2) 0%, rgba(139, 92, 246, 0.2) 100%);
  border-radius: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36px;
  color: #06b6d4;
  margin-bottom: 20px;
}

.empty-state h3 {
  margin: 0 0 8px 0;
  font-size: 20px;
  color: #e2e8f0;
}

.empty-state p {
  margin: 0;
  color: #64748b;
}

.log-entries {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.log-entry {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(30, 41, 59, 0.4);
  border-radius: 10px;
  border-left: 3px solid;
  transition: all 0.2s ease;
}

.log-entry:hover {
  background: rgba(30, 41, 59, 0.6);
}

.log-time {
  font-size: 11px;
  color: #64748b;
  font-family: monospace;
  white-space: nowrap;
  min-width: 70px;
}

.log-level {
  min-width: 60px;
}

.level-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.level-badge.info {
  background: rgba(59, 130, 246, 0.2);
  color: #3b82f6;
}

.level-badge.error{
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.level-badge.warning{
  background: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
}

.level-badge.success{
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
}

.level-badge.debug{
  background: rgba(100, 116, 139, 0.2);
  color: #94a3b8;
}

.log-message {
  flex: 1;
  font-size: 12px;
  color: #cbd5e1;
  line-height: 1.5;
  word-break: break-word;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
}

.log-actions {
  display: flex;
  gap: 4px;
}

.action-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: rgba(6, 182, 212, 0.15);
  color: #06b6d4;
}

.log-error {
  border-left-color: #ef4444;
  background: rgba(239, 68, 68, 0.05);
}

.log-warning {
  border-left-color: #f59e0b;
  background: rgba(245, 158, 11, 0.05);
}

.log-info {
  border-left-color: #3b82f6;
}

.log-success {
  border-left-color: #10b981;
  background: rgba(16, 185, 129, 0.05);
}

.log-debug {
  border-left-color: #64748b;
}
</style>
