<template>
  <div class="app-wrapper">
    <!-- 错误通知组件 -->
    <ErrorNotification ref="errorNotificationRef" />
    
    <!-- 顶部导航栏 -->
    <el-header class="app-header">
      <div class="header-content">
        <div class="logo-section">
          <el-icon size="24"><Monitor /></el-icon>
          <span class="logo-text">Ollama 英特尔优化版</span>
        </div>
        <div class="header-actions">
          <el-button link @click="toggleTheme" class="header-btn">
            <el-icon v-if="isDark"><Sunny /></el-icon>
            <el-icon v-else><Moon /></el-icon>
          </el-button>
          <el-button link @click="showSettings" class="header-btn">
            <el-icon><Setting /></el-icon>
          </el-button>
        </div>
      </div>
    </el-header>

    <div class="app-container">
      <!-- 侧边栏 -->
      <el-aside width="260px" class="sidebar">
        <el-menu
          :default-active="$route.path"
          :router="true"
          class="sidebar-menu"
          :collapse="false"
        >
          <el-menu-item index="/dashboard" class="menu-item">
            <el-icon><House /></el-icon>
            <span>仪表板</span>
          </el-menu-item>
          <el-menu-item index="/chat" class="menu-item">
            <el-icon><ChatLineRound /></el-icon>
            <span>聊天</span>
          </el-menu-item>
          <el-menu-item index="/chat-sessions" class="menu-item">
            <el-icon><ChatDotRound /></el-icon>
            <span>会话管理</span>
          </el-menu-item>
          <el-menu-item index="/models" class="menu-item">
            <el-icon><Files /></el-icon>
            <span>模型管理</span>
          </el-menu-item>
          <el-menu-item index="/online" class="menu-item">
            <el-icon><Connection /></el-icon>
            <span>在线模型</span>
          </el-menu-item>
          <el-menu-item index="/logs" class="menu-item">
            <el-icon><Document /></el-icon>
            <span>日志管理</span>
          </el-menu-item>
          <el-menu-item index="/settings" class="menu-item">
            <el-icon><Setting /></el-icon>
            <span>设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  Monitor, Moon, Sunny, Setting, House, 
  ChatLineRound, ChatDotRound, Files, Connection,
  Document
} from '@element-plus/icons-vue'
import ErrorNotification from '@/components/ErrorNotification.vue'

const router = useRouter()
const isDark = ref(true)
const errorNotificationRef = ref(null)

/**
 * 切换主题 - 在深色和浅色主题之间切换
 */
const toggleTheme = () => {
  isDark.value = !isDark.value
  document.body.classList.toggle('dark-theme', isDark.value)
  document.body.classList.toggle('light-theme', !isDark.value)
}

/**
 * 显示设置页面
 */
const showSettings = () => {
  router.push('/settings')
}

onMounted(() => {
  document.body.classList.add('dark-theme')
})
</script>

<style lang="scss">
.app-wrapper {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-primary);
  background-image: 
    radial-gradient(circle at 25% 25%, rgba(99, 102, 241, 0.05) 0%, transparent 50%),
    radial-gradient(circle at 75% 75%, rgba(168, 85, 247, 0.05) 0%, transparent 50%);
}

.app-header {
  padding: 0;
  height: 70px;
  margin: var(--spacing-lg);
  margin-bottom: 0;
  border-radius: var(--radius-xl);
  border: 1px solid var(--border-color);
  background: var(--bg-elevated);
  backdrop-filter: blur(20px);
  box-shadow: var(--shadow);

  .header-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 100%;
    padding: 0 var(--spacing-2xl);

    .logo-section {
      display: flex;
      align-items: center;
      gap: var(--spacing-lg);

      .logo-text {
        font-size: var(--font-size-2xl);
        font-weight: var(--font-weight-bold);
        background: var(--gradient-primary);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
      }

      .el-icon {
        color: var(--color-primary);
      }
    }

    .header-actions {
      display: flex;
      gap: var(--spacing-lg);

      .header-btn {
        color: var(--text-primary);
        font-size: 18px;
        transition: all var(--transition-base);

        &:hover {
          color: var(--color-primary);
          transform: translateY(-1px);
        }
      }
    }
  }
}

.app-container {
  display: flex;
  flex: 1;
  overflow: hidden;
  padding: var(--spacing-lg);
  gap: var(--spacing-lg);
}

.sidebar {
  border-radius: var(--radius-2xl);
  border: 1px solid var(--border-color);
  height: calc(100vh - 122px);
  overflow: hidden;
  background: var(--bg-elevated);
  backdrop-filter: blur(16px);
  box-shadow: var(--shadow);
  position: relative;

  /* 边缘渐变发光效果 */
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    border-radius: var(--radius-2xl);
    padding: 1px;
    background: var(--gradient-primary);
    -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
    -webkit-mask-composite: xor;
    mask-composite: exclude;
    pointer-events: none;
    opacity: 0.3;
  }

  .sidebar-menu {
    border: none;
    height: 100%;
    background: transparent;
    padding: var(--spacing-lg) 0;

    .menu-item {
      height: 56px;
      line-height: 56px;
      margin: var(--spacing-xs) var(--spacing-md);
      border-radius: var(--radius-lg);
      color: var(--text-secondary);
      transition: all var(--transition-base);
      position: relative;
      overflow: hidden;
      padding-left: var(--spacing-2xl);
      font-weight: var(--font-weight-medium);

      /* 选中项样式 */
      &.is-active {
        background: var(--gradient-primary-light);
        color: var(--color-primary);
        box-shadow: 
          inset 0 1px 0 rgba(6, 182, 212, 0.3),
          0 4px 16px rgba(6, 182, 212, 0.3);
        border-left: 3px solid var(--color-primary);
        padding-left: calc(var(--spacing-2xl) - 3px);
      }

      &:hover {
        background: rgba(6, 182, 212, 0.1);
        color: var(--text-primary);
        box-shadow: 0 2px 8px rgba(6, 182, 212, 0.2);
      }

      .el-icon {
        font-size: 18px;
        margin-right: var(--spacing-lg);
        transition: all var(--transition-base);
      }

      span {
        font-size: var(--font-size-base);
        letter-spacing: 0.02em;
      }
    }
  }
}

.main-content {
  padding: var(--spacing-2xl);
  overflow-y: auto;
  background: transparent;
  flex: 1;
  border-radius: var(--radius-xl);
}
</style>
