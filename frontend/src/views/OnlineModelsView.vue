<template>
  <div class="online-models-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-title">
        <div class="title-icon">
          <el-icon size="24"><Download /></el-icon>
        </div>
        <div class="title-content">
          <h2>在线模型库</h2>
          <p class="subtitle">探索和下载Ollama生态中的AI模型</p>
        </div>
      </div>
      <div class="header-stats">
        <div class="stat-item">
          <span class="stat-value">{{ total }}</span>
          <span class="stat-label">可用模型</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ onlineModels.length }}</span>
          <span class="stat-label">已加载</span>
        </div>
      </div>
    </div>

    <!-- 搜索和筛选栏 -->
    <div class="search-bar">
      <div class="search-input-wrapper">
        <el-icon class="search-icon"><Search /></el-icon>
        <el-input
          v-model="searchQuery"
          placeholder="搜索模型名称或描述..."
          clearable
          @input="handleSearchDebounced"
          class="tech-input"
        />
      </div>
      <div class="filter-group">
        <el-select
          v-model="filterType"
          placeholder="类型"
          clearable
          class="tech-select"
          @change="applyFilters"
        >
          <el-option label="全部类型" value="" />
          <el-option label="文本生成" value="text" />
          <el-option label="代码生成" value="code" />
          <el-option label="多模态" value="multimodal" />
          <el-option label="嵌入模型" value="embedding" />
        </el-select>
        <el-select
          v-model="filterSize"
          placeholder="规模"
          clearable
          class="tech-select"
          @change="applyFilters"
        >
          <el-option label="全部规模" value="" />
          <el-option label="小型 (<1B)" value="small" />
          <el-option label="中型 (1B-10B)" value="medium" />
          <el-option label="大型 (10B-70B)" value="large" />
          <el-option label="超大 (>70B)" value="xlarge" />
        </el-select>
        <el-select
          v-model="sortBy"
          placeholder="排序"
          class="tech-select"
          @change="applyFilters"
        >
          <el-option label="默认排序" value="default" />
          <el-option label="名称 A-Z" value="name_asc" />
          <el-option label="名称 Z-A" value="name_desc" />
          <el-option label="大小升序" value="size_asc" />
          <el-option label="大小降序" value="size_desc" />
        </el-select>
      </div>
    </div>

    <!-- 筛选标签 -->
    <div class="filter-tags" v-if="hasActiveFilters">
      <span class="filter-label">当前筛选:</span>
      <el-tag v-if="searchQuery" closable @close="clearSearch" type="info">
        搜索: {{ searchQuery }}
      </el-tag>
      <el-tag v-if="filterType" closable @close="filterType = ''" type="primary">
        类型: {{ getTypeLabel(filterType) }}
      </el-tag>
      <el-tag v-if="filterSize" closable @close="filterSize = ''" type="success">
        规模: {{ getSizeLabel(filterSize) }}
      </el-tag>
      <el-button link type="primary" @click="clearAllFilters">清除全部</el-button>
    </div>

    <!-- 模型网格 -->
    <div class="models-grid" v-loading="loading" element-loading-text="加载模型中...">
      <div 
        v-for="model in filteredAndSortedModels" 
        :key="model.name" 
        class="model-card"
        @click="showModelDetail(model)"
      >
        <div class="card-header">
          <div class="model-icon">
            <el-icon size="20"><Box /></el-icon>
          </div>
          <div class="model-title">
            <h4>{{ model.name }}</h4>
            <div class="model-tags">
              <el-tag size="small" :type="getModelTypeTag(model)">
                {{ getModelTypeLabel(model) }}
              </el-tag>
              <el-tag size="small" type="info">{{ model.size }}</el-tag>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="model-desc">{{ model.description || '暂无描述' }}</p>
          <div class="model-meta">
            <span v-if="model.details?.parameter_size">
              <el-icon><Cpu /></el-icon>
              {{ model.details.parameter_size }}
            </span>
            <span v-if="model.details?.quantization_level">
              <el-icon><Setting /></el-icon>
              {{ model.details.quantization_level }}
            </span>
          </div>
        </div>
        <div class="card-footer">
          <el-button 
            type="primary" 
            size="small" 
            @click.stop="pullModel(model)"
            :disabled="pullLoading && currentPullModel === model.name"
            class="pull-btn"
          >
            <el-icon><Download /></el-icon>
            {{ pullLoading && currentPullModel === model.name ? '拉取中...' : '拉取模型' }}
          </el-button>
          <el-button 
            size="small" 
            @click.stop="pullInBackground(model)"
            :disabled="pullLoading && currentPullModel === model.name"
            class="pull-btn-secondary"
          >
            <el-icon><Download /></el-icon>
            后台拉取
          </el-button>
        </div>
      </div>
    </div>

    <!-- 加载更多 -->
    <div v-if="hasMore" class="load-more">
      <el-button @click="loadMoreModels" :loading="loading" class="load-more-btn">
        <el-icon><Refresh /></el-icon>
        加载更多模型
      </el-button>
    </div>
    <div v-else-if="total > 0 && !loading" class="no-more">
      <el-icon size="40"><CircleCheck /></el-icon>
      <span>已加载全部 {{ total }} 个模型</span>
    </div>

    <!-- 拉取进度对话框 -->
    <el-dialog
      v-model="pullDialogVisible"
      :title="null"
      width="500px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
      class="pull-dialog"
    >
      <div class="pull-content">
        <!-- 成功状态 -->
        <div v-if="pullStatus === 'success'" class="pull-success">
          <div class="success-icon">
            <el-icon size="60"><SuccessFilled /></el-icon>
          </div>
          <h3>模型拉取完成!</h3>
          <p class="model-name">{{ currentPullModel }}</p>
          <p class="success-message">模型已成功下载并可以使用</p>
        </div>
        
        <!-- 错误状态 -->
        <div v-else-if="pullStatus === 'exception'" class="pull-error">
          <div class="error-icon">
            <el-icon size="60"><WarningFilled /></el-icon>
          </div>
          <h3>拉取失败</h3>
          <p class="model-name">{{ currentPullModel }}</p>
          <p class="error-message">{{ pullError || '拉取过程中发生错误' }}</p>
        </div>
        
        <!-- 进度状态 -->
        <div v-else class="pull-progress">
          <div class="progress-header">
            <div class="model-info">
              <el-icon size="24" class="pulse"><Download /></el-icon>
              <div>
                <h3>正在拉取模型</h3>
                <p class="model-name">{{ currentPullModel }}</p>
              </div>
            </div>
          </div>
          
          <div class="progress-bar-container">
            <div class="progress-bar">
              <div 
                class="progress-fill" 
                :style="{ width: pullProgress + '%' }"
              ></div>
            </div>
            <div class="progress-info">
              <span class="progress-percent">{{ pullProgress }}%</span>
              <span class="progress-status">{{ pullStatusText || '准备中...' }}</span>
            </div>
          </div>
          
          <div class="progress-details">
            <div class="detail-item">
              <el-icon><Loading /></el-icon>
              <span>{{ pullCurrentTask || '初始化...' }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <!-- 成功状态按钮 -->
          <template v-if="pullStatus === 'success'">
            <el-button type="primary" @click="closePullDialog" class="tech-btn primary">
              <el-icon><Check /></el-icon>
              完成
            </el-button>
          </template>
          
          <!-- 错误状态按钮 -->
          <template v-else-if="pullStatus === 'exception'">
            <el-button @click="closePullDialog" class="tech-btn">关闭</el-button>
            <el-button type="primary" @click="retryPull" class="tech-btn primary">
              <el-icon><RefreshRight /></el-icon>
              重试
            </el-button>
          </template>
          
          <!-- 进度状态按钮 -->
          <template v-else>
            <el-button @click="cancelPull" class="tech-btn danger">
              <el-icon><Close /></el-icon>
              取消拉取
            </el-button>
          </template>
        </div>
      </template>
    </el-dialog>

    <!-- 模型详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="null"
      width="500px"
      class="detail-dialog"
    >
      <div v-if="selectedModel" class="detail-content">
        <div class="detail-header">
          <div class="model-icon large">
            <el-icon size="32"><Box /></el-icon>
          </div>
          <h3>{{ selectedModel.name }}</h3>
          <div class="detail-tags">
            <el-tag :type="getModelTypeTag(selectedModel)">
              {{ getModelTypeLabel(selectedModel) }}
            </el-tag>
            <el-tag type="info">{{ selectedModel.size }}</el-tag>
          </div>
        </div>
        
        <div class="detail-body">
          <p class="detail-desc">{{ selectedModel.description || '暂无描述' }}</p>
          
          <div class="detail-meta">
            <div class="meta-item" v-if="selectedModel.details?.parameter_size">
              <el-icon><Cpu /></el-icon>
              <div>
                <span class="meta-label">参数规模</span>
                <span class="meta-value">{{ selectedModel.details.parameter_size }}</span>
              </div>
            </div>
            <div class="meta-item" v-if="selectedModel.details?.quantization_level">
              <el-icon><Setting /></el-icon>
              <div>
                <span class="meta-label">量化级别</span>
                <span class="meta-value">{{ selectedModel.details.quantization_level }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <el-button @click="detailDialogVisible = false" class="tech-btn">取消</el-button>
        <el-button type="primary" @click="confirmPullModel(selectedModel.name)" class="tech-btn primary">
          <el-icon><Download /></el-icon>
          拉取模型
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  Download, Search, Box, Refresh, Cpu, Setting,
  CircleCheck, SuccessFilled, WarningFilled, Loading,
  Check, Close, RefreshRight
} from '@element-plus/icons-vue'
import { GetOnlineModels, PullModel, SearchOnlineModels, CancelPull } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { ElMessage } from 'element-plus'

// 响应式数据
const onlineModels = ref([])
const loading = ref(false)
const searchQuery = ref('')
const filterType = ref('')
const filterSize = ref('')
const sortBy = ref('default')
const detailDialogVisible = ref(false)
const selectedModel = ref(null)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const hasMore = ref(true)

// 拉取进度相关
const pullDialogVisible = ref(false)
const pullLoading = ref(false)
const pullProgress = ref(0)
const pullStatus = ref('')
const pullStatusText = ref('')
const pullCurrentTask = ref('')
const pullError = ref('')
const currentPullModel = ref('')

let searchDebounceTimer = null

const hasActiveFilters = computed(() => {
  return searchQuery.value || filterType.value || filterSize.value
})

const parseParamSize = (sizeStr) => {
  if (!sizeStr) return 0
  const match = sizeStr.match(/(\d+(?:\.\d+)?)/i)
  if (!match) return 0
  const num = parseFloat(match[1])
  if (sizeStr.toLowerCase().includes('m')) return num / 1000
  return num
}

const parseFileSize = (sizeStr) => {
  if (!sizeStr) return 0
  const match = sizeStr.match(/(\d+(?:\.\d+)?)/i)
  if (!match) return 0
  const num = parseFloat(match[1])
  if (sizeStr.toLowerCase().includes('gb')) return num * 1024 * 1024 * 1024
  if (sizeStr.toLowerCase().includes('mb')) return num * 1024 * 1024
  if (sizeStr.toLowerCase().includes('kb')) return num * 1024
  return num
}

const getParamSizeCategory = (sizeStr) => {
  const size = parseParamSize(sizeStr)
  if (size < 1) return 'small'
  if (size < 10) return 'medium'
  if (size < 70) return 'large'
  return 'xlarge'
}

const getModelType = (model) => {
  const name = model.name?.toLowerCase() || ''
  const desc = model.description?.toLowerCase() || ''
  
  if (name.includes('code') || name.includes('codellama') || name.includes('starcoder') || 
      desc.includes('code') || desc.includes('programming')) {
    return 'code'
  }
  if (name.includes('clip') || name.includes('llava') || name.includes('vision') ||
      desc.includes('multimodal') || desc.includes('vision')) {
    return 'multimodal'
  }
  if (name.includes('embed') || name.includes('nomic-embed') || name.includes('all-minilm') ||
      desc.includes('embedding') || desc.includes('vector')) {
    return 'embedding'
  }
  return 'text'
}

const getModelTypeTag = (model) => {
  const type = getModelType(model)
  switch (type) {
    case 'code': return 'warning'
    case 'multimodal': return 'danger'
    case 'embedding': return 'success'
    default: return 'primary'
  }
}

const getModelTypeLabel = (model) => {
  const type = getModelType(model)
  switch (type) {
    case 'code': return '代码'
    case 'multimodal': return '多模态'
    case 'embedding': return '嵌入'
    default: return '文本'
  }
}

const getTypeLabel = (type) => {
  switch (type) {
    case 'text': return '文本生成'
    case 'code': return '代码生成'
    case 'multimodal': return '多模态'
    case 'embedding': return '嵌入模型'
    default: return type
  }
}

const getSizeLabel = (size) => {
  switch (size) {
    case 'small': return '小型 (<1B)'
    case 'medium': return '中型 (1B-10B)'
    case 'large': return '大型 (10B-70B)'
    case 'xlarge': return '超大 (>70B)'
    default: return size
  }
}

const filteredAndSortedModels = computed(() => {
  let result = [...onlineModels.value]
  
  if (filterType.value) {
    result = result.filter(model => getModelType(model) === filterType.value)
  }
  
  if (filterSize.value) {
    result = result.filter(model => {
      const category = getParamSizeCategory(model.details?.parameter_size)
      return category === filterSize.value
    })
  }
  
  switch (sortBy.value) {
    case 'name_asc':
      result.sort((a, b) => (a.name || '').localeCompare(b.name || ''))
      break
    case 'name_desc':
      result.sort((a, b) => (b.name || '').localeCompare(a.name || ''))
      break
    case 'size_asc':
      result.sort((a, b) => parseFileSize(a.size) - parseFileSize(b.size))
      break
    case 'size_desc':
      result.sort((a, b) => parseFileSize(b.size) - parseFileSize(a.size))
      break
  }
  
  return result
})

const handleSearchDebounced = () => {
  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer)
  }
  searchDebounceTimer = setTimeout(() => {
    performSearch()
  }, 300)
}

