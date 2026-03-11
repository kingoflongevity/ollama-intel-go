<template>
  <div class="tech-chat-container">
    <div class="tech-chat-header">
      <div class="header-left">
        <div class="header-title">
          <div class="title-icon">
            <el-icon><ChatDotRound /></el-icon>
          </div>
          <div class="title-text">
            <h2>AI 对话终端</h2>
            <span class="subtitle">Intelligent Conversation System</span>
          </div>
        </div>
      </div>
      <div class="header-right">
        <div class="tech-select-wrapper">
          <el-select v-model="selectedModel" placeholder="选择模型" size="default" class="tech-select">
            <el-option
              v-for="model in models"
              :key="model.name"
              :label="model.name"
              :value="model.name"
            />
          </el-select>
        </div>
        <div class="tech-select-wrapper">
          <el-select v-model="selectedRole" placeholder="选择角色" size="default" class="tech-select" @change="handleRoleChange">
            <el-option label="默认助手" value="default" />
            <el-option label="代码专家" value="code" />
            <el-option label="视频脚本专家" value="video" />
            <el-option label="写作专家" value="writing" />
            <el-option label="商业顾问" value="business" />
            <el-option label="教育专家" value="education" />
          </el-select>
        </div>
        <div class="header-actions">
          <button class="tech-btn tech-btn-outline" @click="clearChat">
            <el-icon><Delete /></el-icon>
            清空
          </button>
          <button class="tech-btn" :class="showSearch ? 'tech-btn-active' : 'tech-btn-primary'" @click="toggleSearch">
            <el-icon><Search /></el-icon>
            {{ showSearch ? '关闭搜索' : '联网搜索' }}
          </button>
        </div>
      </div>
    </div>

    <div class="chat-main-area">
      <div class="chat-messages" ref="messagesContainerRef">
        <div v-if="messages.length === 0" class="empty-chat">
          <div class="empty-icon">
            <el-icon><ChatDotRound /></el-icon>
          </div>
          <h3>开始新的对话</h3>
          <p>选择一个模型，输入您的问题开始对话</p>
          <div class="quick-actions">
            <div class="quick-action" @click="setQuickPrompt('帮我写一段Python代码')">
              <el-icon><Document /></el-icon>
              <span>编写代码</span>
            </div>
            <div class="quick-action" @click="setQuickPrompt('解释一下什么是机器学习')">
              <el-icon><Reading /></el-icon>
              <span>知识问答</span>
            </div>
            <div class="quick-action" @click="setQuickPrompt('帮我翻译这段文字')">
              <el-icon><Edit /></el-icon>
              <span>翻译文本</span>
            </div>
          </div>
        </div>
        
        <div v-for="(msg, index) in messages" :key="index" class="message" :class="msg.role">
          <div class="message-avatar">
            <el-icon v-if="msg.role === 'user'" size="20"><User /></el-icon>
            <el-icon v-else-if="msg.role === 'assistant'" size="20"><ChatDotRound /></el-icon>
            <el-icon v-else-if="msg.role === 'system'" size="20"><Operation /></el-icon>
            <el-icon v-else-if="msg.role === 'search'" size="20"><Search /></el-icon>
          </div>
          <div class="message-body">
            <div class="message-header">
              <span class="message-role">
                {{ msg.role === 'user' ? '用户' : 
                   msg.role === 'assistant' ? 'AI助手' : 
                   msg.role === 'system' ? '系统' : '搜索' }}
              </span>
              <span class="message-time">{{ msg.timestamp }}</span>
            </div>
            <div class="message-content" v-html="formatMessage(msg.content)" />
            <div v-if="msg.role === 'search' && msg.results" class="search-results">
              <div v-for="(result, idx) in msg.results" :key="idx" class="search-result-item">
                <a :href="result.url" target="_blank">{{ result.title }}</a>
                <p>{{ result.snippet }}</p>
              </div>
            </div>
          </div>
        </div>
        
        <div v-if="isLoading" class="message assistant loading">
          <div class="message-avatar">
            <el-icon size="20"><ChatDotRound /></el-icon>
          </div>
          <div class="message-body">
            <div class="typing-indicator">
              <span></span>
              <span></span>
              <span></span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showSearch" class="search-input-panel">
      <div class="panel-header">
        <el-icon><Search /></el-icon>
        <span>联网搜索</span>
      </div>
      <div class="panel-body">
        <el-input
          v-model="searchQuery"
          placeholder="输入搜索关键词..."
          maxlength="500"
          @keydown.enter="performSearch"
          :disabled="isLoading"
          class="tech-input"
        />
        <button class="tech-btn tech-btn-primary" :disabled="!searchQuery.trim() || isLoading" @click="performSearch">
          <el-icon><Search /></el-icon>
          搜索
        </button>
      </div>
    </div>

    <div class="chat-input-panel">
      <div class="input-wrapper">
        <el-input
          v-model="inputMessage"
          :rows="3"
          type="textarea"
          placeholder="输入消息... (Enter发送, Shift+Enter换行)"
          maxlength="2000"
          @keydown.enter="handleEnter"
          :disabled="isLoading"
          class="tech-textarea"
        />
        <div class="input-footer">
          <div class="input-info">
            <span class="char-count">{{ inputMessage.length }}/2000</span>
            <span class="model-info">模型: {{ selectedModel }}</span>
          </div>
          <button class="tech-btn tech-btn-primary tech-btn-send" :disabled="!inputMessage.trim() || isLoading" @click="sendMessage">
            <el-icon><Promotion /></el-icon>
            发送
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from 'vue'
import { User, ChatDotRound, Promotion, Search, Operation, Delete, Document, Reading, Edit } from '@element-plus/icons-vue'
import { ListModels, ChatStream } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useSessionStore } from '@/stores/sessionStore'
import { useConfigStore } from '@/stores/configStore'
import { useRouter } from 'vue-router'

