<template>
  <div class="app-wrapper">
    <!-- 顶部导航栏 -->
    <el-header class="app-header glass-card">
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
      <el-aside width="260px" class="sidebar glass-card glow-effect gradient-border">
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
  ChatLineRound, Files, Connection,
  Document
} from '@element-plus/icons-vue'

const router = useRouter()
const isDark = ref(true) // 默认使用深色主题

const toggleTheme = () => {
  isDark.value = !isDark.value
  document.body.classList.toggle('dark-theme', isDark.value)
  document.body.classList.toggle('light-theme', !isDark.value)
}

const showSettings = () => {
  router.push('/settings')
}

onMounted(() => {
  // 初始化主题为深色
  document.body.classList.add('dark-theme')
})
</script>

<style lang="scss">
/* 全局主题变量 */
:root {
  /* 深色主题变量 */
  --dark-bg-primary: #0b0f17;
  --dark-bg-secondary: #1a2338;
  --dark-text-primary: #E6EAF2;
  --dark-text-secondary: rgba(230, 234, 242, 0.7);
  --dark-accent-primary: #06b6d4;
  --dark-accent-secondary: #8b5cf6;
  --dark-glass-border: rgba(255, 255, 255, 0.06);
  
  /* 浅色主题变量 */
  --light-bg-primary: #f8fafc;
  --light-bg-secondary: #ffffff;
  --light-text-primary: #1e293b;
  --light-text-secondary: rgba(30, 41, 59, 0.7);
  --light-accent-primary: #0284c7;
  --light-accent-secondary: #7c3aed;
  --light-glass-border: rgba(0, 0, 0, 0.06);
}

/* 深色主题 */
body.dark-theme {
  background: radial-gradient(circle at 20% 10%, #1a2338, #0b0f17 60%);
  color: var(--dark-text-primary);
  font-family: Inter, "PingFang SC", sans-serif;
  
  --bg-primary: var(--dark-bg-primary);
  --bg-secondary: var(--dark-bg-secondary);
  --text-primary: var(--dark-text-primary);
  --text-secondary: var(--dark-text-secondary);
  --accent-primary: var(--dark-accent-primary);
  --accent-secondary: var(--dark-accent-secondary);
  --glass-border: var(--dark-glass-border);
  --accent-gradient: linear-gradient(135deg, var(--dark-accent-primary), var(--dark-accent-secondary));
}

/* 浅色主题 */
body.light-theme {
  background: radial-gradient(circle at 20% 10%, #f1f5f9, #ffffff 60%);
  color: var(--light-text-primary);
  font-family: Inter, "PingFang SC", sans-serif;
  
  --bg-primary: var(--light-bg-primary);
  --bg-secondary: var(--light-bg-secondary);
  --text-primary: var(--light-text-primary);
  --text-secondary: var(--light-text-secondary);
  --accent-primary: var(--light-accent-primary);
  --accent-secondary: var(--light-accent-secondary);
  --glass-border: var(--light-glass-border);
  --accent-gradient: linear-gradient(135deg, var(--light-accent-primary), var(--light-accent-secondary));
}

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
  margin: 16px;
  margin-bottom: 0;
  border-radius: 16px;
  border: 1px solid var(--glass-border);
  background: rgba(18, 30, 58, 0.9);
  backdrop-filter: blur(20px);

  .header-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 100%;
    padding: 0 24px;

    .logo-section {
      display: flex;
      align-items: center;
      gap: 16px;

      .logo-text {
        font-size: 20px;
        font-weight: 700;
        background: var(--accent-gradient);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
      }

      .el-icon {
        color: var(--accent-primary);
      }
    }

    .header-actions {
      display: flex;
      gap: 16px;

      .header-btn {
        color: var(--text-primary);
        font-size: 18px;

        &:hover {
          color: var(--accent-primary);
        }
      }
    }
  }
}

.app-container {
  display: flex;
  flex: 1;
  overflow: hidden;
  padding: 16px;
  gap: 16px;
}

