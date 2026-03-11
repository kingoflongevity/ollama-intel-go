<template>
  <div class="tech-sessions-container">
    <div class="tech-page-header">
      <div class="header-left">
        <div class="header-title">
          <div class="title-icon">
            <el-icon><ChatDotRound /></el-icon>
          </div>
          <div class="title-text">
            <h2>会话管理中心</h2>
            <span class="subtitle">Session Management</span>
          </div>
        </div>
        <div class="session-stats">
          <div class="stat-item">
            <span class="stat-value">{{ sessions.length }}</span>
            <span class="stat-label">会话总数</span>
          </div>
          <div class="stat-item">
            <span class="stat-value">{{ totalMessages }}</span>
            <span class="stat-label">消息总数</span>
          </div>
        </div>
      </div>
      <div class="header-right">
        <button class="tech-btn tech-btn-primary" @click="createNewSession">
          <el-icon><Plus /></el-icon>
          新建会话
        </button>
      </div>
    </div>

    <div class="sessions-content">
      <div v-if="sessions.length === 0" class="empty-state">
        <div class="empty-icon">
          <el-icon><ChatDotRound /></el-icon>
        </div>
        <h3>暂无会话记录</h3>
        <p>开始聊天时会自动创建会话，或点击上方按钮新建会话</p>
        <button class="tech-btn tech-btn-primary" @click="createNewSession">
          <el-icon><Plus /></el-icon>
          新建会话
        </button>
      </div>

      <div v-else class="sessions-list">
        <div
          v-for="session in sortedSessions"
          :key="session.id"
          class="session-card"
          :class="{ active: currentSessionId === session.id }"
          @click="selectSession(session.id)"
        >
          <div class="session-card-header">
            <div class="session-icon">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <div class="session-info">
              <h4 class="session-name" v-if="editingSessionId !== session.id">{{ session.name }}</h4>
              <el-input
                v-else
                v-model="editingName"
                size="small"
                @blur="saveSessionName(session.id)"
                @keydown.enter="saveSessionName(session.id)"
                @keydown.escape="cancelEdit"
                ref="editInputRef"
                autofocus
                class="tech-input-sm"
              />
              <span class="session-time">{{ formatTime(session.updatedAt) }}</span>
            </div>
            <div class="session-actions" v-show="editingSessionId !== session.id">
              <button class="action-btn" @click.stop="startEditSession(session)" title="重命名">
                <el-icon><Edit /></el-icon>
              </button>
              <button class="action-btn danger" @click.stop="confirmDeleteSession(session)" title="删除">
                <el-icon><Delete /></el-icon>
              </button>
            </div>
          </div>
          
          <div class="session-card-body">
            <div class="session-meta">
              <div class="meta-item">
                <el-icon><ChatLineRound /></el-icon>
                <span>{{ session.messageCount || 0 }} 条消息</span>
              </div>
              <div class="meta-item" v-if="session.model">
                <el-icon><Box /></el-icon>
                <span>{{ session.model }}</span>
              </div>
            </div>
            <div v-if="session.preview" class="session-preview">{{ session.preview }}</div>
          </div>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="deleteDialogVisible"
      title=""
      width="420px"
      class="tech-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <div class="dialog-icon danger">
            <el-icon><Delete /></el-icon>
          </div>
          <div class="dialog-title">
            <h3>确认删除</h3>
            <span>此操作不可恢复</span>
          </div>
        </div>
      </template>
      
      <div class="dialog-content">
        <p class="confirm-text">确定要删除会话 "<strong>{{ sessionToDelete?.name }}</strong>" 吗？</p>
        <p class="confirm-hint">删除后，该会话的所有消息记录将被永久删除。</p>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <button class="tech-btn" @click="deleteDialogVisible = false">取消</button>
          <button class="tech-btn tech-btn-danger" @click="handleDeleteSession">确定删除</button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { Plus, ChatDotRound, Edit, Delete, ChatLineRound, Box } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useSessionStore } from '@/stores/sessionStore'

const router = useRouter()
const sessionStore = useSessionStore()

const sessions = computed(() => sessionStore.sessions.value)
const currentSessionId = computed(() => sessionStore.currentSessionId.value)