const router = useRouter()
const sessionStore = useSessionStore()
const configStore = useConfigStore()

const messages = ref([])
const inputMessage = ref('')
const isLoading = ref(false)
const selectedModel = ref('llama3:8b')
const models = ref([])
const messagesContainerRef = ref(null)
const currentStreamMessageIndex = ref(-1)
const showSearch = ref(false)
const selectedRole = ref('default')
const searchQuery = ref('')
const ws = ref(null)
const wsConnected = ref(false)

const currentSessionId = computed(() => sessionStore.currentSessionId.value)

const SELECTED_MODEL_KEY = 'ollama-selected-model'

const saveSelectedModel = () => {
  try {
    localStorage.setItem(SELECTED_MODEL_KEY, selectedModel.value)
  } catch (error) {
    console.error('保存模型选择失败:', error)
  }
}

const loadSelectedModel = () => {
  try {
    const saved = localStorage.getItem(SELECTED_MODEL_KEY)
    if (saved) {
      selectedModel.value = saved
    }
  } catch (error) {
    console.error('加载模型选择失败:', error)
  }
}

watch(selectedModel, () => {
  saveSelectedModel()
})

const saveMessages = () => {
  if (currentSessionId.value) {
    sessionStore.saveMessages(currentSessionId.value, messages.value)
  }
}

const loadMessages = () => {
  if (currentSessionId.value) {
    messages.value = sessionStore.getMessages(currentSessionId.value)
  } else {
    messages.value = []
  }
}

const formatMessage = (text) => {
  return text.replace(/\n/g, '<br>')
}

const connectWebSocket = () => {
  try {
    const wsUrl = configStore.getWebSocketUrl()
    console.log('正在连接WebSocket:', wsUrl)
    ws.value = new WebSocket(wsUrl)
    
    ws.value.onopen = () => {
      console.log('WebSocket连接已建立')
      wsConnected.value = true
    }
    
    ws.value.onmessage = (event) => {
      const data = JSON.parse(event.data)
      handleWebSocketMessage(data)
    }
    
    ws.value.onerror = (error) => {
      console.error('WebSocket错误:', error)
      wsConnected.value = false
    }
    
    ws.value.onclose = () => {
      console.log('WebSocket连接已关闭')
      wsConnected.value = false
    }
  } catch (error) {
    console.error('WebSocket连接失败:', error)
  }
}

