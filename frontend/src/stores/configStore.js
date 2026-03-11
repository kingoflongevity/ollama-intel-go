import { ref, watch } from 'vue'

const CONFIG_KEY = 'ollama_config'

const defaultConfig = {
  ollamaHost: '127.0.0.1:11434',
  wsPort: 11435,
  theme: 'dark',
  language: 'zh-CN'
}

const config = ref({ ...defaultConfig })

/**
 * 加载配置
 */
const loadConfig = () => {
  try {
    const saved = localStorage.getItem(CONFIG_KEY)
    if (saved) {
      const parsed = JSON.parse(saved)
      config.value = { ...defaultConfig, ...parsed }
    }
  } catch (error) {
    console.error('加载配置失败:', error)
  }
}

/**
 * 保存配置
 */
const saveConfig = () => {
  try {
    localStorage.setItem(CONFIG_KEY, JSON.stringify(config.value))
  } catch (error) {
    console.error('保存配置失败:', error)
  }
}

/**
 * 获取 Ollama API 地址
 */
const getOllamaApiUrl = () => {
  return `http://${config.value.ollamaHost}`
}

/**
 * 获取 WebSocket 地址
 */
const getWebSocketUrl = () => {
  const host = config.value.ollamaHost.split(':')[0]
  return `ws://${host}:${config.value.wsPort}/ws/chat`
}

/**
 * 更新配置
 */
const updateConfig = (newConfig) => {
  config.value = { ...config.value, ...newConfig }
  saveConfig()
}

/**
 * 重置配置
 */
const resetConfig = () => {
  config.value = { ...defaultConfig }
  saveConfig()
}

// 监听配置变化自动保存
watch(config, saveConfig, { deep: true })

// 初始化时加载配置
loadConfig()

export function useConfigStore() {
  return {
    config,
    loadConfig,
    saveConfig,
    updateConfig,
    resetConfig,
    getOllamaApiUrl,
    getWebSocketUrl
  }
}
