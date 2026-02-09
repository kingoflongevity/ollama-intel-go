<template>
  <div class="card">
    <h3>快捷操作</h3>
    <div class="actions">
      <button class="btn primary" @click="startService">启动服务</button>
      <button class="btn danger" @click="stopService">停止服务</button>
      <button class="btn ghost">刷新模型</button>
    </div>
  </div>
</template>

<script setup>
import { StartService, StopService } from '../../wailsjs/go/main/App'
import { ElMessage } from 'element-plus'

// 方法
const startService = async () => {
  try {
    await StartService()
    ElMessage.success('服务启动中...')
  } catch (error) {
    console.error('启动服务失败:', error)
    ElMessage.error('启动服务失败')
  }
}

const stopService = async () => {
  try {
    await StopService()
    ElMessage.success('服务停止中...')
  } catch (error) {
    console.error('停止服务失败:', error)
    ElMessage.error('停止服务失败')
  }
}
</script>

<style scoped>
/* 通用卡片样式 */
.card {
  background: rgba(18, 22, 32, 0.75);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  padding: 24px;
  border: 1px solid var(--glass-border);
  box-shadow: 0 0 40px rgba(76,130,255,0.08);
}

h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 20px;
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
}

.actions {
  margin-top: 20px;
  display: flex;
  gap: 16px;
}

.btn {
  height: 40px;
  padding: 0 20px;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  font-weight: 500;
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
  transition: all 0.3s ease;
}

.btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.3);
}

.primary {
  background: var(--accent-gradient);
  color: #fff;
}

.danger {
  background: linear-gradient(135deg,#FF4D4F,#FF7875);
  color: #fff;
}

.ghost {
  background: rgba(255,255,255,0.06);
  color: var(--text-primary);
}

/* 浅色主题特殊样式 */
body.light-theme .card {
  background: rgba(255, 255, 255, 0.75);
  box-shadow: 0 0 40px rgba(76,130,255,0.12);
}

body.light-theme .ghost {
  background: rgba(0,0,0,0.06);
}

body.light-theme .btn:hover {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
}
</style>