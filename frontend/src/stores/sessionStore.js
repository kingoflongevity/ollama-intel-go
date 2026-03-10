import { ref, computed } from 'vue'
import { reactive } from 'vue'

const SESSIONS_STORAGE_KEY = 'ollama-chat-sessions'
const CURRENT_SESSION_KEY = 'ollama-current-session'
const MESSAGES_PREFIX = 'ollama-chat-messages-'

const sessions = ref([])
const currentSessionId = ref(null)

const generateId = () => {
  return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
}

const loadSessions = () => {
  try {
    const stored = localStorage.getItem(SESSIONS_STORAGE_KEY)
    if (stored) {
      sessions.value = JSON.parse(stored)
    }
    
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

const createSession = (name = null) => {
  const sessionName = name || `新会话 ${sessions.value.length + 1}`
  const newSession = {
    id: generateId(),
    name: sessionName,
    createdAt: Date.now(),
    updatedAt: Date.now(),
    messageCount: 0,
    model: '',
    preview: ''
  }
  
  sessions.value.unshift(newSession)
  currentSessionId.value = newSession.id
  saveSessions()
  
  localStorage.setItem(`${MESSAGES_PREFIX}${newSession.id}`, JSON.stringify([]))
  
  window.dispatchEvent(new CustomEvent('sessionCreated', {
    detail: { session: newSession }
  }))
  
  return newSession
}

const selectSession = (sessionId) => {
  if (currentSessionId.value === sessionId) return
  
  currentSessionId.value = sessionId
  saveSessions()
  
  window.dispatchEvent(new CustomEvent('sessionChanged', {
    detail: { sessionId }
  }))
}

const getCurrentSession = computed(() => {
  return sessions.value.find(s => s.id === currentSessionId.value) || null
})

const updateSession = (sessionId, updates) => {
  const session = sessions.value.find(s => s.id === sessionId)
  if (session) {
    Object.assign(session, updates)
    session.updatedAt = Date.now()
    saveSessions()
  }
}

const updateSessionPreview = (sessionId, preview, model) => {
  const session = sessions.value.find(s => s.id === sessionId)
  if (session) {
    if (preview) session.preview = preview.substring(0, 50)
    if (model) session.model = model
    session.updatedAt = Date.now()
    saveSessions()
  }
}

const incrementMessageCount = (sessionId) => {
  const session = sessions.value.find(s => s.id === sessionId)
  if (session) {
    session.messageCount = (session.messageCount || 0) + 1
    session.updatedAt = Date.now()
    saveSessions()
  }
}

const deleteSession = (sessionId) => {
  const index = sessions.value.findIndex(s => s.id === sessionId)
  if (index > -1) {
    sessions.value.splice(index, 1)
    localStorage.removeItem(`${MESSAGES_PREFIX}${sessionId}`)
    
    if (currentSessionId.value === sessionId) {
      if (sessions.value.length > 0) {
        currentSessionId.value = sessions.value[0].id
      } else {
        currentSessionId.value = null
      }
    }
    
    saveSessions()
  }
}

const renameSession = (sessionId, newName) => {
  const session = sessions.value.find(s => s.id === sessionId)
  if (session) {
    session.name = newName
    session.updatedAt = Date.now()
    saveSessions()
  }
}

const getMessages = (sessionId) => {
  try {
    const stored = localStorage.getItem(`${MESSAGES_PREFIX}${sessionId}`)
    return stored ? JSON.parse(stored) : []
  } catch (error) {
    console.error('加载消息失败:', error)
    return []
  }
}

const saveMessages = (sessionId, messages) => {
  try {
    localStorage.setItem(`${MESSAGES_PREFIX}${sessionId}`, JSON.stringify(messages))
    
    if (messages.length > 0) {
      const lastMsg = messages[messages.length - 1]
      const preview = lastMsg.role === 'user' ? lastMsg.content : 
                      messages.find(m => m.role === 'user')?.content || ''
      updateSessionPreview(sessionId, preview, null)
    }
    
    const session = sessions.value.find(s => s.id === sessionId)
    if (session) {
      session.messageCount = messages.length
      session.updatedAt = Date.now()
      saveSessions()
    }
  } catch (error) {
    console.error('保存消息失败:', error)
  }
}

const ensureSession = () => {
  if (!currentSessionId.value) {
    const session = createSession()
    return session.id
  }
  return currentSessionId.value
}

loadSessions()

export const useSessionStore = () => {
  return {
    sessions,
    currentSessionId,
    getCurrentSession,
    createSession,
    selectSession,
    updateSession,
    updateSessionPreview,
    incrementMessageCount,
    deleteSession,
    renameSession,
    getMessages,
    saveMessages,
    ensureSession,
    loadSessions,
    saveSessions
  }
}
