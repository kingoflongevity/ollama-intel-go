<template>
  <div class="online-models-container">
    <div class="models-header">
      <h2>在线模型</h2>
      <div class="header-actions">
        <el-input
          v-model="searchQuery"
          placeholder="搜索在线模型..."
          prefix-icon="Search"
          style="width: 300px;"
          clearable
        />
      </div>
    </div>

    <el-table
      :data="onlineModels"
      v-loading="loading"
      class="models-table"
      row-key="name"
    >
      <el-table-column prop="name" label="模型名称" width="200">
        <template #default="{ row }">
          <div class="model-name-cell">
            <el-icon><Box /></el-icon>
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column prop="description" label="描述" min-width="300" />
      
      <el-table-column prop="size" label="大小" width="120" />
      
      <el-table-column label="参数规模" width="120">
        <template #default="{ row }">
          {{ row.details?.parameter_size || '-' }}
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="pullModel(row)">
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
        class="load-more-button"
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
              <el-tag type="danger" effect="dark">{{ pullError }}</el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </template>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="cancelPull" :disabled="!pullLoading">
            取消拉取
          </el-button>
          <el-button v-if="!pullLoading" @click="detailDialogVisible = false">关闭</el-button>
          <el-button type="primary" v-if="!pullLoading" @click="confirmPullModel(selectedModel.name)">
            确认拉取
          </el-button>
          <el-button type="success" v-else-if="pullStatus === 'success'" @click="detailDialogVisible = false">
            完成
          </el-button>
          <el-button type="danger" v-else-if="pullStatus === 'exception'" @click="detailDialogVisible = false">
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
import { GetOnlineModels, PullModel, SearchOnlineModels, ListModels } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { ElMessage } from 'element-plus'

// 响应式数据
const onlineModels = ref([])
const loading = ref(false)
const searchQuery = ref('')
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

// 监听搜索关键词变化
watch(searchQuery, async (newQuery) => {
  // 重置页码
  currentPage.value = 1
  onlineModels.value = []
  hasMore.value = true
  
  loading.value = true
  try {
    const results = await SearchOnlineModels(newQuery, currentPage.value, pageSize.value)
    onlineModels.value = results.models
    total.value = results.total
    hasMore.value = onlineModels.value.length < results.total
  } catch (error) {
    console.error('搜索在线模型失败:', error)
    ElMessage.error('搜索在线模型失败')
  } finally {
    loading.value = false
  }
})

// 加载在线模型列表
const loadOnlineModels = async () => {
  // 重置页码
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

// 加载更多模型
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
    
    // 将新模型添加到现有列表
    onlineModels.value = [...onlineModels.value, ...results.models]
    total.value = results.total
    hasMore.value = onlineModels.value.length < results.total
  } catch (error) {
    console.error('加载更多模型失败:', error)
    ElMessage.error('加载更多模型失败: ' + error.message)
    // 加载失败时恢复页码
    currentPage.value--
  } finally {
    loading.value = false
  }
}

// 重新加载本地模型列表
const reloadLocalModels = async () => {
  try {
    // 触发一个全局事件通知ModelsView组件重新加载
    window.dispatchEvent(new CustomEvent('localModelsUpdated'))
    console.log('本地模型列表已更新，通知ModelsView组件重新加载')
  } catch (error) {
    console.error('重新加载本地模型列表失败:', error)
  }
}

// 进度条格式化
const pullProgressFormat = (percentage) => {
  return `${percentage}%`
}

// 取消拉取
const cancelPull = () => {
  // 这里可以添加取消拉取的逻辑
  // 目前后端没有提供取消拉取的API，所以只是关闭对话框
  detailDialogVisible.value = false
  resetPullState()
}

// 重置拉取状态
const resetPullState = () => {
  pullLoading.value = false
  pullProgress.value = 0
  pullStatus.value = ''
  pullStatusText.value = ''
  pullCurrentTask.value = ''
  pullError.value = ''
  currentPullModel.value = ''
}

// 监听模型拉取进度事件
const setupPullProgressListener = () => {
  EventsOn('model_pull_progress', (eventData) => {
    const { model, status, progress, message, time } = eventData
    
    // 只处理当前正在拉取的模型
    if (model !== currentPullModel.value) {
      return
    }
    
    // 更新进度
    if (progress >= 0) {
      pullProgress.value = Math.round(progress)
    }
    
    // 更新状态
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
        // 拉取完成后通知本地模型列表更新
        setTimeout(() => {
          reloadLocalModels()
        }, 1000) // 延迟1秒，确保模型已完全安装
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

// 拉取模型
const pullModel = (model) => {
  selectedModel.value = model
  detailDialogVisible.value = true
}

// 确认拉取模型
const confirmPullModel = async (modelName) => {
  currentPullModel.value = modelName
  
  pullLoading.value = true
  pullStatusText.value = '开始拉取'
  pullCurrentTask.value = `准备拉取模型: ${modelName}`
  
  try {
    await PullModel(modelName)
    // 拉取已经开始，不需要关闭对话框，等待进度更新
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
.online-models-container {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  height: 100%;
  display: flex;
  flex-direction: column;
}

body.dark-theme .online-models-container {
  background: #1e1e1e;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.models-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
}

body.dark-theme .models-header {
  background: #2d2d2d;
  border-bottom-color: #3c3c3c;
}

.models-header h2 {
  margin: 0;
  color: #303133;
}

body.dark-theme .models-header h2 {
  color: #e4e6eb;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.models-table {
  flex: 1;
}

.model-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.load-more-container {
  display: flex;
  justify-content: center;
  padding: 20px;
  border-top: 1px solid #e4e7ed;
}

body.dark-theme .load-more-container {
  border-top-color: #3c3c3c;
}

.load-more-button {
  min-width: 150px;
}

.no-more-container {
  display: flex;
  justify-content: center;
  padding: 40px 20px;
  border-top: 1px solid #e4e7ed;
}

body.dark-theme .no-more-container {
  border-top-color: #3c3c3c;
}

/* 暗色主题 - 表格样式 */
body.dark-theme .models-table {
  --el-table-bg-color: #1e1e1e;
  --el-table-text-color: #e4e6eb;
  --el-table-header-bg-color: #2d2d2d;
  --el-table-header-text-color: #e4e6eb;
  --el-table-border-color: #3c3c3c;
  --el-table-row-hover-bg-color: rgba(255, 255, 255, 0.05);
}

body.dark-theme .models-table :deep(.el-table) {
  background-color: #1e1e1e !important;
  color: #e4e6eb !important;
}

body.dark-theme .models-table :deep(.el-table__header-wrapper th) {
  background-color: #2d2d2d !important;
  color: #e4e6eb !important;
  border-bottom-color: #3c3c3c !important;
}

body.dark-theme .models-table :deep(.el-table__body-wrapper) {
  background-color: #1e1e1e !important;
}

body.dark-theme .models-table :deep(.el-table__row) {
  background-color: #1e1e1e !important;
  color: #e4e6eb !important;
  border-bottom-color: #3c3c3c !important;
}

body.dark-theme .models-table :deep(.el-table__row:hover) {
  background-color: rgba(255, 255, 255, 0.05) !important;
}

body.dark-theme .models-table :deep(.el-table__empty-block) {
  background-color: #1e1e1e !important;
  color: #a0aec0 !important;
}
</style>