const performSearch = async () => {
  currentPage.value = 1
  onlineModels.value = []
  hasMore.value = true
  
  loading.value = true
  try {
    const results = await SearchOnlineModels(searchQuery.value, currentPage.value, pageSize.value)
    onlineModels.value = results.models
    total.value = results.total
    hasMore.value = onlineModels.value.length < results.total
  } catch (error) {
    console.error('搜索在线模型失败:', error)
    ElMessage.error('搜索在线模型失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {}

const clearSearch = () => {
  searchQuery.value = ''
  loadOnlineModels()
}

const clearAllFilters = () => {
  searchQuery.value = ''
  filterType.value = ''
  filterSize.value = ''
  sortBy.value = 'default'
  loadOnlineModels()
}

const loadOnlineModels = async () => {
  currentPage.value = 1
  onlineModels.value = []
  hasMore.value = true
  
  loading.value = true
  try {
    const results = await GetOnlineModels(currentPage.value, pageSize.value)
    onlineModels.value = results.models
    total.value = results.total
    hasMore.value = onlineModels.value.length < results.total
  } catch (error) {
    console.error('加载在线模型失败:', error)
    ElMessage.error('加载在线模型失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const loadMoreModels = async () => {
  if (!hasMore.value || loading.value) return
  
  currentPage.value++
  loading.value = true
  
  try {
    let results
    if (searchQuery.value) {
      results = await SearchOnlineModels(searchQuery.value, currentPage.value, pageSize.value)
    } else {
      results = await GetOnlineModels(currentPage.value, pageSize.value)
    }
    
    onlineModels.value = [...onlineModels.value, ...results.models]
    total.value = results.total
    hasMore.value = onlineModels.value.length < results.total
  } catch (error) {
    console.error('加载更多模型失败:', error)
    ElMessage.error('加载更多模型失败: ' + error.message)
    currentPage.value--
  } finally {
    loading.value = false
  }
}

const showModelDetail = (model) => {
  selectedModel.value = model
  detailDialogVisible.value = true
}

const pullModel = (model) => {
  selectedModel.value = model
  confirmPullModel(model.name)
}

const confirmPullModel = async (modelName) => {
  detailDialogVisible.value = false
  currentPullModel.value = modelName
  
  // 重置状态
  pullProgress.value = 0
  pullStatus.value = ''
  pullStatusText.value = '初始化...'
  pullCurrentTask.value = '准备拉取模型'
  pullError.value = ''
  pullLoading.value = true
  pullDialogVisible.value = true
  
  try {
    await PullModel(modelName)
  } catch (error) {
    console.error('拉取模型失败:', error)
    pullStatus.value = 'exception'
    pullStatusText.value = '拉取失败'
    pullError.value = error.message || '拉取模型时发生错误'
    ElMessage.error('拉取模型失败: ' + error.message)
  }
}

const pullInBackground = (model) => {
  currentPullModel.value = model.name
  pullProgress.value = 0
  pullStatus.value = ''
  pullStatusText.value = '后台拉取中'
  pullCurrentTask.value = '准备后台拉取模型'
  pullError.value = ''
  pullLoading.value = true
  
  PullModel(model.name).then(() => {
    ElMessage.success(`模型 ${model.name} 已开始后台拉取`)
  }).catch(error => {
    console.error('后台拉取模型失败:', error)
    ElMessage.error('后台拉取模型失败: ' + error.message)
    pullLoading.value = false
  })
}

const cancelPull = async () => {
  if (!currentPullModel.value) {
    closePullDialog()
    return
  }

  try {
    const result = await CancelPull(currentPullModel.value)
    if (result.success) {
      ElMessage.success('已取消模型拉取')
      closePullDialog()
    } else {
      ElMessage.warning(result.message || '取消失败')
    }
  } catch (error) {
    console.error('取消拉取失败:', error)
    ElMessage.error('取消拉取失败: ' + error.message)
  }
}

const retryPull = () => {
  if (currentPullModel.value) {
    confirmPullModel(currentPullModel.value)
  }
}

const closePullDialog = () => {
  pullDialogVisible.value = false
  pullLoading.value = false
  pullProgress.value = 0
  pullStatus.value = ''
  pullStatusText.value = ''
  pullCurrentTask.value = ''
  pullError.value = ''
  
  // 通知本地模型列表更新
  if (pullStatus.value === 'success') {
    window.dispatchEvent(new CustomEvent('localModelsUpdated'))
  }
}

const setupPullProgressListener = () => {
  EventsOn('model_pull_progress', (eventData) => {
    console.log('收到进度事件:', eventData)
    const { model, status, progress, message } = eventData
    
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
        pullStatusText.value = '完成'
        pullCurrentTask.value = message || '模型拉取完成'
        pullProgress.value = 100
        // 通知本地模型列表更新
        window.dispatchEvent(new CustomEvent('localModelsUpdated'))
        break
      case 'cancelled':
        pullStatus.value = 'warning'
        pullStatusText.value = '已取消'
        pullCurrentTask.value = message || '用户取消了拉取'
        pullError.value = message
        break
      case 'error':
        pullStatus.value = 'exception'
        pullStatusText.value = '失败'
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

onMounted(() => {
  loadOnlineModels()
  setupPullProgressListener()
})
</script>

<style scoped>
.online-models-page {
  padding: 24px;
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0f1a 0%, #1a1f2e 100%);
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 16px;
}

.title-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #06b6d4 0%, #3b82f6 100%);
  border-radius: 12px;
  color: white;
}

.title-content h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #fff;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
}

.header-stats {
  display: flex;
  gap: 32px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  display: block;
  font-size: 28px;
  font-weight: 700;
  color: #06b6d4;
}

.stat-label {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
}

/* 搜索栏 */
.search-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}

.search-input-wrapper {
  flex: 1;
  position: relative;
}

.search-icon {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(255, 255, 255, 0.4);
  z-index: 1;
}

.tech-input {
  width: 100%;
}

.tech-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding-left: 44px;
  box-shadow: none;
}

.tech-input :deep(.el-input__wrapper:hover) {
  border-color: rgba(6, 182, 212, 0.5);
}

.tech-input :deep(.el-input__wrapper.is-focus) {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.tech-input :deep(.el-input__inner) {
  color: #fff;
}

.tech-input :deep(.el-input__inner::placeholder) {
  color: rgba(255, 255, 255, 0.4);
}

.filter-group {
  display: flex;
  gap: 12px;
}

.tech-select {
  width: 120px;
}

.tech-select :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  box-shadow: none;
}

.tech-select :deep(.el-input__wrapper:hover) {
  border-color: rgba(6, 182, 212, 0.5);
}

.tech-select :deep(.el-input__inner) {
  color: #fff;
}

/* 筛选标签 */
.filter-tags {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 12px;
}

.filter-label {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
}

/* 模型网格 */
.models-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  min-height: 200px;
}

.model-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.model-card:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(6, 182, 212, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 8px 32px rgba(6, 182, 212, 0.1);
}

.card-header {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
}

.model-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.2) 0%, rgba(59, 130, 246, 0.2) 100%);
  border-radius: 10px;
  color: #06b6d4;
  flex-shrink: 0;
}

