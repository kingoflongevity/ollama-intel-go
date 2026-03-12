<template>
  <div class="tech-models-container">
    <div class="tech-page-header">
      <div class="header-left">
        <div class="header-title">
          <div class="title-icon">
            <el-icon><Box /></el-icon>
          </div>
          <div class="title-text">
            <h2>模型管理中心</h2>
            <span class="subtitle">Model Management System</span>
          </div>
        </div>
        <div class="model-stats">
          <div class="stat-item">
            <span class="stat-value">{{ models.length }}</span>
            <span class="stat-label">本地模型</span>
          </div>
        </div>
      </div>
      <div class="header-right">
        <div class="search-box">
          <el-icon class="search-icon"><Search /></el-icon>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索模型..."
            class="tech-search-input"
          />
        </div>
        <button class="tech-btn tech-btn-primary" @click="showPullDialog">
          <el-icon><Download /></el-icon>
          拉取模型
        </button>
        <button class="tech-btn" @click="refreshModels" :disabled="loading">
          <el-icon :class="{ 'is-loading': loading }"><Refresh /></el-icon>
          刷新
        </button>
      </div>
    </div>

    <div class="models-content">
      <div v-if="filteredModels.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">
          <el-icon><Box /></el-icon>
        </div>
        <h3>暂无本地模型</h3>
        <p>点击"拉取模型"按钮从在线模型库下载模型</p>
        <button class="tech-btn tech-btn-primary" @click="showPullDialog">
          <el-icon><Download /></el-icon>
          拉取模型
        </button>
      </div>

      <div v-else class="models-grid">
        <div v-for="model in filteredModels" :key="model.name" class="model-card">
          <div class="model-card-header">
            <div class="model-icon">
              <el-icon><Box /></el-icon>
            </div>
            <div class="model-info">
              <h4 class="model-name">{{ model.name }}</h4>
              <span class="model-id">ID: {{ getModelId(model.digest) }}</span>
              <span class="model-size">{{ model.size }}</span>
            </div>
          </div>
          
          <div class="model-card-body">
            <div class="model-detail">
              <span class="detail-label">参数规模</span>
              <span class="detail-value">{{ model.details?.parameter_size || '-' }}</span>
            </div>
            <div class="model-detail">
              <span class="detail-label">量化级别</span>
              <span class="detail-value">{{ model.details?.quantization_level || '-' }}</span>
            </div>
            <div class="model-detail">
              <span class="detail-label">格式</span>
              <span class="detail-value">{{ model.details?.format || '-' }}</span>
            </div>
            <div class="model-detail">
              <span class="detail-label">修改时间</span>
              <span class="detail-value">{{ formatDate(model.modified_at) }}</span>
            </div>
          </div>

          <div class="model-card-footer">
            <button class="tech-btn tech-btn-sm" @click="showModelInfo(model)">
              <el-icon><InfoFilled /></el-icon>
              详情
            </button>
            <button class="tech-btn tech-btn-sm tech-btn-danger" @click="deleteModel(model)">
              <el-icon><Delete /></el-icon>
              删除
            </button>
          </div>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="pullDialogVisible"
      title=""
      width="520px"
      :close-on-click-modal="!pullLoading"
      :close-on-press-escape="!pullLoading"
      class="tech-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <div class="dialog-icon">
            <el-icon><Download /></el-icon>
          </div>
          <div class="dialog-title">
            <h3>拉取模型</h3>
            <span>从模型库下载新模型</span>
          </div>
        </div>
      </template>
      
      <div class="dialog-content">
        <template v-if="!pullLoading">
          <div class="form-group">
            <label class="form-label">模型名称</label>
            <el-autocomplete
              v-model="pullForm.name"
              :fetch-suggestions="queryModelSearch"
              placeholder="输入模型名称，如 llama3:8b"
              class="tech-autocomplete"
              clearable
            />
          </div>
          <div class="form-help">
            <el-icon><InfoFilled /></el-icon>
            <span>支持格式: model_name:tag，例如 llama3:8b, mistral:7b</span>
          </div>
        </template>
        
        <template v-else>
          <div class="pull-progress-section">
            <div class="progress-header">
              <span class="progress-title">{{ pullStatusText }}</span>
              <span class="progress-percent">{{ pullProgress }}%</span>
            </div>
            <div class="progress-bar-container">
              <div class="progress-bar" :style="{ width: pullProgress + '%' }" :class="pullStatus"></div>
            </div>
            <div class="progress-info">
              <div class="info-row">
                <span class="info-label">当前任务</span>
                <span class="info-value">{{ pullCurrentTask }}</span>
              </div>
              <div v-if="pullError" class="info-row error">
                <span class="info-label">错误信息</span>
                <span class="info-value">{{ pullError }}</span>
              </div>
            </div>
          </div>
        </template>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <button v-if="!pullLoading" class="tech-btn" @click="pullDialogVisible = false">取消</button>
          <button v-if="!pullLoading" class="tech-btn tech-btn-primary" @click="confirmPullModel">确认拉取</button>
          <button v-if="pullLoading" class="tech-btn tech-btn-danger" @click="cancelPull">取消拉取</button>
          <button v-if="pullStatus === 'success'" class="tech-btn tech-btn-success" @click="pullDialogVisible = false">完成</button>
          <button v-if="pullStatus === 'exception'" class="tech-btn" @click="pullDialogVisible = false">关闭</button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="infoDialogVisible"
      title=""
      width="520px"
      class="tech-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <div class="dialog-icon info">
            <el-icon><InfoFilled /></el-icon>
          </div>
          <div class="dialog-title">
            <h3>模型详情</h3>
            <span>查看模型完整信息</span>
          </div>
        </div>
      </template>
      
      <div v-if="selectedModel" class="dialog-content">
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">模型名称</span>
            <span class="info-value">{{ selectedModel.name }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">模型大小</span>
            <span class="info-value">{{ selectedModel.size }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">参数规模</span>
            <span class="info-value">{{ selectedModel.details?.parameter_size || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">量化级别</span>
            <span class="info-value">{{ selectedModel.details?.quantization_level || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">格式</span>
            <span class="info-value">{{ selectedModel.details?.format || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">模型家族</span>
            <span class="info-value">{{ getModelFamilies(selectedModel.details?.families) }}</span>
          </div>
          <div class="info-item full-width">
            <span class="info-label">摘要 (Digest)</span>
            <span class="info-value digest">{{ selectedModel.digest || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">修改时间</span>
            <span class="info-value">{{ formatDate(selectedModel.modified_at) }}</span>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <button class="tech-btn tech-btn-primary" @click="infoDialogVisible = false">关闭</button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Download, Refresh, Search, Box, InfoFilled, Delete } from '@element-plus/icons-vue'
import { ListModels, PullModel, DeleteModel, ShowModel } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { ElMessage, ElMessageBox } from 'element-plus'

const models = ref([])
const loading = ref(false)
const searchQuery = ref('')
const pullDialogVisible = ref(false)
const infoDialogVisible = ref(false)
const selectedModel = ref(null)
const pullForm = ref({ name: '' })
const pullLoading = ref(false)
const pullProgress = ref(0)
const pullStatus = ref('')
const pullStatusText = ref('')
const pullCurrentTask = ref('')
const pullError = ref('')
const currentPullModel = ref('')

const filteredModels = computed(() => {
  if (!searchQuery.value) {
    return models.value
  }
  return models.value.filter(model =>
    model.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN')
}

const getModelId = (digest) => {
  if (!digest) return '-'
  return digest.substring(0, 12)
}

const getModelFamilies = (families) => {
  if (!families || !Array.isArray(families)) return '-'
  return families.join(', ')
}

const cancelPull = () => {
  pullDialogVisible.value = false
  resetPullState()
}

const resetPullState = () => {
  pullLoading.value = false
  pullProgress.value = 0
  pullStatus.value = ''
  pullStatusText.value = ''
  pullCurrentTask.value = ''
  pullError.value = ''
  currentPullModel.value = ''
}

const setupPullProgressListener = () => {
  EventsOn('model_pull_progress', (eventData) => {
    console.log('收到进度事件:', eventData)
    const { model, status, progress, message, time } = eventData
    
    if (model !== currentPullModel.value) {
      return
    }
    
    // 更新进度（只有当 progress >= 0 时才更新）
    if (progress !== undefined && progress !== null && progress >= 0) {
      pullProgress.value = Math.round(progress)
    }
    
    switch (status) {
      case 'started':
        pullStatus.value = ''
        pullStatusText.value = '开始拉取'
        pullCurrentTask.value = message || '正在初始化...'
        break
      case 'downloading':
        pullStatus.value = ''
        pullStatusText.value = '下载中'
        pullCurrentTask.value = message || '正在下载...'
        break
      case 'status':
        // 状态更新，不改变进度
        pullCurrentTask.value = message || '处理中...'
        break
      case 'completed':
        pullStatus.value = 'success'
        pullStatusText.value = '拉取完成'
        pullCurrentTask.value = message || '模型拉取完成'
        pullProgress.value = 100
        loadModelsWithRetry()
        break
      case 'error':
        pullStatus.value = 'exception'
        pullStatusText.value = '拉取失败'
        pullCurrentTask.value = message || '拉取失败'
        pullError.value = message
        break
      default:
        // 未知状态，显示消息
        if (message) {
          pullCurrentTask.value = message
        }
    }
  })
}

const loadModels = async () => {
  loading.value = true
  try {
    const response = await ListModels()
    models.value = response
  } catch (error) {
    console.error('加载模型列表失败:', error)
    ElMessage.error('加载模型列表失败')
  } finally {
    loading.value = false
  }
}

const loadModelsWithRetry = async (maxRetries = 3, delay = 2000) => {
  let retryCount = 0
  
  const attemptLoad = async () => {
    try {
      retryCount++
      console.log(`尝试加载模型列表 (第 ${retryCount} 次)...`)
      await loadModels()
      console.log('模型列表加载成功')
      return true
    } catch (error) {
      console.error(`第 ${retryCount} 次加载失败:`, error)
      return false
    }
  }
  
  await new Promise(resolve => setTimeout(resolve, delay))
  
  const success = await attemptLoad()
  
  if (!success && retryCount < maxRetries) {
    for (let i = retryCount; i < maxRetries; i++) {
      const retryDelay = delay + (i * 1000)
      console.log(`将在 ${retryDelay}ms 后进行第 ${i + 1} 次重试...`)
      await new Promise(resolve => setTimeout(resolve, retryDelay))
      
      const retrySuccess = await attemptLoad()
      if (retrySuccess) {
        break
      }
    }
  }
}

const refreshModels = () => {
  loadModels()
}

const showPullDialog = () => {
  pullForm.value = { name: '' }
  pullDialogVisible.value = true
}

const queryModelSearch = (queryString, cb) => {
  const modelSuggestions = [
    { value: 'llama3:8b' },
    { value: 'llama3:70b' },
    { value: 'mistral:7b' },
    { value: 'gemma:2b' },
    { value: 'phi3:3.8b' },
    { value: 'qwen2:7b' },
    { value: 'mixtral:8x7b' },
  ]

  const results = queryString
    ? modelSuggestions.filter(item => 
        item.value.toLowerCase().includes(queryString.toLowerCase())
      )
    : modelSuggestions

  cb(results)
}

const confirmPullModel = async () => {
  if (!pullForm.value.name) {
    ElMessage.warning('请输入模型名称')
    return
  }

  const modelName = pullForm.value.name
  
  pullLoading.value = true
  pullStatusText.value = '开始拉取'
  pullCurrentTask.value = `准备拉取模型: ${modelName}`
  
  try {
    const result = await PullModel(modelName)
    // 使用后端返回的规范化模型名称
    if (result && result.model) {
      currentPullModel.value = result.model
      console.log('设置当前拉取模型:', result.model)
    }
  } catch (error) {
    console.error('拉取模型失败:', error)
    pullStatus.value = 'exception'
    pullStatusText.value = '拉取失败'
    pullError.value = error.message || '拉取模型时发生错误'
    ElMessage.error('拉取模型失败')
  }
}

const showModelInfo = async (model) => {
  try {
    const details = await ShowModel(model.name)
    selectedModel.value = { ...model, ...details }
    infoDialogVisible.value = true
  } catch (error) {
    console.error('获取模型详情失败:', error)
    selectedModel.value = model
    infoDialogVisible.value = true
  }
}

const deleteModel = async (model) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模型 "${model.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await DeleteModel(model.name)
    ElMessage.success('模型删除成功')
    loadModels()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除模型失败:', error)
      ElMessage.error('删除模型失败')
    }
  }
}

onMounted(() => {
  loadModels()
  setupPullProgressListener()
  
  window.addEventListener('localModelsUpdated', () => {
    console.log('收到本地模型更新事件，重新加载模型列表')
    loadModels()
  })
})
</script>

<style scoped>
.tech-models-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(135deg, #0a0a0f 0%, #1a1a2e 50%, #0f0f1a 100%);
  position: relative;
  overflow: hidden;
}

.tech-models-container::before {
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

.tech-page-header {
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
  gap: 24px;
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

.model-stats {
  display: flex;
  gap: 20px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 16px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 8px;
  border: 1px solid rgba(6, 182, 212, 0.2);
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: #06b6d4;
}

.stat-label {
  font-size: 11px;
  color: #64748b;
  text-transform: uppercase;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: #64748b;
  font-size: 16px;
}

.tech-search-input {
  width: 220px;
  padding: 8px 12px 8px 36px;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 8px;
  color: #e2e8f0;
  font-size: 13px;
  outline: none;
  transition: all 0.3s ease;
}

.tech-search-input::placeholder {
  color: #475569;
}

.tech-search-input:focus {
  border-color: rgba(6, 182, 212, 0.6);
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
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

.tech-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.tech-btn-danger {
  background: transparent;
  border: 1px solid rgba(239, 68, 68, 0.5);
  color: #ef4444;
}

.tech-btn-danger:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: #ef4444;
}

.tech-btn-success {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  border: none;
  color: white;
}

.tech-btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.models-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  position: relative;
  z-index: 1;
}

.models-content::-webkit-scrollbar {
  width: 6px;
}

.models-content::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.5);
}

.models-content::-webkit-scrollbar-thumb {
  background: rgba(6, 182, 212, 0.3);
  border-radius: 3px;
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
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

.empty-state h3 {
  margin: 0 0 8px 0;
  font-size: 20px;
  color: #e2e8f0;
}

.empty-state p {
  margin: 0 0 24px 0;
  color: #64748b;
}

.models-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.model-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(51, 65, 85, 0.5);
  border-radius: 16px;
  padding: 20px;
  transition: all 0.3s ease;
}

.model-card:hover {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(6, 182, 212, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

.model-card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(51, 65, 85, 0.5);
}

.model-icon {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
}

.model-info {
  flex: 1;
}

.model-name {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #e2e8f0;
}

.model-id {
  font-size: 11px;
  color: #06b6d4;
  font-family: monospace;
  background: rgba(6, 182, 212, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  margin-top: 2px;
}

.model-size {
  font-size: 12px;
  color: #64748b;
}

.model-card-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}

.model-detail {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-label {
  font-size: 12px;
  color: #64748b;
}

.detail-value {
  font-size: 12px;
  color: #94a3b8;
  font-weight: 500;
}

.model-card-footer {
  display: flex;
  gap: 8px;
  padding-top: 16px;
  border-top: 1px solid rgba(51, 65, 85, 0.5);
}

.model-card-footer .tech-btn {
  flex: 1;
  justify-content: center;
}

.tech-dialog :deep(.el-dialog) {
  background: rgba(15, 15, 25, 0.95);
  border: 1px solid rgba(6, 182, 212, 0.2);
  border-radius: 16px;
}

.tech-dialog :deep(.el-dialog__header) {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(51, 65, 85, 0.5);
}

.tech-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

.tech-dialog :deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid rgba(51, 65, 85, 0.5);
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dialog-icon {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
}

.dialog-icon.info {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
}

.dialog-title h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #e2e8f0;
}

.dialog-title span {
  font-size: 12px;
  color: #64748b;
}

.dialog-content {
  min-height: 100px;
}

.form-group {
  margin-bottom: 16px;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: #94a3b8;
  margin-bottom: 8px;
}

.tech-autocomplete {
  width: 100%;
}

.tech-autocomplete :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 8px;
  box-shadow: none;
}

.tech-autocomplete :deep(.el-input__inner) {
  color: #e2e8f0;
}

.form-help {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #64748b;
}

.pull-progress-section {
  padding: 8px 0;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.progress-title {
  font-size: 14px;
  font-weight: 500;
  color: #e2e8f0;
}

.progress-percent {
  font-size: 14px;
  font-weight: 600;
  color: #06b6d4;
}

.progress-bar-container {
  height: 8px;
  background: rgba(30, 41, 59, 0.8);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 16px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #06b6d4 0%, #22d3ee 100%);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-bar.success {
  background: linear-gradient(90deg, #10b981 0%, #34d399 100%);
}

.progress-bar.exception {
  background: linear-gradient(90deg, #ef4444 0%, #f87171 100%);
}

.progress-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-row.error .info-value {
  color: #ef4444;
}

.info-label {
  font-size: 12px;
  color: #64748b;
}

.info-value {
  font-size: 12px;
  color: #94a3b8;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: rgba(30, 41, 59, 0.4);
  border-radius: 8px;
}

.info-item.full-width {
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
}

.info-item .info-label {
  font-size: 13px;
  color: #64748b;
}

.info-item .info-value {
  font-size: 13px;
  color: #e2e8f0;
  font-weight: 500;
}

.info-item .info-value.digest {
  font-family: monospace;
  font-size: 11px;
  color: #94a3b8;
  word-break: break-all;
  background: rgba(15, 23, 42, 0.5);
  padding: 6px 10px;
  border-radius: 4px;
  width: 100%;
}

.is-loading {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
