<template>
  <div class="page-container">
    <div class="page-header">
      <h2>在线模型</h2>
      <div class="header-actions">
        <el-input
          v-model="searchQuery"
          placeholder="搜索在线模型..."
          prefix-icon="Search"
          style="width: 250px;"
          clearable
          @input="handleSearchDebounced"
          class="unified-input"
        />
        <el-select
          v-model="filterType"
          placeholder="模型类型"
          clearable
          style="width: 140px;"
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
          placeholder="参数规模"
          clearable
          style="width: 140px;"
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
          placeholder="排序方式"
          style="width: 150px;"
          @change="applyFilters"
        >
          <el-option label="默认排序" value="default" />
          <el-option label="名称 A-Z" value="name_asc" />
          <el-option label="名称 Z-A" value="name_desc" />
          <el-option label="大小升序" value="size_asc" />
          <el-option label="大小降序" value="size_desc" />
          <el-option label="参数量升序" value="params_asc" />
          <el-option label="参数量降序" value="params_desc" />
        </el-select>
      </div>
    </div>

    <!-- 筛选标签 -->
    <div class="filter-tags" v-if="hasActiveFilters">
      <span class="filter-label">当前筛选:</span>
      <el-tag v-if="searchQuery" closable @close="clearSearch" type="info" class="unified-tag">
        搜索: {{ searchQuery }}
      </el-tag>
      <el-tag v-if="filterType" closable @close="filterType = ''" type="primary" class="unified-tag">
        类型: {{ getTypeLabel(filterType) }}
      </el-tag>
      <el-tag v-if="filterSize" closable @close="filterSize = ''" type="success" class="unified-tag">
        规模: {{ getSizeLabel(filterSize) }}
      </el-tag>
      <el-button link type="primary" @click="clearAllFilters">
        清除全部
      </el-button>
    </div>

    <el-table
      :data="filteredAndSortedModels"
      v-loading="loading"
      class="unified-table"
      row-key="name"
    >
      <el-table-column prop="name" label="模型名称" width="200" sortable>
        <template #default="{ row }">
          <div class="model-name-cell">
            <el-icon><Box /></el-icon>
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column prop="description" label="描述" min-width="250" />
      
      <el-table-column prop="size" label="大小" width="100" sortable>
        <template #default="{ row }">
          <el-tag size="small" type="info" class="unified-tag">{{ row.size }}</el-tag>
        </template>
      </el-table-column>
      
      <el-table-column label="参数规模" width="120" sortable>
        <template #default="{ row }">
          <el-tag size="small" :type="getParamSizeType(row.details?.parameter_size)" class="unified-tag">
            {{ row.details?.parameter_size || '-' }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="getModelTypeTag(row)" class="unified-tag">
            {{ getModelTypeLabel(row) }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="pullModel(row)" class="unified-button">
            <el-icon><Download /></el-icon>
            拉取
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 加载更多按钮 -->
    <div v-if="hasMore" class="load-more-container">
      <el-button 
        type="primary" 
        @click="loadMoreModels" 
        :loading="loading"
        class="load-more-button unified-button"
      >
        <el-icon><Refresh /></el-icon>
        获取更多
      </el-button>
    </div>
    
    <div v-else-if="total > 0" class="no-more-container">
      <el-empty description="没有更多模型了" />
    </div>

    <!-- 模型详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="pullLoading ? '拉取模型 - ' + selectedModel?.name : '模型详情'"
      width="600px"
      :close-on-click-modal="!pullLoading"
      :close-on-press-escape="!pullLoading"
      class="unified-dialog"
    >
      <template v-if="!pullLoading && selectedModel">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="模型名称">
            {{ selectedModel.name }}
          </el-descriptions-item>
          <el-descriptions-item label="描述">
            {{ selectedModel.description }}
          </el-descriptions-item>
          <el-descriptions-item label="大小">
            {{ selectedModel.size }}
          </el-descriptions-item>
          <el-descriptions-item label="参数规模">
            {{ selectedModel.details?.parameter_size || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="量化级别">
            {{ selectedModel.details?.quantization_level || '-' }}
          </el-descriptions-item>
        </el-descriptions>
      </template>
      <template v-else-if="pullLoading">
        <div class="pull-progress-container">
          <el-progress
            :percentage="pullProgress"
            :status="pullStatus"
            :format="pullProgressFormat"
            :stroke-width="20"
            style="margin-bottom: 20px"
          />
          <el-descriptions :column="1" border>
            <el-descriptions-item label="拉取状态">
              {{ pullStatusText }}
            </el-descriptions-item>
            <el-descriptions-item label="当前任务">
              {{ pullCurrentTask }}
            </el-descriptions-item>
            <el-descriptions-item label="错误信息" v-if="pullError">
              <el-tag type="danger" effect="dark" class="unified-tag">{{ pullError }}</el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </template>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="cancelPull" :disabled="!pullLoading" class="unified-button">
            取消拉取
          </el-button>
          <el-button v-if="!pullLoading" @click="detailDialogVisible = false" class="unified-button">关闭</el-button>
          <el-button type="primary" v-if="!pullLoading" @click="confirmPullModel(selectedModel.name)" class="unified-button">
            确认拉取
          </el-button>
          <el-button type="success" v-else-if="pullStatus === 'success'" @click="detailDialogVisible = false" class="unified-button">
            完成
          </el-button>
          <el-button type="danger" v-else-if="pullStatus === 'exception'" @click="detailDialogVisible = false" class="unified-button">
            关闭
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { Search, Box, Download, Refresh } from '@element-plus/icons-vue'
import { GetOnlineModels, PullModel, SearchOnlineModels, ListModels, CancelPull } from '../../wailsjs/go/main/App'
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
// 分页相关
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const hasMore = ref(true)
// 拉取进度相关
const pullLoading = ref(false)
const pullProgress = ref(0)
const pullStatus = ref('')
const pullStatusText = ref('')
const pullCurrentTask = ref('')
const pullError = ref('')
const currentPullModel = ref('')

// 搜索防抖定时器
let searchDebounceTimer = null

/**
 * 是否有激活的筛选条件
 */
const hasActiveFilters = computed(() => {
  return searchQuery.value || filterType.value || filterSize.value
})

/**
 * 解析参数大小（返回数值，单位为 B）
 */
const parseParamSize = (sizeStr) => {
  if (!sizeStr) return 0
  const match = sizeStr.match(/(\d+(?:\.\d+)?)/i)
  if (!match) return 0
  const num = parseFloat(match[1])
  if (sizeStr.toLowerCase().includes('m')) return num / 1000
  return num
}

/**
 * 解析文件大小（返回字节数）
 */
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

/**
 * 获取参数规模分类
 */
const getParamSizeCategory = (sizeStr) => {
  const size = parseParamSize(sizeStr)
  if (size < 1) return 'small'
  if (size < 10) return 'medium'
  if (size < 70) return 'large'
  return 'xlarge'
}

/**
 * 获取参数规模标签类型
 */
const getParamSizeType = (sizeStr) => {
  const category = getParamSizeCategory(sizeStr)
  switch (category) {
    case 'small': return 'success'
    case 'medium': return 'primary'
    case 'large': return 'warning'
    case 'xlarge': return 'danger'
    default: return 'info'
  }
}

/**
 * 获取模型类型
 */
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

/**
 * 获取模型类型标签
 */
const getModelTypeTag = (model) => {
  const type = getModelType(model)
  switch (type) {
    case 'code': return 'warning'
    case 'multimodal': return 'danger'
    case 'embedding': return 'success'
    default: return 'primary'
  }
}

/**
 * 获取模型类型标签文本
 */
const getModelTypeLabel = (model) => {
  const type = getModelType(model)
  switch (type) {
    case 'code': return '代码'
    case 'multimodal': return '多模态'
    case 'embedding': return '嵌入'
    default: return '文本'
  }
}

/**
 * 获取类型标签文本
 */
const getTypeLabel = (type) => {
  switch (type) {
    case 'text': return '文本生成'
    case 'code': return '代码生成'
    case 'multimodal': return '多模态'
    case 'embedding': return '嵌入模型'
    default: return type
  }
}

/**
 * 获取规模标签文本
 */
const getSizeLabel = (size) => {
  switch (size) {
    case 'small': return '小型 (<1B)'
    case 'medium': return '中型 (1B-10B)'
    case 'large': return '大型 (10B-70B)'
    case 'xlarge': return '超大 (>70B)'
    default: return size
  }
}

/**
 * 筛选和排序后的模型列表
 */
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
    case 'params_asc':
      result.sort((a, b) => parseParamSize(a.details?.parameter_size) - parseParamSize(b.details?.parameter_size))
      break
    case 'params_desc':
      result.sort((a, b) => parseParamSize(b.details?.parameter_size) - parseParamSize(a.details?.parameter_size))
      break
    default:
      break
  }
  
  return result
})

/**
 * 防抖搜索处理
 */
const handleSearchDebounced = () => {
  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer)
  }
  searchDebounceTimer = setTimeout(() => {
    performSearch()
  }, 300)
}