.model-title h4 {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
}

.model-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.model-tags .el-tag {
  background: rgba(255, 255, 255, 0.1);
  border: none;
  color: rgba(255, 255, 255, 0.8);
}

.card-body {
  margin-bottom: 16px;
}

.model-desc {
  margin: 0 0 12px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.model-meta {
  display: flex;
  gap: 16px;
}

.model-meta span {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.4);
}

.card-footer {
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.pull-btn {
  width: 100%;
  background: linear-gradient(135deg, #06b6d4 0%, #3b82f6 100%);
  border: none;
  border-radius: 10px;
  font-weight: 500;
}

.pull-btn:hover {
  opacity: 0.9;
}

.pull-btn-secondary {
  width: 100%;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 10px;
  color: #06b6d4;
  font-weight: 500;
  margin-top: 8px;
}

.pull-btn-secondary:hover {
  background: rgba(6, 182, 212, 0.1);
  border-color: rgba(6, 182, 212, 0.5);
}

/* 加载更多 */
.load-more {
  display: flex;
  justify-content: center;
  padding: 32px;
}

.load-more-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #fff;
  border-radius: 12px;
  padding: 12px 32px;
}

.load-more-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(6, 182, 212, 0.5);
}

.no-more {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px;
  color: rgba(255, 255, 255, 0.4);
}

