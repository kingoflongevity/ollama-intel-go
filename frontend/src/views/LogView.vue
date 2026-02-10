<template>
  <div class="log-view">
    <el-page-header @back="$router.back()">
      <template #content>
        <span class="text-large font-600 mr-3">日志管理</span>
      </template>
      <template #extra>
        <el-button-group>
          <el-button type="primary" @click="clearLogs" :disabled="logs.length === 0">
            <el-icon><Delete /></el-icon>
            清空日志
          </el-button>
          <el-button @click="toggleAutoScroll" :type="autoScroll ? 'success' : 'default'">
            <el-icon><VideoPlay /></el-icon>
            {{ autoScroll ? '自动滚动开启' : '自动滚动关闭' }}
          </el-button>
          <el-button @click="exportLogs">
            <el-icon><Download /></el-icon>
            导出日志
          </el-button>
        </el-button-group>
      </template>
    </el-page-header>

    <el-divider />

    <div class="log-container">
      <div class="log-toolbar">
        <el-input
          v-model="filterText"
          placeholder="过滤日志内容..."
          clearable
          style="width: 300px"
          @clear="clearFilter"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        
        <div class="log-stats">
          <el-tag type="info">总日志数: {{ logs.length }}</el-tag>
          <el-tag type="success" v-if="filteredLogs.length !== logs.length">
            过滤后: {{ filteredLogs.length }}
          </el-tag>
          <el-tag :type="logLevelTag.type">
            {{ logLevelTag.text }}
          </el-tag>
        </div>
      </div>

      <div class="log-content" ref="logContentRef">
        <div
          v-for="(log, index) in filteredLogs"
          :key="index"
          class="log-entry"
          :class="getLogLevelClass(log.level)"
          @click="copyLog(log.message)"
        >
          <div class="log-time">{{ log.time }}</div>
          <div class="log-level">{{ log.level }}</div>
          <div class="log-message">{{ log.message }}</div>
          <div class="log-actions">
            <el-icon class="copy-icon" @click.stop="copyLog(log.message)">
              <CopyDocument />
            </el-icon>
          </div>
        </div>
        
        <div v-if="filteredLogs.length === 0" class="empty-log">
          <el-empty description="暂无日志" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Download, Search, CopyDocument, VideoPlay } from '@element-plus/icons-vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const logs = ref([])
const filterText = ref('')
const autoScroll = ref(true)
const logContentRef = ref(null)

// 监听后端日志事件
onMounted(() => {
  // 立即添加一条测试日志，确认前端日志功能正常
  addLog('INFO', '日志管理页面已加载，开始监听后端事件')
  
  // 监听通用日志事件
  EventsOn('log', (logMessage) => {
    addLog('INFO', logMessage)
  })
  
  // 监听模型拉取进度事件
  EventsOn('model_pull_progress', (eventData) => {
    const { model, status, progress, message, time } = eventData
    let level = 'INFO'
    if (status === 'error') level = 'ERROR'
    if (status === 'completed') level = 'SUCCESS'
    
    addLog(level, `[模型拉取] ${model}: ${message} (进度: ${progress}%)`)
  })
  
  // 监听其他可能的事件
  EventsOn('service_status', (data) => {
    addLog('INFO', `[服务状态] ${data.status}`)
  })
})

onUnmounted(() => {
  // 清理事件监听
  EventsOff('log')
  EventsOff('model_pull_progress')
  EventsOff('service_status')
})

// 添加日志
const addLog = (level, message) => {
  const timestamp = new Date()
  const timeStr = timestamp.toLocaleTimeString('zh-CN', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    fractionalSecondDigits: 3
  })
  
  logs.value.push({
    time: timeStr,
    level,
    message,
    timestamp
  })
  
  // 限制日志数量，避免内存占用过大
  if (logs.value.length > 1000) {
    logs.value = logs.value.slice(-500)
  }
  
  // 自动滚动到底部
  if (autoScroll.value) {
    nextTick(() => {
      const container = logContentRef.value
      if (container) {
        container.scrollTop = container.scrollHeight
      }
    })
  }
}

// 过滤日志
const filteredLogs = computed(() => {
  if (!filterText.value.trim()) {
    return logs.value
  }
  
  const filter = filterText.value.toLowerCase()
  return logs.value.filter(log => 
    log.message.toLowerCase().includes(filter) ||
    log.level.toLowerCase().includes(filter) ||
    log.time.toLowerCase().includes(filter)
  )
})

// 获取日志级别对应的样式类
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

