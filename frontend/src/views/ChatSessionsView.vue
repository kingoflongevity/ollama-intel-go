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
          <div v-if="session.preview" class="session-preview">{{ session.preview }}</div>
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

      <el-empty v-if="sessions.length === 0" description="暂无会话，开始聊天时会自动创建会话" />
    </div>

    <el-dialog
      v-model="deleteDialogVisible"
      title="确认删除"
      width="400px"
    >
      <p>确定要删除会话 "{{ sessionToDelete?.name }}" 吗？此操作不可恢复。</p>
      <template #footer>
        <el-button @click="deleteDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="handleDeleteSession">确定删除</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { Plus, ChatDotRound, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useSessionStore } from '@/stores/sessionStore'

const router = useRouter()
const sessionStore = useSessionStore()

const sessions = computed(() => sessionStore.sessions.value)
const currentSessionId = computed(() => sessionStore.currentSessionId.value)

const editingSessionId = ref(null)
const editingName = ref('')
const deleteDialogVisible = ref(false)
const sessionToDelete = ref(null)
const editInputRef = ref(null)

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date
  
  if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000)
    return minutes <= 1 ? '刚刚' : `${minutes}分钟前`
  }
  
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  if (date.toDateString() === yesterday.toDateString()) {
    return '昨天 ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

const createNewSession = () => {
  const session = sessionStore.createSession()
  ElMessage.success('会话创建成功')
  router.push('/chat')
}

const selectSession = (sessionId) => {
  sessionStore.selectSession(sessionId)
  router.push('/chat')
}

const startEditSession = async (session) => {
  editingSessionId.value = session.id
  editingName.value = session.name
  
  await nextTick()
  if (editInputRef.value) {
    editInputRef.value.focus()
  }
}

const saveSessionName = (sessionId) => {
  if (!editingName.value.trim()) {
    ElMessage.warning('会话名称不能为空')
    return
  }
  
  sessionStore.renameSession(sessionId, editingName.value.trim())
  editingSessionId.value = null
  editingName.value = ''
}

const cancelEdit = () => {
  editingSessionId.value = null
  editingName.value = ''
}

const confirmDeleteSession = (session) => {
  sessionToDelete.value = session
  deleteDialogVisible.value = true
}

const handleDeleteSession = () => {
  if (!sessionToDelete.value) return
  
  sessionStore.deleteSession(sessionToDelete.value.id)
  ElMessage.success('会话删除成功')
  
  deleteDialogVisible.value = false
  sessionToDelete.value = null
}

onMounted(() => {
  sessionStore.loadSessions()
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

.session-preview {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
