<template>
  <div class="card">
    <h3>最近对话</h3>
    <ul class="activity">
      <li v-for="chat in recentChats" :key="chat.id">{{ chat.content }}</li>
    </ul>
  </div>
</template>

<script setup>
import { defineProps } from 'vue'

defineProps({
  recentChats: {
    type: Array,
    default: () => []
  }
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

h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
}

.activity {
  margin-top: 16px;
  list-style: none;
  padding: 0;
}

.activity li {
  padding: 12px 0;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  opacity: .8;
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
  color: var(--text-primary);
}

.activity li:last-child {
  border-bottom: none;
}

/* 浅色主题特殊样式 */
body.light-theme .card {
  background: rgba(255, 255, 255, 0.75);
  box-shadow: 0 0 40px rgba(76,130,255,0.12);
}

body.light-theme .activity li {
  border-bottom: 1px solid rgba(0,0,0,0.05);
}
</style>