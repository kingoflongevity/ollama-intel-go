<template>
  <div class="models-container">
    <div class="models-header">
      <h2>模型管理</h2>
      <div class="header-actions">
        <el-input
          v-model="searchQuery"
          placeholder="搜索模型..."
          prefix-icon="Search"
          style="width: 240px;"
          clearable
        />
        <el-button type="primary" @click="showPullDialog">
          <el-icon><Download /></el-icon>
          拉取模型
        </el-button>
        <el-button @click="refreshModels">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <el-table
      :data="filteredModels"
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
      
      <el-table-column prop="size" label="大小" width="120" />
      
      <el-table-column label="参数规模" width="120">
        <template #default="{ row }">
          {{ row.details?.parameter_size || '-' }}
        </template>
      </el-table-column>
      
      <el-table-column label="量化级别" width="150">
        <template #default="{ row }">
          {{ row.details?.quantization_level || '-' }}
        </template>
      </el-table-column>
      
      <el-table-column label="修改时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.modified_at) }}
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showModelInfo(row)">
            详情
          </el-button>
          <el-button size="small" type="danger" @click="deleteModel(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 拉取模型对话框 -->
    <el-dialog
      v-model="pullDialogVisible"
      title="拉取模型"
      width="500px"
      :close-on-click-modal="!pullLoading"
      :close-on-press-escape="!pullLoading"
    >
      <template v-if="!pullLoading">
        <el-form :model="pullForm" label-width="100px">
          <el-form-item label="模型名称">
            <el-autocomplete
              v-model="pullForm.name"
              :fetch-suggestions="queryModelSearch"
              placeholder="输入模型名称，如 llama3:8b"
              style="width: 100%;"
              clearable
            />
          </el-form-item>
        </el-form>
      </template>
      <template v-else>
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
          <el-button type="primary" v-if="!pullLoading" @click="confirmPullModel" :loading="pullLoading">
            确认拉取
          </el-button>
          <el-button type="success" v-else-if="pullStatus === 'success'" @click="pullDialogVisible = false">
            完成
          </el-button>
          <el-button type="danger" v-else-if="pullStatus === 'exception'" @click="pullDialogVisible = false">
            关闭
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 模型详情对话框 -->
    <el-dialog
      v-model="infoDialogVisible"
      title="模型详情"
      width="600px"
    >
      <el-descriptions v-if="selectedModel" :column="1" border>
        <el-descriptions-item label="模型名称">
          {{ selectedModel.name }}
        </el-descriptions-item>
        <el-descriptions-item label="模型大小">
          {{ selectedModel.size }}
        </el-descriptions-item>
        <el-descriptions-item label="参数规模">
          {{ selectedModel.details?.parameter_size || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="量化级别">
          {{ selectedModel.details?.quantization_level || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="格式">
          {{ selectedModel.details?.format || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="修改时间">
          {{ formatDate(selectedModel.modified_at) }}
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="infoDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Download, Refresh, Search, Box } from '@element-plus/icons-vue'
import { ListModels, PullModel, DeleteModel, ShowModel } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { ElMessage, ElMessageBox } from 'element-plus'

// 响应式数据
const models = ref([])
const loading = ref(false)
const searchQuery = ref('')
const pullDialogVisible = ref(false)
const infoDialogVisible = ref(false)
const selectedModel = ref(null)
const pullForm = ref({ name: '' })
const pullLoading = ref(false)
// 拉取进度相关
const pullProgress = ref(0)
const pullStatus = ref('')
const pullStatusText = ref('')
const pullCurrentTask = ref('')
const pullError = ref('')
const currentPullModel = ref('')

// 计算属性
const filteredModels = computed(() => {
  if (!searchQuery.value) {
    return models.value
  }
  return models.value.filter(model =>
    model.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN')
}

// 进度条格式化
const pullProgressFormat = (percentage) => {
  return `${percentage}%`
}

// 取消拉取
const cancelPull = () => {
  // 这里可以添加取消拉取的逻辑
  // 目前后端没有提供取消拉取的API，所以只是关闭对话框
  pullDialogVisible.value = false
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
        // 拉取完成后重新加载模型列表
        setTimeout(() => {
          loadModels()
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

// 加载模型列表
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

// 刷新模型列表
const refreshModels = () => {
  loadModels()
}

// 显示拉取模型对话框
const showPullDialog = () => {
  pullForm.value = { name: '' }
  pullDialogVisible.value = true
}

// 模型搜索建议
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

// 确认拉取模型
const confirmPullModel = async () => {
  if (!pullForm.value.name) {
    ElMessage.warning('请输入模型名称')
    return
  }

  const modelName = pullForm.value.name
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
    ElMessage.error('拉取模型失败')
  }
}

// 显示模型详情
const showModelInfo = async (model) => {
  try {
    // 获取详细信息
    const details = await ShowModel(model.name)
    selectedModel.value = { ...model, ...details }
    infoDialogVisible.value = true
  } catch (error) {
    console.error('获取模型详情失败:', error)
    // 如果获取详情失败，至少显示基本信息
    selectedModel.value = model
    infoDialogVisible.value = true
  }
}

// 删除模型
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
    loadModels() // 重新加载模型列表
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除模型失败:', error)
      ElMessage.error('删除模型失败')
    }
  }
}

// 初始化
onMounted(() => {
  loadModels()
  setupPullProgressListener()
  
  // 监听本地模型更新事件
  window.addEventListener('localModelsUpdated', () => {
    console.log('收到本地模型更新事件，重新加载模型列表')
    loadModels()
  })
})
</script>

<style scoped>
.models-container {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  height: 100%;
  display: flex;
  flex-direction: column;
}

body.dark-theme .models-container {
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

/* 暗色主题 - 表格样式 */
body.dark-theme .models-table :deep(.el-table) {
  background-color: #1e1e1e;
  color: #e4e6eb;
}

body.dark-theme .models-table :deep(.el-table__header-wrapper th) {
  background-color: #2d2d2d;
  color: #e4e6eb;
  border-bottom-color: #3c3c3c;
}

body.dark-theme .models-table :deep(.el-table__row) {
  border-bottom-color: #3c3c3c;
}

body.dark-theme .models-table :deep(.el-table__row:hover) {
  background-color: rgba(255, 255, 255, 0.05);
}

body.dark-theme .models-table :deep(.el-table__empty-block) {
  background-color: #1e1e1e;
  color: #a0aec0;
}
</style>