// 日志级别标签
const logLevelTag = computed(() => {
  const errorCount = logs.value.filter(log => log.level === 'ERROR').length
  const warningCount = logs.value.filter(log => log.level === 'WARNING').length
  
  if (errorCount > 0) {
    return { type: 'danger', text: `错误: ${errorCount}` }
  } else if (warningCount > 0) {
    return { type: 'warning', text: `警告: ${warningCount}` }
  } else {
    return { type: 'success', text: '正常' }
  }
})

// 清空日志
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

// 切换自动滚动
const toggleAutoScroll = () => {
  autoScroll.value = !autoScroll.value
  ElMessage.info(`自动滚动 ${autoScroll.value ? '开启' : '关闭'}`)
}

// 清除过滤
const clearFilter = () => {
  filterText.value = ''
}

// 复制日志
const copyLog = async (message) => {
  try {
    await navigator.clipboard.writeText(message)
    ElMessage.success('日志已复制到剪贴板')
  } catch (err) {
    console.error('复制失败:', err)
    ElMessage.error('复制失败')
  }
}

// 导出日志
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
  
  ElMessage.success(`日志已导出，共 ${logs.value.length} 条`)
}
</script>

<style scoped>
.log-view {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.log-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f8f9fa;
  border-radius: 8px;
  padding: 20px;
  overflow: hidden;
}

.log-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 15px;
  border-bottom: 1px solid #e4e7ed;
}

.log-stats {
  display: flex;
  gap: 10px;
}

.log-content {
  flex: 1;
  overflow-y: auto;
  background: white;
  border-radius: 6px;
  padding: 15px;
  border: 1px solid #e4e7ed;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
}

.log-entry {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  margin-bottom: 6px;
  border-radius: 4px;
  border-left: 4px solid;
  cursor: pointer;
  transition: background-color 0.2s;
}

.log-entry:hover {
  background-color: #f5f7fa;
}

.log-time {
  width: 90px;
  color: #909399;
  font-size: 11px;
  flex-shrink: 0;
}

.log-level {
  width: 70px;
  font-weight: bold;
  text-align: center;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  margin: 0 10px;
  flex-shrink: 0;
}

.log-message {
  flex: 1;
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.5;
}

.log-actions {
  width: 30px;
  text-align: center;
  flex-shrink: 0;
}

.copy-icon {
  color: #909399;
  cursor: pointer;
  transition: color 0.2s;
}

.copy-icon:hover {
  color: #409eff;
}

/* 日志级别样式 */
.log-error {
  border-left-color: #f56c6c;
  background-color: #fef0f0;
}

.log-error .log-level {
  background-color: #f56c6c;
  color: white;
}

.log-warning {
  border-left-color: #e6a23c;
  background-color: #fdf6ec;
}

.log-warning .log-level {
  background-color: #e6a23c;
  color: white;
}

.log-info {
  border-left-color: #409eff;
  background-color: #ecf5ff;
}

.log-info .log-level {
  background-color: #409eff;
  color: white;
}

.log-success {
  border-left-color: #67c23a;
  background-color: #f0f9eb;
}

.log-success .log-level {
  background-color: #67c23a;
  color: white;
}

.log-debug {
  border-left-color: #909399;
  background-color: #f4f4f5;
}

.log-debug .log-level {
  background-color: #909399;
  color: white;
}

.empty-log {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

/* 暗色主题支持 */
body.dark-theme .log-view {
  background-color: #1e1e1e;
}

body.dark-theme .log-container {
  background: #2d2d2d;
}

body.dark-theme .log-toolbar {
  border-bottom-color: #4a4a4a;
}

body.dark-theme .log-content {
  background: #1e1e1e !important;
  border-color: #4a4a4a !important;
}

body.dark-theme .log-entry {
  background-color: #1e1e1e !important;
  color: #e4e6eb !important;
}

body.dark-theme .log-entry:hover {
  background-color: #2d2d2d !important;
}

body.dark-theme .log-time {
  color: #a0aec0;
}

body.dark-theme .copy-icon {
  color: #a0aec0;
}

body.dark-theme .copy-icon:hover {
  color: #409eff;
}

/* 暗色主题 - Element Plus 组件样式 */
body.dark-theme :deep(.el-empty) {
  --el-empty-fill-color: #1e1e1e;
  --el-empty-text-color: #a0aec0;
}

body.dark-theme :deep(.el-empty__description) {
  color: #a0aec0 !important;
}

body.dark-theme :deep(.el-tag) {
  --el-tag-bg-color: #2d2d2d;
  --el-tag-border-color: #4a4a4a;
  --el-tag-text-color: #e4e6eb;
}

body.dark-theme :deep(.el-input__wrapper) {
  --el-input-bg-color: #2d2d2d;
  --el-input-border-color: #4a4a4a;
  --el-input-text-color: #e4e6eb;
  --el-input-placeholder-color: #a0aec0;
}
</style>