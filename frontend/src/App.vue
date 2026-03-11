<template>
  <div class="tech-app-wrapper">
    <ErrorNotification ref="errorNotificationRef" />
    
    <header class="tech-header">
      <div class="header-left">
        <div class="logo-icon">
          <el-icon size="22"><Cpu /></el-icon>
        </div>
        <div class="logo-text">
          <span class="logo-title">Ollama 英特尔优化版</span>
          <span class="logo-subtitle">Intel Optimized</span>
        </div>
      </div>
      <div class="header-right">
        <button class="header-btn" @click="toggleTheme" :title="isDark ? '切换到亮色模式' : '切换到暗色模式'">
          <el-icon v-if="isDark" size="18"><Sunny /></el-icon>
          <el-icon v-else size="18"><Moon /></el-icon>
        </button>
        <button class="header-btn" @click="showSettings" title="设置">
          <el-icon size="18"><Setting /></el-icon>
        </button>
      </div>
    </header>

    <div class="tech-app-container">
      <aside class="tech-sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">导航菜单</span>
        </div>
        <nav class="sidebar-nav">
          <router-link 
            v-for="item in menuItems" 
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.path === item.path }"
          >
            <div class="nav-icon">
              <el-icon size="18"><component :is="item.icon" /></el-icon>
            </div>
            <span class="nav-label">{{ item.label }}</span>
            <div class="nav-indicator"></div>
          </router-link>
        </nav>
        <div class="sidebar-footer">
          <div class="version-info">
            <span>v1.0.0</span>
          </div>
        </div>
      </aside>

      <main class="tech-main-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  Cpu, Moon, Sunny, Setting, House, 
  ChatLineRound, ChatDotRound, Box, Connection,
  Document, Download
} from '@element-plus/icons-vue'
import ErrorNotification from '@/components/ErrorNotification.vue'

const router = useRouter()
const isDark = ref(true)
const errorNotificationRef = ref(null)

const menuItems = [
  { path: '/dashboard', label: '仪表盘', icon: House },
  { path: '/chat', label: 'AI 对话', icon: ChatLineRound },
  { path: '/chat-sessions', label: '会话管理', icon: ChatDotRound },
  { path: '/models', label: '本地模型', icon: Box },
  { path: '/online', label: '在线模型', icon: Download },
  { path: '/logs', label: '系统日志', icon: Document },
  { path: '/settings', label: '系统设置', icon: Setting },
]

const toggleTheme = () => {
  isDark.value = !isDark.value
  document.body.classList.toggle('dark-theme', isDark.value)
  document.body.classList.toggle('light-theme', !isDark.value)
}

const showSettings = () => {
  router.push('/settings')
}

onMounted(() => {
  document.body.classList.add('dark-theme')
})
</script>

<style lang="scss">
.tech-app-wrapper {
  height: 100vh;
  width: 100vw;
  display: flex;
  flex-direction: column;
  background: #0a0a0f;
  position: relative;
  overflow: hidden;
  margin: 0;
  padding: 0;
}

.tech-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 24px;
  background: rgba(15, 15, 25, 0.8);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(6, 182, 212, 0.2);
  position: relative;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 14px;
}

.logo-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #06b6d4 0%, #8b5cf6 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 4px 15px rgba(6, 182, 212, 0.3);
}

.logo-text {
  display: flex;
  flex-direction: column;
}

.logo-title {
  font-size: 16px;
  font-weight: 600;
  color: #e2e8f0;
  letter-spacing: 0.5px;
}

.logo-subtitle {
  font-size: 10px;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.header-right {
  display: flex;
  gap: 8px;
}

.header-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(51, 65, 85, 0.5);
  border-radius: 10px;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.3s ease;
}

.header-btn:hover {
  background: rgba(6, 182, 212, 0.15);
  border-color: rgba(6, 182, 212, 0.4);
  color: #06b6d4;
  transform: translateY(-1px);
}

.tech-app-container {
  display: flex;
  flex: 1;
  overflow: hidden;
  position: relative;
  z-index: 1;
}

.tech-sidebar {
  width: 220px;
  display: flex;
  flex-direction: column;
  background: rgba(15, 15, 25, 0.6);
  border-right: 1px solid rgba(6, 182, 212, 0.15);
  position: relative;
  z-index: 10;
}

.tech-sidebar::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 1px;
  background: linear-gradient(180deg, 
    transparent 0%, 
    rgba(6, 182, 212, 0.3) 20%, 
    rgba(139, 92, 246, 0.3) 50%, 
    rgba(6, 182, 212, 0.3) 80%, 
    transparent 100%
  );
}

.sidebar-header {
  padding: 20px 20px 16px;
  border-bottom: 1px solid rgba(51, 65, 85, 0.3);
}

.sidebar-title {
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 1.5px;
}

.sidebar-nav {
  flex: 1;
  padding: 12px 10px;
  overflow-y: auto;
}

.sidebar-nav::-webkit-scrollbar {
  width: 4px;
}

.sidebar-nav::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-nav::-webkit-scrollbar-thumb {
  background: rgba(6, 182, 212, 0.2);
  border-radius: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  margin-bottom: 4px;
  border-radius: 10px;
  color: #94a3b8;
  text-decoration: none;
  position: relative;
  transition: all 0.3s ease;
  cursor: pointer;
}

.nav-item:hover {
  background: rgba(30, 41, 59, 0.6);
  color: #e2e8f0;
}

.nav-item.active {
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.15) 0%, rgba(139, 92, 246, 0.1) 100%);
  color: #06b6d4;
  box-shadow: 
    inset 0 1px 0 rgba(6, 182, 212, 0.2),
    0 4px 12px rgba(6, 182, 212, 0.15);
}

.nav-item.active .nav-icon {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(6, 182, 212, 0.4);
}

.nav-item.active .nav-indicator {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  background: linear-gradient(180deg, #06b6d4, #8b5cf6);
  border-radius: 0 2px 2px 0;
}

.nav-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 8px;
  transition: all 0.3s ease;
}

.nav-label {
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.3px;
}

.sidebar-footer {
  padding: 16px 20px;
  border-top: 1px solid rgba(51, 65, 85, 0.3);
}

.version-info {
  display: flex;
  align-items: center;
  justify-content: center;
}

.version-info span {
  font-size: 11px;
  color: #475569;
  padding: 4px 12px;
  background: rgba(30, 41, 59, 0.4);
  border-radius: 20px;
  border: 1px solid rgba(51, 65, 85, 0.3);
}

.tech-main-content {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.tech-main-content::-webkit-scrollbar {
  width: 6px;
}

.tech-main-content::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.5);
}

.tech-main-content::-webkit-scrollbar-thumb {
  background: rgba(6, 182, 212, 0.3);
  border-radius: 3px;
}
</style>