/**
 * 执行搜索
 */
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

/**
 * 应用筛选
 */
const applyFilters = () => {
  // 筛选是在前端进行的，不需要重新请求数据
}

/**
 * 清除搜索
 */
const clearSearch = () => {
  searchQuery.value = ''
  loadOnlineModels()
}

/**
 * 清除所有筛选
 */
const clearAllFilters = () => {
  searchQuery.value = ''
  filterType.value = ''
  filterSize.value = ''
  sortBy.value = 'default'
  loadOnlineModels()
}

/**
 * 加载在线模型列表
 */
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

/**
 * 加载更多模型
 */
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

/**
 * 重新加载本地模型列表
 */
const reloadLocalModels = async () => {
  try {
    window.dispatchEvent(new CustomEvent('localModelsUpdated'))
    console.log('本地模型列表已更新，通知ModelsView组件重新加载')
  } catch (error) {
    console.error('重新加载本地模型列表失败:', error)
  }
}

/**
 * 进度条格式化
 */
const pullProgressFormat = (percentage) => {
  return `${percentage}%`
}

/**
 * 取消拉取
 */
const cancelPull = async () => {
  if (!currentPullModel.value) {
    detailDialogVisible.value = false
    resetPullState()
    return
  }

  try {
    const result = await CancelPull(currentPullModel.value)
    if (result.success) {
      ElMessage.success('已取消模型拉取')
      detailDialogVisible.value = false
      resetPullState()
    } else {
      ElMessage.warning(result.message || '取消失败')
    }
  } catch (error) {
    console.error('取消拉取失败:', error)
    ElMessage.error('取消拉取失败: ' + error.message)
  }
}

