<template>
  <div class="chat-container">
    <div class="chat-header">
      <h2>AI 对话</h2>
      <div class="chat-controls">
        <el-select v-model="selectedModel" placeholder="选择模型" size="small" style="width: 150px;">
          <el-option
            v-for="model in models"
            :key="model.name"
            :label="model.name"
            :value="model.name"
          />
        </el-select>
        <el-select v-model="selectedRole" placeholder="选择角色" size="small" style="width: 150px;" @change="handleRoleChange">
          <el-option label="默认助手" value="default" />
          <el-option label="代码专家" value="code" />
          <el-option label="视频脚本专家" value="video" />
          <el-option label="写作专家" value="writing" />
          <el-option label="商业顾问" value="business" />
          <el-option label="教育专家" value="education" />
        </el-select>
        <el-button size="small" @click="clearChat">清空对话</el-button>
        <el-button size="small" @click="toggleSearch" :type="showSearch ? 'primary' : 'default'">
          <el-icon><Search /></el-icon>
          {{ showSearch ? '关闭搜索' : '联网搜索' }}
        </el-button>
      </div>
    </div>

    <div class="chat-messages" ref="messagesContainerRef">
      <div v-for="(msg, index) in messages" :key="index" class="message">
        <div class="message-avatar">
          <el-icon v-if="msg.role === 'user'" size="20"><User /></el-icon>
          <el-icon v-else-if="msg.role === 'assistant'" size="20"><ChatDotRound /></el-icon>
          <el-icon v-else-if="msg.role === 'system'" size="20"><Operation /></el-icon>
          <el-icon v-else-if="msg.role === 'search'" size="20"><Search /></el-icon>
        </div>
        <div class="message-content">
          <div class="message-role">
            {{ msg.role === 'user' ? '用户' : 
               msg.role === 'assistant' ? 'AI助手' : 
               msg.role === 'system' ? '系统' : '搜索' }}
          </div>
          <div class="message-text" v-html="formatMessage(msg.content)" />
          <div v-if="msg.role === 'search' && msg.results" class="search-results">
            <div v-for="(result, idx) in msg.results" :key="idx" class="search-result-item">
              <a :href="result.url" target="_blank">{{ result.title }}</a>
              <p>{{ result.snippet }}</p>
            </div>
          </div>
          <div class="message-time">{{ msg.timestamp }}</div>
        </div>
      </div>
      <div v-if="isLoading" class="loading-message">
        <div class="message-avatar">
          <el-icon size="20"><ChatDotRound /></el-icon>
        </div>
        <div class="message-content">
          <div class="typing-indicator">
            <span></span>
            <span></span>
            <span></span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showSearch" class="search-input-area">
      <el-input
        v-model="searchQuery"
        placeholder="输入搜索关键词..."
        maxlength="500"
        show-word-limit
        @keydown.enter="performSearch"
        :disabled="isLoading"
      />
      <div class="input-controls">
        <el-button 
          type="primary" 
          :disabled="!searchQuery.trim() || isLoading"
          @click="performSearch"
        >
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>

    <div class="chat-input-area">
      <el-input
        v-model="inputMessage"
        :rows="3"
        type="textarea"
        placeholder="输入消息..."
        maxlength="2000"
        show-word-limit
        @keydown.enter="handleEnter"
        :disabled="isLoading"
      />
      <div class="input-controls">
        <el-button 
          type="primary" 
          :disabled="!inputMessage.trim() || isLoading"
          @click="sendMessage"
        >
          <el-icon><Promotion /></el-icon>
          发送
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from 'vue'
import { User, ChatDotRound, Promotion, Search, Operation } from '@element-plus/icons-vue'
import { ListModels } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useSessionStore } from '@/stores/sessionStore'
import { useRouter } from 'vue-router'

const router = useRouter()
const sessionStore = useSessionStore()

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
    const wsUrl = `ws://127.0.0.1:11435/ws/chat`
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

/**
 * 发送消息 - 核心功能
 * 发送消息时自动创建会话并保存聊天记录
 */
const sendMessage = async () => {
  if (!inputMessage.value.trim()) return

  // 确保有当前会话，如果没有则自动创建
  const sessionId = sessionStore.ensureSession()
  
  // 如果是新会话，根据第一条消息自动命名
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
  
  const msg = inputMessage.value
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

  // 更新会话使用的模型
  sessionStore.updateSession(sessionId, { model: selectedModel.value })

  try {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      connectWebSocket()
      await new Promise(resolve => setTimeout(resolve, 1000))
    }

    const chatRequest = {
      type: 'chat',
      model: selectedModel.value,
      messages: messages.value.filter(m => m.role !== 'assistant').map(m => ({
        role: m.role,
        content: m.content
      }))
    }

    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(chatRequest))
    } else {
      await sendMessageWithHTTP()
    }
  } catch (error) {
    console.error('发送消息失败:', error)
    if (currentStreamMessageIndex.value >= 0 && currentStreamMessageIndex.value < messages.value.length) {
      messages.value[currentStreamMessageIndex.value] = {
        role: 'assistant',
        content: '抱歉，连接AI服务时出现错误，请稍后重试。',
        timestamp: new Date().toLocaleTimeString()
      }
      saveMessages()
    }
    isLoading.value = false
    currentStreamMessageIndex.value = -1
  }
}

