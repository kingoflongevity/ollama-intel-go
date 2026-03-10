<template>
  <div class="chat-sessions-container">
    <div class="sessions-header">
      <h2>会话管理</h2>
      <el-button type="primary" @click="createNewSession">
        <el-icon><Plus /></el-icon>
        新建会话
      </el-button>
    </div>

    <div class="sessions-list">
      <div
        v-for="session in sessions"
        :key="session.id"
        class="session-item"
        :class="{ active: currentSessionId === session.id }"
        @click="selectSession(session.id)"
      >
        <div class="session-info">
          <div class="session-name">
            <el-icon><ChatDotRound /></el-icon>
            <span v-if="editingSessionId !== session.id">{{ session.name }}</span>
            <el-input
              v-else
              v-model="editingName"
              size="small"
              @blur="saveSessionName(session.id)"
              @keydown.enter="saveSessionName(session.id)"
              @keydown.escape="cancelEdit"
              ref="editInputRef"
              autofocus
            />
          </div>
          <div class="session-meta">
            <span class="session-time">{{ formatTime(session.updatedAt) }}</span>
            <span class="session-count">{{ session.messageCount || 0 }} 条消息</span>
          </div>
        </div>
        <div class="session-actions" v-show="editingSessionId !== session.id">
          <el-button
            link
            size="small"
            @click.stop="startEditSession(session)"
            title="重命名"
          >
            <el-icon><Edit /></el-icon>
          </el-button>
          <el-button
            link
            size="small"
            type="danger"
            @click.stop="confirmDeleteSession(session)"
            title="删除"
          >
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
      </div>

      <el-empty v-if="sessions.length === 0" description="暂无会话，点击上方按钮创建新会话" />
    </div>

    <!-- 删除确认对话框 -->
    <el-dialog
      v-model="deleteDialogVisible"
      title="确认删除"
      width="400px"
    >
      <p>确定要删除会话 "{{ sessionToDelete?.name }}" 吗？此操作不可恢复。</p>
      <template #footer>
        <el-button @click="deleteDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="deleteSession">确定删除</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { Plus, ChatDotRound, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

// 存储键名
const SESSIONS_STORAGE_KEY = 'ollama-chat-sessions'
const CURRENT_SESSION_KEY = 'ollama-current-session'

// 响应式数据
const sessions = ref([])
const currentSessionId = ref(null)
const editingSessionId = ref(null)
const editingName = ref('')
const deleteDialogVisible = ref(false)
const sessionToDelete = ref(null)
const editInputRef = ref(null)

/**
 * 生成唯一ID
 */
const generateId = () => {
  return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
}

/**
 * 格式化时间
 */
