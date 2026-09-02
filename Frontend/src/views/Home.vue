<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '230px'" class="aside">
      <div class="logo">
        <svg viewBox="0 0 100 100" width="32" height="32">
          <defs>
            <linearGradient id="grad" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" style="stop-color:#667eea;stop-opacity:1" />
              <stop offset="100%" style="stop-color:#764ba2;stop-opacity:1" />
            </linearGradient>
          </defs>
          <circle cx="50" cy="50" r="45" fill="url(#grad)"/>
          <text x="50" y="60" font-family="Arial" font-size="28" font-weight="bold" fill="white" text-anchor="middle">G</text>
        </svg>
        <span v-show="!isCollapse" class="logo-text">GoAWD</span>
      </div>

      <el-menu
        :default-active="route.path"
        :collapse="isCollapse"
        router
        class="sidebar-menu"
      >
        <template v-for="item in menuItems" :key="item.path">
          <el-menu-item :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapse = !isCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/main' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentRoute?.meta?.title">
              {{ currentRoute.meta.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <a href="https://github.com/JimsZack/AoiAWD" target="_blank" class="github-link">
            <el-icon><Link /></el-icon>
            GitHub
          </a>
          <el-button type="danger" text @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
            退出
          </el-button>
        </div>
      </el-header>

      <el-main class="main">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useWebSocket } from '@/composables/useWebSocket'
import { useAlertStore } from '@/stores'

const route = useRoute()
const router = useRouter()
const alertStore = useAlertStore()
const isCollapse = ref(false)

const menuItems = [
  { path: '/main', title: '首页', icon: 'HomeFilled' },
  { path: '/webLog', title: 'Web日志', icon: 'Monitor' },
  { path: '/pwnLog', title: 'PWN日志', icon: 'Warning' },
  { path: '/fileLog', title: '文件日志', icon: 'Document' },
  { path: '/processLog', title: '进程日志', icon: 'Cpu' },
  { path: '/warnLog', title: '系统警告', icon: 'Bell' }
]

const currentRoute = computed(() => route)

const handleMessage = (type: string) => {
  if (type === 'alert') {
    alertStore.fetchAlerts()
  }
}

useWebSocket(handleMessage)

const handleLogout = async () => {
  await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
  sessionStorage.removeItem('token')
  router.push('/login')
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.aside {
  background: #1a1a2e;
  transition: width 0.3s;
  overflow: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo-text {
  color: white;
  font-size: 18px;
  font-weight: 600;
}

.sidebar-menu {
  border-right: none;
  background: transparent;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 230px;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: white;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.github-link {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #606266;
  text-decoration: none;
}

.github-link:hover {
  color: #667eea;
}

.main {
  background: #f0f2f5;
  padding: 20px;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