const sendMessageWithHTTP = async () => {
  try {
    const request = {
      model: selectedModel.value,
      messages: messages.value.filter(m => m.role !== 'assistant').map(m => ({
        role: m.role,
        content: m.content
      })),
      stream: true
    }

    const response = await fetch('http://127.0.0.1:11434/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request)
    })

    if (!response.ok) throw new Error('HTTP请求失败')

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let fullContent = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      const chunk = decoder.decode(value, { stream: true })
      const lines = chunk.split('\n')

      for (const line of lines) {
        if (line.trim() === '') continue
        try {
          const data = JSON.parse(line)
          if (data.message && data.message.content) {
            fullContent += data.message.content
            if (currentStreamMessageIndex.value >= 0 && currentStreamMessageIndex.value < messages.value.length) {
              messages.value[currentStreamMessageIndex.value] = {
                role: 'assistant',
                content: fullContent,
                timestamp: new Date().toLocaleTimeString()
              }
              saveMessages()
            }
          }
          if (data.done) break
        } catch (e) {
          console.error('解析流式响应失败:', e)
        }
      }
    }
  } catch (error) {
    console.error('HTTP API发送失败:', error)
    throw error
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

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainerRef.value) {
    messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight
  }
}

watch(messages, () => scrollToBottom(), { deep: true })
watch(isLoading, () => scrollToBottom())

// 监听会话切换
watch(currentSessionId, (newId, oldId) => {
  if (newId !== oldId) {
    loadMessages()
  }
})

onMounted(() => {
  loadModels()
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
  
  // 监听会话切换事件
  window.addEventListener('sessionChanged', (event) => {
    loadMessages()
  })
  
  // 如果没有消息，显示欢迎消息
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
.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

body.dark-theme .chat-container {
  background: #1e1e1e;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
}

body.dark-theme .chat-header {
  background: #2d2d2d;
  border-bottom-color: #3c3c3c;
}

.chat-header h2 {
  margin: 0;
  color: #303133;
}

body.dark-theme .chat-header h2 {
  color: #e4e6eb;
}

.chat-controls {
  display: flex;
  gap: 12px;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #4f46e5;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

body.dark-theme .message-avatar {
  background: #6366f1;
}

.message-content {
  flex: 1;
}

.message-role {
  font-weight: 600;
  margin-bottom: 4px;
  color: #606266;
  font-size: 14px;
}

body.dark-theme .message-role {
  color: #a0aec0;
}

.message-text {
  background: #f8f9fa;
  padding: 12px 16px;
  border-radius: 12px;
  line-height: 1.6;
  color: #303133;
  white-space: pre-wrap;
}

body.dark-theme .message-text {
  background: #2d2d2d;
  color: #e4e6eb;
}

.message-time {
  text-align: right;
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

body.dark-theme .message-time {
  color: #718096;
}

.loading-message {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.typing-indicator {
  display: flex;
  gap: 4px;
  align-items: center;
  height: 24px;
  padding: 12px 16px;
  background: #f8f9fa;
  border-radius: 12px;
}

body.dark-theme .typing-indicator {
  background: #2d2d2d;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  background: #909399;
  border-radius: 50%;
  display: inline-block;
  animation: typing 1.4s infinite ease-in-out both;
}

.typing-indicator span:nth-child(1) { animation-delay: -0.32s; }
.typing-indicator span:nth-child(2) { animation-delay: -0.16s; }

@keyframes typing {
  0%, 80%, 100% { transform: scale(0.8); opacity: 0.5; }
  40% { transform: scale(1.2); opacity: 1; }
}

.chat-input-area {
  padding: 20px;
  border-top: 1px solid #e4e7ed;
  background: white;
}

body.dark-theme .chat-input-area {
  background: #1e1e1e;
  border-top-color: #3c3c3c;
}

.input-controls {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}

.el-textarea :deep(.el-textarea__inner) {
  border-radius: 12px;
  border: 1px solid #dcdfe6;
  resize: vertical;
}

body.dark-theme .el-textarea :deep(.el-textarea__inner) {
  border-color: #4a4a4a;
  background: #2d2d2d;
  color: #e4e6eb;
}

.search-input-area {
  padding: 12px 20px;
  border-top: 1px solid #e4e7ed;
  background: #fafafa;
}

body.dark-theme .search-input-area {
  background: #2d2d2d;
  border-top-color: #3c3c3c;
}
</style>
