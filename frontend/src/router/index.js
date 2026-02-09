import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'
import ChatView from '@/views/ChatView.vue'
import ModelsView from '@/views/ModelsView.vue'
import OnlineModelsView from '@/views/OnlineModelsView.vue'
import SettingsView from '@/views/SettingsView.vue'
import LogView from '@/views/LogView.vue'

const routes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: DashboardView
  },
  {
    path: '/chat',
    name: 'Chat',
    component: ChatView
  },
  {
    path: '/models',
    name: 'Models',
    component: ModelsView
  },
  {
    path: '/online',
    name: 'OnlineModels',
    component: OnlineModelsView
  },
  {
    path: '/settings',
    name: 'Settings',
    component: SettingsView
  },
  {
    path: '/logs',
    name: 'Logs',
    component: LogView
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router