.sidebar {
  border-radius: 18px;
  border: 1px solid var(--glass-border);
  height: calc(100vh - 122px);
  overflow: hidden;
  background: rgba(18, 30, 58, 0.9);
  backdrop-filter: blur(16px);
  position: relative;

  /* 边缘渐变发光效果 */
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    border-radius: 18px;
    padding: 1px;
    background: linear-gradient(135deg, rgba(6, 182, 212, 0.4), rgba(139, 92, 246, 0.4), rgba(6, 182, 212, 0.4));
    -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
    -webkit-mask-composite: xor;
    mask-composite: exclude;
    pointer-events: none;
  }

  .sidebar-menu {
    border: none;
    height: 100%;
    background: transparent;
    padding: 16px 0;

    .menu-item {
      height: 60px;
      line-height: 60px;
      margin: 6px 12px;
      border-radius: 14px;
      color: var(--text-secondary);
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;
      padding-left: 24px;
      box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05);

      /* 选中项蓝色霓虹指示条 */
      &.is-active {
        background: linear-gradient(90deg, rgba(6, 182, 212, 0.2), rgba(139, 92, 246, 0.1));
        color: var(--accent-primary);
        box-shadow: 
          inset 0 1px 0 rgba(6, 182, 212, 0.3),
          0 4px 16px rgba(6, 182, 212, 0.3),
          inset 3px 0 0 var(--accent-primary);
        border-left: 3px solid var(--accent-primary);
        animation: pulse 2s infinite;
      }

      &:hover {
        background: rgba(6, 182, 212, 0.15);
        color: var(--text-primary);
        box-shadow: 
          inset 0 1px 0 rgba(6, 182, 212, 0.2),
          0 2px 8px rgba(6, 182, 212, 0.2);
      }

      .el-icon {
        font-size: 20px;
        margin-right: 16px;
        transition: all 0.3s ease;
      }

      span {
        font-size: 15px;
        font-weight: 500;
        letter-spacing: 0.02em;
      }
    }
  }
}

/* 脉冲动画 */
@keyframes pulse {
  0% {
    box-shadow: 
      inset 0 1px 0 rgba(6, 182, 212, 0.3),
      0 4px 16px rgba(6, 182, 212, 0.3),
      inset 3px 0 0 var(--accent-primary);
  }
  50% {
    box-shadow: 
      inset 0 1px 0 rgba(6, 182, 212, 0.4),
      0 6px 20px rgba(6, 182, 212, 0.4),
      inset 3px 0 0 var(--accent-primary);
  }
  100% {
    box-shadow: 
      inset 0 1px 0 rgba(6, 182, 212, 0.3),
      0 4px 16px rgba(6, 182, 212, 0.3),
      inset 3px 0 0 var(--accent-primary);
  }
}

.main-content {
  padding: 24px;
  overflow-y: auto;
  background: transparent;
  flex: 1;
  border-radius: 16px;
}

/* 滚动条样式 */
.main-content::-webkit-scrollbar {
  width: 8px;
}

.main-content::-webkit-scrollbar-track {
  background: var(--bg-secondary);
  border-radius: 4px;
}

.main-content::-webkit-scrollbar-thumb {
  background: var(--border-color, var(--glass-border));
  border-radius: 4px;
}

.main-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-secondary);
}

/* 浅色主题特殊样式 */
body.light-theme .app-header,
body.light-theme .sidebar {
  background: rgba(255, 255, 255, 0.9);
}

body.light-theme .sidebar .menu-item {
  box-shadow: inset 0 1px 0 rgba(0, 0, 0, 0.05);
}

body.light-theme .sidebar .menu-item.is-active {
  background: linear-gradient(90deg, rgba(2, 132, 199, 0.2), rgba(124, 58, 237, 0.1));
  box-shadow: 
    inset 0 1px 0 rgba(2, 132, 199, 0.3),
    0 4px 16px rgba(2, 132, 199, 0.3),
    inset 3px 0 0 var(--accent-primary);
}

body.light-theme .sidebar .menu-item:hover {
  background: rgba(2, 132, 199, 0.15);
  box-shadow: 
    inset 0 1px 0 rgba(2, 132, 199, 0.2),
    0 2px 8px rgba(2, 132, 199, 0.2);
}
</style>