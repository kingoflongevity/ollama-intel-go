<template>
  <div class="page-container chat-container">
    <div class="page-header chat-header">
      <h2>AI 对话</h2>
      <div class="header-actions">
        <el-select v-model="selectedModel" placeholder="选择模型" size="default" style="width: 150px;">
          <el-option
            v-for="model in models"
            :key="model.name"
            :label="model.name"
            :value="model.name"
          />
        </el-select>
        <el-select v-model="selectedRole" placeholder="选择角色" size="default" style="width: 150px;" @change="handleRoleChange">
          <el-option label="默认助手" value="default" />
          <el-option label="代码专家" value="code" />
          <el-option label="视频脚本专家" value="video" />
          <el-option label="写作专家" value="writing" />
          <el-option label="商业顾问" value="business" />
          <el-option label="教育专家" value="education" />
        </el-select>
        <el-button @click="clearChat" class="unified-button">清空对话</el-button>
        <el-button @click="toggleSearch" :type="showSearch ? 'primary' : 'default'" class="unified-button">
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
        class="unified-input"
      />
      <div class="input-controls">
        <el-button 
          type="primary" 
          :disabled="!searchQuery.trim() || isLoading"
          @click="performSearch"
          class="unified-button"
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
        class="unified-input"
      />
      <div class="input-controls">
        <el-button 
          type="primary" 
          :disabled="!inputMessage.trim() || isLoading"
          @click="sendMessage"
          class="unified-button"
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

/**
 * 保存消息到会话
 */
const saveMessages = () => {
  if (currentSessionId.value) {
    sessionStore.saveMessages(currentSessionId.value, messages.value)
  }
}

/**
 * 加载消息
 */
const loadMessages = () => {
  if (currentSessionId.value) {
    messages.value = sessionStore.getMessages(currentSessionId.value)
  } else {
    messages.value = []
  }
}

/**
 * 格式化消息内容
 */
const formatMessage = (text) => {
  return text.replace(/\n/g, '<br>')
}

/**
 * 连接WebSocket
 */
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

/**
 * 处理WebSocket消息
 */
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
 */
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

/**
 * 使用HTTP发送消息
 */
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

/**
 * 处理回车键
 */
const handleEnter = (event) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}

/**
 * 清空对话
 */
const clearChat = () => {
  messages.value = []
  saveMessages()
}

/**
 * 加载模型列表
 */
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

/**
 * 切换搜索
 */
const toggleSearch = () => {
  showSearch.value = !showSearch.value
}

/**
 * 执行搜索
 */
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

/**
 * 处理角色切换
 */
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

/**
 * 滚动到底部
 */
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
.chat-container {
  display: flex;
  flex-direction: column;
}

.chat-header {
  flex-shrink: 0;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.message {
  display: flex;
  gap: var(--spacing-md);
  align-items: flex-start;
}

.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--gradient-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.message-content {
  flex: 1;
}

.message-role {
  font-weight: var(--font-weight-semibold);
  margin-bottom: var(--spacing-xs);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.message-text {
  background: var(--bg-tertiary);
  padding: var(--spacing-md) var(--spacing-lg);
  border-radius: var(--radius-lg);
  line-height: var(--line-height-relaxed);
  color: var(--text-primary);
  white-space: pre-wrap;
}

.message-time {
  text-align: right;
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  margin-top: var(--spacing-xs);
}

.loading-message {
  display: flex;
  gap: var(--spacing-md);
  align-items: flex-start;
}

.typing-indicator {
  display: flex;
  gap: var(--spacing-xs);
  align-items: center;
  height: 24px;
  padding: var(--spacing-md) var(--spacing-lg);
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  background: var(--color-primary);
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

.search-input-area,
.chat-input-area {
  padding: var(--spacing-xl);
  border-top: 1px solid var(--border-color);
  background: var(--bg-elevated);
}

.input-controls {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-md);
}

.search-results {
  margin-top: var(--spacing-md);
}

.search-result-item {
  margin-bottom: var(--spacing-md);
  padding: var(--spacing-md);
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  
  a {
    color: var(--color-primary);
    font-weight: var(--font-weight-medium);
    text-decoration: none;
    
    &:hover {
      text-decoration: underline;
    }
  }
  
  p {
    margin: var(--spacing-xs) 0 0;
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }
}
</style>