const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date
  
  // 一小时内
  if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000)
    return minutes <= 1 ? '刚刚' : `${minutes}分钟前`
  }
  
  // 今天
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  
  // 昨天
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  if (date.toDateString() === yesterday.toDateString()) {
    return '昨天 ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  
  // 其他
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

/**
 * 加载会话列表
 */
const loadSessions = () => {
  try {
    const stored = localStorage.getItem(SESSIONS_STORAGE_KEY)
    if (stored) {
      sessions.value = JSON.parse(stored)
    }
    
    // 加载当前会话ID
    const currentId = localStorage.getItem(CURRENT_SESSION_KEY)
    if (currentId && sessions.value.find(s => s.id === currentId)) {
      currentSessionId.value = currentId
    } else if (sessions.value.length > 0) {
      currentSessionId.value = sessions.value[0].id
    }
  } catch (error) {
    console.error('加载会话列表失败:', error)
    sessions.value = []
  }
}

/**
 * 保存会话列表
 */
const saveSessions = () => {
  try {
    localStorage.setItem(SESSIONS_STORAGE_KEY, JSON.stringify(sessions.value))
    if (currentSessionId.value) {
      localStorage.setItem(CURRENT_SESSION_KEY, currentSessionId.value)
    }
  } catch (error) {
    console.error('保存会话列表失败:', error)
  }
}

/**
 * 创建新会话
 */
const createNewSession = () => {
  const newSession = {
    id: generateId(),
    name: `新会话 ${sessions.value.length + 1}`,
    createdAt: Date.now(),
    updatedAt: Date.now(),
    messageCount: 0
  }
  
  sessions.value.unshift(newSession)
  currentSessionId.value = newSession.id
  saveSessions()
  
  // 触发会话切换事件
  emitSessionChange(newSession.id)
  
  ElMessage.success('会话创建成功')
}

/**
 * 选择会话
 */
const selectSession = (sessionId) => {
  if (currentSessionId.value === sessionId) return
  
  currentSessionId.value = sessionId
  saveSessions()
  
  // 触发会话切换事件
  emitSessionChange(sessionId)
}

/**
 * 触发会话切换事件
 */
const emitSessionChange = (sessionId) => {
  window.dispatchEvent(new CustomEvent('sessionChanged', {
    detail: { sessionId }
  }))
}

/**
 * 开始编辑会话名称
 */
const startEditSession = async (session) => {
  editingSessionId.value = session.id
  editingName.value = session.name
  
  await nextTick()
  // 聚焦输入框
  if (editInputRef.value) {
    editInputRef.value.focus()
  }
}

/**
 * 保存会话名称
 */
const saveSessionName = (sessionId) => {
  if (!editingName.value.trim()) {
    ElMessage.warning('会话名称不能为空')
    return
  }
  
  const session = sessions.value.find(s => s.id === sessionId)
  if (session) {
    session.name = editingName.value.trim()
    session.updatedAt = Date.now()
    saveSessions()
  }
  
  editingSessionId.value = null
  editingName.value = ''
}

/**
 * 取消编辑
 */
const cancelEdit = () => {
  editingSessionId.value = null
  editingName.value = ''
}

/**
 * 确认删除会话
 */
const confirmDeleteSession = (session) => {
  sessionToDelete.value = session
  deleteDialogVisible.value = true
}

/**
 * 删除会话
 */
const deleteSession = () => {
  if (!sessionToDelete.value) return
  
  const index = sessions.value.findIndex(s => s.id === sessionToDelete.value.id)
  if (index > -1) {
    sessions.value.splice(index, 1)
    
    // 删除会话的消息记录
    localStorage.removeItem(`ollama-chat-messages-${sessionToDelete.value.id}`)
    
    // 如果删除的是当前会话，切换到其他会话
    if (currentSessionId.value === sessionToDelete.value.id) {
      if (sessions.value.length > 0) {
        currentSessionId.value = sessions.value[0].id
        emitSessionChange(currentSessionId.value)
      } else {
        currentSessionId.value = null
      }
    }
    
    saveSessions()
    ElMessage.success('会话删除成功')
  }
  
  deleteDialogVisible.value = false
  sessionToDelete.value = null
}

/**
 * 更新会话消息数量
 */
const updateSessionMessageCount = (sessionId, count) => {
  const session = sessions.value.find(s => s.id === sessionId)
  if (session) {
    session.messageCount = count
    session.updatedAt = Date.now()
    saveSessions()
  }
}

// 暴露方法给外部调用
defineExpose({
  updateSessionMessageCount,
  getCurrentSessionId: () => currentSessionId.value,
  getSessions: () => sessions.value
})

// 初始化
onMounted(() => {
  loadSessions()
  
  // 监听消息数量更新事件
  window.addEventListener('messageCountUpdated', (event) => {
    const { sessionId, count } = event.detail
    updateSessionMessageCount(sessionId, count)
  })
})
</script>

<style scoped>
.chat-sessions-container {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  height: 100%;
  display: flex;
  flex-direction: column;
}

body.dark-theme .chat-sessions-container {
  background: #1e1e1e;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.sessions-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
}

body.dark-theme .sessions-header {
  background: #2d2d2d;
  border-bottom-color: #3c3c3c;
}

.sessions-header h2 {
  margin: 0;
  color: #303133;
}

body.dark-theme .sessions-header h2 {
  color: #e4e6eb;
}

.sessions-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  margin-bottom: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid transparent;
}

.session-item:hover {
  background: #f5f7fa;
  border-color: #e4e7ed;
}

.session-item.active {
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.1), rgba(139, 92, 246, 0.1));
  border-color: var(--accent-primary, #06b6d4);
}

body.dark-theme .session-item:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: #3c3c3c;
}

body.dark-theme .session-item.active {
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.2), rgba(139, 92, 246, 0.2));
  border-color: var(--accent-primary, #06b6d4);
}

.session-info {
  flex: 1;
  min-width: 0;
}

.session-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

body.dark-theme .session-name {
  color: #e4e6eb;
}

.session-name .el-icon {
  flex-shrink: 0;
  color: var(--accent-primary, #06b6d4);
}

.session-name .el-input {
  width: 150px;
}

.session-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #909399;
}

body.dark-theme .session-meta {
  color: #718096;
}

.session-time,
.session-count {
  display: flex;
  align-items: center;
}

.session-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.session-item:hover .session-actions {
  opacity: 1;
}

.session-actions .el-button {
  padding: 4px;
}
</style>
