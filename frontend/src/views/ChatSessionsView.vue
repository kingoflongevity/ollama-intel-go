<template>
  <div class="page-container">
    <div class="page-header">
      <h2>会话管理</h2>
      <el-button type="primary" @click="createNewSession" class="unified-button">
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
              class="unified-input"
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
            class="unified-button"
          >
            <el-icon><Edit /></el-icon>
          </el-button>
          <el-button
            link
            size="small"
            type="danger"
            @click.stop="confirmDeleteSession(session)"
            title="删除"
            class="unified-button"
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
      class="unified-dialog"
    >
      <p>确定要删除会话 "{{ sessionToDelete?.name }}" 吗？此操作不可恢复。</p>
      <template #footer>
        <el-button @click="deleteDialogVisible = false" class="unified-button">取消</el-button>
        <el-button type="danger" @click="handleDeleteSession" class="unified-button">确定删除</el-button>
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

/**
 * 格式化时间显示
 */
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

/**
 * 创建新会话
 */
const createNewSession = () => {
  const session = sessionStore.createSession()
  ElMessage.success('会话创建成功')
  router.push('/chat')
}

/**
 * 选择会话
 */
const selectSession = (sessionId) => {
  sessionStore.selectSession(sessionId)
  router.push('/chat')
}

/**
 * 开始编辑会话名称
 */
const startEditSession = async (session) => {
  editingSessionId.value = session.id
  editingName.value = session.name
  
  await nextTick()
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
  
  sessionStore.renameSession(sessionId, editingName.value.trim())
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
 * 处理删除会话
 */
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
.sessions-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-md);
}

.session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md) var(--spacing-lg);
  margin-bottom: var(--spacing-sm);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all var(--transition-base);
  border: 1px solid transparent;
  background: var(--bg-elevated);
}

.session-item:hover {
  background: var(--bg-tertiary);
  border-color: var(--border-color);
  box-shadow: var(--shadow-sm);
}

.session-item.active {
  background: var(--gradient-primary-light);
  border-color: var(--color-primary);
  box-shadow: 0 4px 16px rgba(6, 182, 212, 0.2);
}

.session-info {
  flex: 1;
  min-width: 0;
}

.session-name {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
  margin-bottom: var(--spacing-xs);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-name .el-icon {
  flex-shrink: 0;
  color: var(--color-primary);
}

.session-name .el-input {
  width: 150px;
}

.session-meta {
  display: flex;
  gap: var(--spacing-md);
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.session-preview {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  margin-top: var(--spacing-xs);
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
  gap: var(--spacing-xs);
  opacity: 0;
  transition: opacity var(--transition-base);
}

.session-item:hover .session-actions {
  opacity: 1;
}

.session-actions .el-button {
  padding: var(--spacing-xs);
}
</style>