const handleWebSocketMessage = (data) => {
  console.log('收到WebSocket消息:', data)
  
  switch (data.type) {
    case 'stream':
      if (currentStreamMessageIndex.value >= 0 && currentStreamMessageIndex.value < messages.value.length) {
        messages.value[currentStreamMessageIndex.value] = {
          role: 'assistant',
          content: data.full_content || data.content || '',
          timestamp: new Date().toLocaleTimeString()
        }
        saveMessages()
      }
      break
    case 'done':
      if (currentStreamMessageIndex.value >= 0 && currentStreamMessageIndex.value < messages.value.length) {
        messages.value[currentStreamMessageIndex.value] = {
          role: 'assistant',
          content: data.content || '',
          timestamp: new Date().toLocaleTimeString()
        }
        saveMessages()
      }
      isLoading.value = false
      currentStreamMessageIndex.value = -1
      break
    case 'error':
      if (currentStreamMessageIndex.value >= 0 && currentStreamMessageIndex.value < messages.value.length) {
        messages.value[currentStreamMessageIndex.value] = {
          role: 'assistant',
          content: data.content || '抱歉，处理您的请求时出现错误。',
          timestamp: new Date().toLocaleTimeString()
        }
        saveMessages()
      }
      isLoading.value = false
      currentStreamMessageIndex.value = -1
      break
    case 'search':
      const searchMessage = {
        role: 'search',
        content: data.content || '',
        results: data.results || [],
        timestamp: new Date().toLocaleTimeString()
      }
      messages.value.push(searchMessage)
      saveMessages()
      break
    case 'role':
      const roleMessage = {
        role: 'system',
        content: data.content || '',
        timestamp: new Date().toLocaleTimeString()
      }
      messages.value.push(roleMessage)
      saveMessages()
      break
  }
}

const sendMessage = async () => {
  if (!inputMessage.value.trim()) return

  const sessionId = sessionStore.ensureSession()
  
  if (messages.value.length === 0) {
    const firstMsg = inputMessage.value.trim()
    const sessionName = firstMsg.length > 20 ? firstMsg.substring(0, 20) + '...' : firstMsg
    sessionStore.renameSession(sessionId, sessionName)
  }

  const userMessage = {
    role: 'user',
    content: inputMessage.value,
    timestamp: new Date().toLocaleTimeString()
  }

  messages.value.push(userMessage)
  saveMessages()
  
  inputMessage.value = ''
  isLoading.value = true

  const aiMessagePlaceholder = {
    role: 'assistant',
    content: '',
    timestamp: new Date().toLocaleTimeString()
  }
  currentStreamMessageIndex.value = messages.value.length
  messages.value.push(aiMessagePlaceholder)
  saveMessages()

  sessionStore.updateSession(sessionId, { model: selectedModel.value })

  try {
    // 通过后端代理调用Ollama API，避免跨域问题
    const request = {
      model: selectedModel.value,
      messages: messages.value.filter(m => m.role !== 'assistant').map(m => ({
        role: m.role,
        content: m.content
      })),
      stream: true
    }
    
    console.log('发送聊天请求:', request.model, '消息数:', request.messages.length)
    
    const results = await ChatStream(request)
    
    console.log('收到聊天结果:', results.length, '条')
    
    if (results && results.length > 0) {
      const lastResult = results[results.length - 1]
      
      if (lastResult.error) {
        messages.value[currentStreamMessageIndex.value] = {
          role: 'assistant',
          content: `错误: ${lastResult.error}`,
          timestamp: new Date().toLocaleTimeString()
        }
      } else {
        // 使用最后一个结果的内容
        messages.value[currentStreamMessageIndex.value] = {
          role: 'assistant',
          content: lastResult.content || '',
          timestamp: new Date().toLocaleTimeString()
        }
      }
      saveMessages()
    } else {
      messages.value[currentStreamMessageIndex.value] = {
        role: 'assistant',
        content: '抱歉，未收到响应。请检查Ollama服务是否正常运行。',
        timestamp: new Date().toLocaleTimeString()
      }
      saveMessages()
    }
  } catch (error) {
    console.error('发送消息失败:', error)
    let errorMessage = '抱歉，连接AI服务时出现错误，请稍后重试。'
    
    if (error.message) {
      errorMessage = `错误: ${error.message}`
    }
    
    if (currentStreamMessageIndex.value >= 0 && currentStreamMessageIndex.value < messages.value.length) {
      messages.value[currentStreamMessageIndex.value] = {
        role: 'assistant',
        content: errorMessage,
        timestamp: new Date().toLocaleTimeString()
      }
      saveMessages()
    }
  } finally {
    isLoading.value = false
    currentStreamMessageIndex.value = -1
  }
}

const handleEnter = (event) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}

const clearChat = () => {
  messages.value = []
  saveMessages()
}

const loadModels = async () => {
  try {
    const modelList = await ListModels()
    models.value = modelList.map(model => ({ name: model.name }))
    if (models.value.length > 0 && !selectedModel.value) {
      selectedModel.value = models.value[0].name
    }
  } catch (error) {
    console.error('加载模型列表失败:', error)
    models.value = [
      { name: 'llama3:8b' },
      { name: 'mistral:7b' },
      { name: 'gemma:2b' }
    ]
  }
}

