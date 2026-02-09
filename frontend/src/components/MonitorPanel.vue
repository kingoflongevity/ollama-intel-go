<template>
  <div class="card">
    <h3>系统监控</h3>
    <div class="monitor-grid">
      <!-- GPU 利用率图表 -->
      <div class="chart-container">
        <div class="chart-header">
          <div class="chart-title">GPU Utilization</div>
        </div>
        <div class="chart-content">
          <div class="chart-line-container">
            <svg class="chart-line" viewBox="0 0 400 200">
              <path d="M0,150 Q50,140 100,130 T200,120 T300,110 T400,100" stroke="var(--accent-primary)" stroke-width="2" fill="none" />
              <defs>
                <linearGradient id="gpuGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                  <stop offset="0%" :style="{ 'stop-color': 'var(--accent-primary)', 'stop-opacity': '0.3' }" />
                  <stop offset="100%" :style="{ 'stop-color': 'var(--accent-primary)', 'stop-opacity': '0' }" />
                </linearGradient>
              </defs>
              <path d="M0,150 Q50,140 100,130 T200,120 T300,110 T400,100 L400,200 L0,200 Z" fill="url(#gpuGradient)" />
            </svg>
          </div>
          <div class="chart-labels">
            <div class="label-item">
              <span class="label-time">Last Hour</span>
            </div>
            <div class="label-item">
              <span class="label-time">30 min</span>
            </div>
            <div class="label-item">
              <span class="label-time">1 hour</span>
            </div>
          </div>
          <div class="chart-values">
            <div class="value-item">95%</div>
            <div class="value-item">70%</div>
            <div class="value-item">40%</div>
            <div class="value-item">20%</div>
          </div>
        </div>
      </div>

      <!-- 内存使用图表 -->
      <div class="chart-container">
        <div class="chart-header">
          <div class="chart-title">Memory Usage</div>
          <div class="memory-info">
            <div class="memory-usage">{{ stats.totalModels || 0 }} Models</div>
          </div>
        </div>
        <div class="chart-content">
          <div class="chart-line-container">
            <svg class="chart-line" viewBox="0 0 400 200">
              <path d="M0,120 Q50,110 100,100 T200,90 T300,80 T400,70" stroke="var(--accent-secondary)" stroke-width="2" fill="none" />
              <defs>
                <linearGradient id="memoryGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                  <stop offset="0%" :style="{ 'stop-color': 'var(--accent-secondary)', 'stop-opacity': '0.3' }" />
                  <stop offset="100%" :style="{ 'stop-color': 'var(--accent-secondary)', 'stop-opacity': '0' }" />
                </linearGradient>
              </defs>
              <path d="M0,120 Q50,110 100,100 T200,90 T300,80 T400,70 L400,200 L0,200 Z" fill="url(#memoryGradient)" />
            </svg>
          </div>
          <div class="chart-labels">
            <div class="label-item">
              <span class="label-time">Last Hour</span>
            </div>
            <div class="label-item">
              <span class="label-time">30 min</span>
            </div>
            <div class="label-item">
              <span class="label-time">1 hour</span>
            </div>
          </div>
          <div class="chart-values">
            <div class="value-item">20.0 GB</div>
            <div class="value-item">12.0 GB</div>
            <div class="value-item">8.0 GB</div>
            <div class="value-item">4.0 GB</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { defineProps } from 'vue'

defineProps({
  stats: {
    type: Object,
    default: () => ({})
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
  margin-bottom: 24px;
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
}

.monitor-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
  position: relative;
  z-index: 1;
}

.chart-container {
  position: relative;
  padding: 20px;
  background: rgba(6, 182, 212, 0.05);
  border-radius: 12px;
  border: 1px solid var(--glass-border);
  backdrop-filter: blur(8px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
  height: 240px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.chart-header .chart-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
}

.memory-info {
  font-size: 14px;
  color: var(--accent-secondary);
  font-weight: 600;
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
}

.chart-content {
  position: relative;
  height: calc(100% - 40px);
}

.chart-line-container {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 40px;
}

.chart-line {
  width: 100%;
  height: 100%;
}

.chart-labels {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  justify-content: space-between;
  padding-top: 12px;
  border-top: 1px solid rgba(6, 182, 212, 0.2);
}

.label-item {
  font-size: 12px;
  color: var(--text-secondary);
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
}

.chart-values {
  position: absolute;
  top: 0;
  left: -40px;
  height: calc(100% - 40px);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding-right: 12px;
}

.value-item {
  font-size: 12px;
  color: var(--text-secondary);
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
}

/* 浅色主题特殊样式 */
body.light-theme .card {
  background: rgba(255, 255, 255, 0.75);
  box-shadow: 0 0 40px rgba(76,130,255,0.12);
}

body.light-theme .chart-container {
  background: rgba(2, 132, 199, 0.05);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

body.light-theme .chart-labels {
  border-top: 1px solid rgba(2, 132, 199, 0.2);
}
</style>