/**
 * 重置拉取状态
 */
const resetPullState = () => {
  pullLoading.value = false
  pullProgress.value = 0
  pullStatus.value = ''
  pullStatusText.value = ''
  pullCurrentTask.value = ''
  pullError.value = ''
  currentPullModel.value = ''
}

/**
 * 监听模型拉取进度事件
 */
const setupPullProgressListener = () => {
  EventsOn('model_pull_progress', (eventData) => {
    const { model, status, progress, message, time } = eventData
    
    if (model !== currentPullModel.value) {
      return
    }
    
    if (progress >= 0) {
      pullProgress.value = Math.round(progress)
    }
    
    switch (status) {
      case 'started':
        pullStatus.value = ''
        pullStatusText.value = '开始拉取'
        pullCurrentTask.value = message
        break
      case 'downloading':
        pullStatus.value = ''
        pullStatusText.value = '下载中'
        pullCurrentTask.value = message
        break
      case 'completed':
        pullStatus.value = 'success'
        pullStatusText.value = '拉取完成'
        pullCurrentTask.value = message
        pullProgress.value = 100
        setTimeout(() => {
          reloadLocalModels()
        }, 2000)
        break
      case 'cancelled':
        pullStatus.value = 'warning'
        pullStatusText.value = '已取消'
        pullCurrentTask.value = message
        pullError.value = message
        break
      case 'error':
        pullStatus.value = 'exception'
        pullStatusText.value = '拉取失败'
        pullCurrentTask.value = message
        pullError.value = message
        break
    }
  })
}

/**
 * 拉取模型
 */
const pullModel = (model) => {
  selectedModel.value = model
  detailDialogVisible.value = true
}

/**
 * 确认拉取模型
 */
const confirmPullModel = async (modelName) => {
  currentPullModel.value = modelName
  
  pullLoading.value = true
  pullStatusText.value = '开始拉取'
  pullCurrentTask.value = `准备拉取模型: ${modelName}`
  
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

// 初始化
onMounted(() => {
  loadOnlineModels()
  setupPullProgressListener()
})
</script>

<style scoped>
.filter-tags {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md) var(--spacing-xl);
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-color);
  flex-wrap: wrap;
}

.filter-label {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.model-name-cell {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-md);
}

.load-more-container {
  display: flex;
  justify-content: center;
  padding: var(--spacing-xl);
  border-top: 1px solid var(--border-color);
}

.load-more-button {
  min-width: 150px;
}

.no-more-container {
  display: flex;
  justify-content: center;
  padding: var(--spacing-3xl) var(--spacing-xl);
  border-top: 1px solid var(--border-color);
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .header-actions {
    width: 100%;
    flex-direction: column;
  }
  
  .header-actions .el-input,
  .header-actions .el-select {
    width: 100% !important;
  }
}
</style>