/* 拉取对话框 */
.pull-dialog :deep(.el-dialog) {
  background: #1a1f2e;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
}

.pull-dialog :deep(.el-dialog__header) {
  display: none;
}

.pull-dialog :deep(.el-dialog__body) {
  padding: 32px;
}

.pull-content {
  text-align: center;
}

/* 成功状态 */
.pull-success {
  padding: 20px 0;
}

.success-icon {
  color: #10b981;
  margin-bottom: 16px;
}

.pull-success h3 {
  margin: 0 0 8px;
  font-size: 20px;
  color: #fff;
}

.model-name {
  margin: 0 0 8px;
  font-size: 14px;
  color: #06b6d4;
  font-family: monospace;
}

.success-message {
  margin: 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.5);
}

/* 错误状态 */
.pull-error {
  padding: 20px 0;
}

.error-icon {
  color: #ef4444;
  margin-bottom: 16px;
}

.pull-error h3 {
  margin: 0 0 8px;
  font-size: 20px;
  color: #fff;
}

.error-message {
  margin: 0;
  font-size: 13px;
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
  padding: 12px 16px;
  border-radius: 10px;
  margin-top: 12px;
}

/* 进度状态 */
.pull-progress {
  padding: 20px 0;
}

.progress-header {
  margin-bottom: 24px;
}