const toggleSearch = () => {
  showSearch.value = !showSearch.value
}

const performSearch = async () => {
  if (!searchQuery.value.trim()) return
  isLoading.value = true

  try {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      connectWebSocket()
      await new Promise(resolve => setTimeout(resolve, 1000))
    }

    const searchRequest = {
      type: 'search',
      messages: [{ role: 'user', content: searchQuery.value }]
    }

    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(searchRequest))
    } else {
      const searchMessage = {
        role: 'search',
        content: `搜索结果: ${searchQuery.value}`,
        results: [
          { title: '搜索结果1', url: 'https://example.com/1', snippet: '这是搜索结果1的摘要信息。' },
          { title: '搜索结果2', url: 'https://example.com/2', snippet: '这是搜索结果2的摘要信息。' }
        ],
        timestamp: new Date().toLocaleTimeString()
      }
      messages.value.push(searchMessage)
      saveMessages()
    }
  } catch (error) {
    console.error('搜索失败:', error)
  } finally {
    isLoading.value = false
  }
}

const handleRoleChange = async () => {
  if (selectedRole.value === 'default') return

  try {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      connectWebSocket()
      await new Promise(resolve => setTimeout(resolve, 1000))
    }

    const roleRequest = { type: 'role', role: selectedRole.value }

    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(roleRequest))
    } else {
      const rolePrompts = {
        code: '你是一位专业的代码专家，擅长解决各种编程问题。',
        video: '你是一位专业的视频脚本编写专家。',
        writing: '你是一位专业的写作专家。',
        business: '你是一位专业的商业顾问。',
        education: '你是一位专业的教育专家。'
      }
      const roleMessage = {
        role: 'system',
        content: rolePrompts[selectedRole.value] || '已切换到指定角色',
        timestamp: new Date().toLocaleTimeString()
      }
      messages.value.push(roleMessage)
      saveMessages()
    }
  } catch (error) {
    console.error('角色切换失败:', error)
  }
}

const setQuickPrompt = (prompt) => {
  inputMessage.value = prompt
}

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainerRef.value) {
    messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight
  }
}

watch(messages, () => scrollToBottom(), { deep: true })
watch(isLoading, () => scrollToBottom())

watch(currentSessionId, (newId, oldId) => {
  if (newId !== oldId) {
    loadMessages()
  }
})

onMounted(() => {
  loadModels().then(() => {
    loadSelectedModel()
  })
  loadMessages()
  connectWebSocket()
  
  EventsOn('chat_stream', (eventData) => {
    if (currentStreamMessageIndex.value >= 0 && currentStreamMessageIndex.value < messages.value.length) {
      messages.value[currentStreamMessageIndex.value] = {
        role: 'assistant',
        content: eventData.full_content || eventData.content || '',
        timestamp: new Date().toLocaleTimeString()
      }
      saveMessages()
    }
  })
  
  window.addEventListener('sessionChanged', (event) => {
    loadMessages()
  })
  
  if (messages.value.length === 0) {
    setTimeout(() => {
      const welcomeMessage = {
        role: 'assistant',
        content: '您好！我是Ollama 英特尔优化版的AI助手。我可以帮助您回答问题、编写代码、创作内容等。请随时告诉我您需要什么帮助！',
        timestamp: new Date().toLocaleTimeString()
      }
      messages.value.push(welcomeMessage)
      saveMessages()
    }, 500)
  }
})

onUnmounted(() => {
  EventsOff('chat_stream')
  if (ws.value) ws.value.close()
  window.removeEventListener('sessionChanged', () => {})
})
</script>

<style scoped>
.tech-chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(135deg, #0a0a0f 0%, #1a1a2e 50%, #0f0f1a 100%);
  position: relative;
  overflow: hidden;
}

.tech-chat-container::before {
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

.tech-chat-header {
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

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tech-select-wrapper {
  position: relative;
}

.tech-select {
  width: 140px;
}

.tech-select :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 8px;
  box-shadow: none;
}

.tech-select :deep(.el-input__inner) {
  color: #e2e8f0;
}

.header-actions {
  display: flex;
  gap: 8px;
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

.tech-btn-active {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  border: none;
  color: white;
}

.tech-btn-outline {
  background: transparent;
  border: 1px solid rgba(239, 68, 68, 0.5);
  color: #ef4444;
}

.tech-btn-outline:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: #ef4444;
}

.chat-main-area {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  position: relative;
  z-index: 1;
}

.chat-messages {
  height: 100%;
  overflow-y: auto;
  padding: 24px;
  padding-bottom: 100px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.chat-messages::-webkit-scrollbar {
  width: 6px;
}

.chat-messages::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.5);
}

