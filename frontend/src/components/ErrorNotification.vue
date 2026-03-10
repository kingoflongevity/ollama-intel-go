<template>
  <div class="error-notification-container">
    <TransitionGroup name="error-list" tag="div" class="error-list">
      <div
        v-for="error in errors"
        :key="error.id"
        class="error-item"
        :class="error.type"
      >
        <div class="error-icon">
          <el-icon v-if="error.type === 'error'"><CircleCloseFilled /></el-icon>
          <el-icon v-else-if="error.type === 'warning'"><WarningFilled /></el-icon>
          <el-icon v-else><InfoFilled /></el-icon>
        </div>
        <div class="error-content">
          <div class="error-title">{{ error.title }}</div>
          <div class="error-message">{{ error.message }}</div>
          <div class="error-actions" v-if="error.retryable">
            <el-button size="small" @click="retryError(error)">
              <el-icon><RefreshRight /></el-icon>
              重试
            </el-button>
          </div>
        </div>
        <el-button
          class="error-close"
          link
          size="small"
          @click="dismissError(error.id)"
        >
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { CircleCloseFilled, WarningFilled, InfoFilled, RefreshRight, Close } from '@element-plus/icons-vue'

// 错误列表
const errors = ref([])
let errorIdCounter = 0

/**
 * 添加错误通知
 * @param {Object} error - 错误对象
 * @param {string} error.title - 错误标题
 * @param {string} error.message - 错误消息
 * @param {string} error.type - 错误类型: 'error' | 'warning' | 'info'
 * @param {boolean} error.retryable - 是否可重试
 * @param {Function} error.retryCallback - 重试回调函数
 * @param {number} error.duration - 显示时长（毫秒），0 表示不自动关闭
 */
const addError = (error) => {
  const errorItem = {
    id: `error_${++errorIdCounter}`,
    title: error.title || '错误',
    message: error.message || '发生了一个未知错误',
    type: error.type || 'error',
    retryable: error.retryable || false,
    retryCallback: error.retryCallback || null,
    duration: error.duration !== undefined ? error.duration : 5000
  }
  
  errors.value.push(errorItem)
  
  // 自动关闭
  if (errorItem.duration > 0) {
    setTimeout(() => {
      dismissError(errorItem.id)
    }, errorItem.duration)
  }
  
  return errorItem.id
}

/**
 * 移除错误通知
 */
const dismissError = (errorId) => {
  const index = errors.value.findIndex(e => e.id === errorId)
  if (index > -1) {
    errors.value.splice(index, 1)
  }
}

/**
 * 重试错误操作
 */
const retryError = (error) => {
  if (error.retryCallback && typeof error.retryCallback === 'function') {
    error.retryCallback()
  }
  dismissError(error.id)
}

/**
 * 清除所有错误
 */
const clearAllErrors = () => {
  errors.value = []
}

/**
 * 处理全局错误事件
 */
const handleGlobalError = (event) => {
  const { detail } = event
  addError({
    title: detail.title || '操作失败',
    message: detail.message || '请稍后重试',
    type: detail.type || 'error',
    retryable: detail.retryable || false,
    retryCallback: detail.retryCallback || null,
    duration: detail.duration || 5000
  })
}

/**
 * 处理 API 错误
 */
const handleApiError = (error, retryCallback = null) => {
  let message = '请求失败，请稍后重试'
  
  if (error.response) {
    // HTTP 错误
    const status = error.response.status
    switch (status) {
      case 400:
        message = '请求参数错误'
        break
      case 401:
        message = '未授权，请重新登录'
        break
      case 403:
        message = '没有权限访问该资源'
        break
      case 404:
        message = '请求的资源不存在'
        break
      case 500:
        message = '服务器内部错误'
        break
      case 502:
        message = '网关错误'
        break
      case 503:
        message = '服务暂时不可用'
        break
      default:
        message = `请求失败 (${status})`
    }
  } else if (error.message) {
    // 网络错误或其他错误
    if (error.message.includes('timeout')) {
      message = '请求超时，请检查网络连接'
    } else if (error.message.includes('Network Error')) {
      message = '网络错误，请检查网络连接'
    } else {
      message = error.message
    }
  }
  
  return addError({
    title: 'API 错误',
    message,
    type: 'error',
    retryable: !!retryCallback,
    retryCallback
  })
}

// 暴露方法给外部调用
defineExpose({
  addError,
  dismissError,
  clearAllErrors,
  handleApiError
})

// 组件挂载时监听全局错误事件
onMounted(() => {
  window.addEventListener('appError', handleGlobalError)
})

// 组件卸载时移除事件监听
onUnmounted(() => {
  window.removeEventListener('appError', handleGlobalError)
})
</script>

<style scoped>
.error-notification-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  max-width: 400px;
  pointer-events: none;
}

.error-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.error-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  pointer-events: auto;
  transition: all 0.3s ease;
}

body.dark-theme .error-item {
  background: #2d2d2d;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.error-item.error {
  border-left: 4px solid #f56c6c;
}

.error-item.warning {
  border-left: 4px solid #e6a23c;
}

.error-item.info {
  border-left: 4px solid #409eff;
}

.error-icon {
  flex-shrink: 0;
  font-size: 20px;
}

.error-item.error .error-icon {
  color: #f56c6c;
}

.error-item.warning .error-icon {
  color: #e6a23c;
}

.error-item.info .error-icon {
  color: #409eff;
}

.error-content {
  flex: 1;
  min-width: 0;
}

.error-title {
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

body.dark-theme .error-title {
  color: #e4e6eb;
}

.error-message {
  font-size: 14px;
  color: #606266;
  word-break: break-word;
}

body.dark-theme .error-message {
  color: #a0aec0;
}

.error-actions {
  margin-top: 12px;
}

.error-close {
  flex-shrink: 0;
  color: #909399;
}

body.dark-theme .error-close {
  color: #718096;
}

.error-close:hover {
  color: #606266;
}

body.dark-theme .error-close:hover {
  color: #a0aec0;
}

/* 过渡动画 */
.error-list-enter-active,
.error-list-leave-active {
  transition: all 0.3s ease;
}

.error-list-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.error-list-leave-to {
  opacity: 0;
  transform: translateX(100%);
}

.error-list-move {
  transition: transform 0.3s ease;
}
</style>
