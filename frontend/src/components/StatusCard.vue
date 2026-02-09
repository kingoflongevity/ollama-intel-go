<template>
  <div class="card status-card">
    <div class="label">{{ title }}</div>
    <div class="value" :class="status">{{ value }}</div>
  </div>
</template>

<script setup>
defineProps({
  title: String,
  value: String,
  status: String
})
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

.status-card { 
  height: 88px; 
}

.label { 
  font-size: 14px; 
  opacity: .7;
  color: var(--text-secondary);
}

.value { 
  font-size: 20px; 
  font-weight: 600; 
  margin-top: 8px;
  color: var(--text-primary);
}

.value.ok { 
  color: #6CFFB0; 
}

/* 浅色主题特殊样式 */
body.light-theme .card {
  background: rgba(255, 255, 255, 0.75);
  box-shadow: 0 0 40px rgba(76,130,255,0.12);
}
</style>