const totalMessages = computed(() => {
  return sessions.value.reduce((total, session) => total + (session.messageCount || 0), 0)
})

const sortedSessions = computed(() => {
  return [...sessions.value].sort((a, b) => {
    const timeA = a.updatedAt ? new Date(a.updatedAt).getTime() : 0
    const timeB = b.updatedAt ? new Date(b.updatedAt).getTime() : 0
    return timeB - timeA
  })
})

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
  
  if (diff < 60000) {
    return '刚刚'
  }
  
  if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000)
    return `${minutes}分钟前`
  }
  
  if (date.toDateString() === now.toDateString()) {
    return '今天 ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
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
.tech-sessions-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(135deg, #0a0a0f 0%, #1a1a2e 50%, #0f0f1a 100%);
  position: relative;
  overflow: hidden;
}

.tech-sessions-container::before {
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

.session-stats {
  display: flex;
  gap: 16px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 16px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 8px;
  border: 1px solid rgba(6, 182, 212, 0.2);
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: #06b6d4;
}

.stat-label {
  font-size: 11px;
  color: #64748b;
  text-transform: uppercase;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
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

.tech-btn-primary {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  border: none;
  color: white;
}

.tech-btn-primary:hover {
  background: linear-gradient(135deg, #22d3ee 0%, #06b6d4 100%);
  box-shadow: 0 4px 15px rgba(6, 182, 212, 0.4);
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

.sessions-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  position: relative;
  z-index: 1;
}

.sessions-content::-webkit-scrollbar {
  width: 6px;
}

.sessions-content::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.5);
}

.sessions-content::-webkit-scrollbar-thumb {
  background: rgba(6, 182, 212, 0.3);
  border-radius: 3px;
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
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
  margin: 0 0 24px 0;
  color: #64748b;
  text-align: center;
}

.sessions-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.session-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(51, 65, 85, 0.5);
  border-radius: 12px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.session-card:hover {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(6, 182, 212, 0.3);
  transform: translateX(4px);
}

.session-card.active {
  background: rgba(6, 182, 212, 0.1);
  border-color: rgba(6, 182, 212, 0.5);
  box-shadow: 0 0 20px rgba(6, 182, 212, 0.15);
}

.session-card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.session-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 16px;
  flex-shrink: 0;
}

.session-card.active .session-icon {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
}

.session-info {
  flex: 1;
  min-width: 0;
}

.session-name {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.session-time {
  font-size: 11px;
  color: #64748b;
}

.session-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.session-card:hover .session-actions {
  opacity: 1;
}

.action-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(100, 116, 139, 0.3);
  border-radius: 6px;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: rgba(51, 65, 85, 0.8);
  border-color: rgba(6, 182, 212, 0.5);
  color: #06b6d4;
}

.action-btn.danger:hover {
  border-color: rgba(239, 68, 68, 0.5);
  color: #ef4444;
}

.session-card-body {
  padding-left: 48px;
}

.session-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #64748b;
}

.meta-item .el-icon {
  font-size: 14px;
}

.session-preview {
  font-size: 12px;
  color: #475569;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tech-input-sm {
  width: 100%;
}

.tech-input-sm :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 6px;
  box-shadow: none;
}

.tech-input-sm :deep(.el-input__inner) {
  color: #e2e8f0;
  font-size: 13px;
}

.tech-dialog :deep(.el-dialog) {
  background: rgba(15, 15, 25, 0.95);
  border: 1px solid rgba(6, 182, 212, 0.2);
  border-radius: 16px;
}

.tech-dialog :deep(.el-dialog__header) {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(51, 65, 85, 0.5);
}

.tech-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

.tech-dialog :deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid rgba(51, 65, 85, 0.5);
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dialog-icon {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
}

.dialog-icon.danger {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
}

.dialog-title h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #e2e8f0;
}

.dialog-title span {
  font-size: 12px;
  color: #64748b;
}

.dialog-content {
  min-height: 60px;
}

.confirm-text {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #cbd5e1;
}

.confirm-text strong {
  color: #06b6d4;
}

.confirm-hint {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