.chat-messages::-webkit-scrollbar-thumb {
  background: rgba(6, 182, 212, 0.3);
  border-radius: 3px;
}

.empty-chat {
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

.empty-chat h3 {
  margin: 0 0 8px 0;
  font-size: 20px;
  color: #e2e8f0;
}

.empty-chat p {
  margin: 0 0 30px 0;
  color: #64748b;
}

.quick-actions {
  display: flex;
  gap: 16px;
}

.quick-action {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 24px;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(6, 182, 212, 0.2);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  color: #94a3b8;
}

.quick-action:hover {
  background: rgba(6, 182, 212, 0.1);
  border-color: rgba(6, 182, 212, 0.5);
  color: #06b6d4;
  transform: translateY(-2px);
}

.quick-action .el-icon {
  font-size: 24px;
}

.quick-action span {
  font-size: 13px;
}

.message {
  display: flex;
  gap: 16px;
  animation: fadeInUp 0.3s ease;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.message.user .message-avatar {
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
}

.message.assistant .message-avatar {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  color: white;
}

.message.system .message-avatar {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  color: white;
}

.message.search .message-avatar {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
}

.message-body {
  flex: 1;
  min-width: 0;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.message-role {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
}

.message-time {
  font-size: 11px;
  color: #475569;
}

.message-content {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(51, 65, 85, 0.5);
  border-radius: 12px;
  padding: 14px 18px;
  color: #cbd5e1;
  line-height: 1.6;
  white-space: pre-wrap;
}

.message.user .message-content {
  background: rgba(59, 130, 246, 0.15);
  border-color: rgba(59, 130, 246, 0.3);
}

.message.assistant .message-content {
  background: rgba(6, 182, 212, 0.1);
  border-color: rgba(6, 182, 212, 0.2);
}

.typing-indicator {
  display: flex;
  gap: 6px;
  align-items: center;
  padding: 14px 18px;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(6, 182, 212, 0.2);
  border-radius: 12px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  background: #06b6d4;
  border-radius: 50%;
  animation: typing 1.4s infinite ease-in-out both;
}

.typing-indicator span:nth-child(1) { animation-delay: -0.32s; }
.typing-indicator span:nth-child(2) { animation-delay: -0.16s; }

@keyframes typing {
  0%, 80%, 100% { transform: scale(0.8); opacity: 0.5; }
  40% { transform: scale(1.2); opacity: 1; }
}

.search-results {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.search-result-item {
  padding: 12px;
  background: rgba(30, 41, 59, 0.4);
  border-radius: 8px;
  border: 1px solid rgba(51, 65, 85, 0.3);
}

.search-result-item a {
  color: #06b6d4;
  font-weight: 500;
  text-decoration: none;
}

.search-result-item a:hover {
  text-decoration: underline;
}

.search-result-item p {
  margin: 6px 0 0 0;
  font-size: 12px;
  color: #64748b;
}

.search-input-panel,
.chat-input-panel {
  background: rgba(15, 15, 25, 0.9);
  backdrop-filter: blur(20px);
  border-top: 1px solid rgba(6, 182, 212, 0.2);
  position: relative;
  z-index: 10;
}

.search-input-panel {
  padding: 16px 24px;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  color: #06b6d4;
  font-size: 13px;
  font-weight: 500;
}

.panel-body {
  display: flex;
  gap: 12px;
}

.tech-input :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 8px;
  box-shadow: none;
}

.tech-input :deep(.el-input__inner) {
  color: #e2e8f0;
}

.chat-input-panel {
  padding: 20px 24px;
}

.input-wrapper {
  max-width: 900px;
  margin: 0 auto;
}

.tech-textarea :deep(.el-textarea__inner) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 12px;
  color: #e2e8f0;
  resize: none;
  font-size: 14px;
  line-height: 1.6;
}

.tech-textarea :deep(.el-textarea__inner:focus) {
  border-color: rgba(6, 182, 212, 0.6);
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.tech-textarea :deep(.el-textarea__inner::placeholder) {
  color: #475569;
}

.input-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
}

.input-info {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #475569;
}

.tech-btn-send {
  padding: 10px 24px;
  font-size: 14px;
}

.tech-btn-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