.model-info {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.model-info h3 {
  margin: 0;
  font-size: 18px;
  color: #fff;
}

.pulse {
  color: #06b6d4;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.progress-bar-container {
  margin-bottom: 20px;
}

.progress-bar {
  height: 8px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 12px;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #06b6d4 0%, #3b82f6 100%);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.progress-percent {
  color: #06b6d4;
  font-weight: 600;
}

.progress-status {
  color: rgba(255, 255, 255, 0.5);
}

.progress-details {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 10px;
  padding: 12px 16px;
}

.detail-item {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.6);
}

/* 对话框按钮 */
.dialog-footer {
  display: flex;
  justify-content: center;
  gap: 12px;
  padding-top: 16px;
}

.tech-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #fff;
  border-radius: 10px;
  padding: 10px 24px;
}

.tech-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.tech-btn.primary {
  background: linear-gradient(135deg, #06b6d4 0%, #3b82f6 100%);
  border: none;
}

.tech-btn.danger {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
  color: #ef4444;
}

/* 详情对话框 */
.detail-dialog :deep(.el-dialog) {
  background: #1a1f2e;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
}

.detail-dialog :deep(.el-dialog__header) {
  display: none;
}

.detail-dialog :deep(.el-dialog__body) {
  padding: 32px;
}

.detail-content {
  text-align: center;
}

.detail-header {
  margin-bottom: 20px;
}

.model-icon.large {
  width: 64px;
  height: 64px;
  margin: 0 auto 16px;
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.2) 0%, rgba(59, 130, 246, 0.2) 100%);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #06b6d4;
}

.detail-header h3 {
  margin: 0 0 12px;
  font-size: 22px;
  color: #fff;
}

.detail-tags {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.detail-tags .el-tag {
  background: rgba(255, 255, 255, 0.1);
  border: none;
  color: rgba(255, 255, 255, 0.8);
}

.detail-desc {
  margin: 0 0 20px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1.6;
}

.detail-meta {
  display: flex;
  justify-content: center;
  gap: 24px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(255, 255, 255, 0.03);
  padding: 12px 20px;
  border-radius: 12px;
}

.meta-item .el-icon {
  color: #06b6d4;
  font-size: 20px;
}

.meta-label {
  display: block;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.4);
}

.meta-value {
  display: block;
  font-size: 15px;
  color: #fff;
  font-weight: 500;
}

/* 响应式 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    text-align: center;
    gap: 20px;
  }
  
  .search-bar {
    flex-direction: column;
  }
  
  .filter-group {
    flex-wrap: wrap;
  }
  
  .models-grid {
    grid-template-columns: 1fr;
  }
}
</style>
