<template>
  <div class="main-page">
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="24" :sm="12" :md="6">
        <StatCard
          label="WebSocket"
          :value="wsStore.connected ? '已连接' : '未连接'"
          icon="Connection"
          :color="wsStore.connected ? '#67C23A' : '#F56C6C'"
        />
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <StatCard label="报警次数" :value="alertStore.total" icon="Warning" color="#E6A23C" />
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <StatCard label="运行时间" :value="info.uptime || '-'" icon="Timer" color="#909399" />
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <StatCard label="插件数" :value="plugins.length" icon="Cpu" color="#409EFF" />
      </el-col>
    </el-row>

    <el-row :gutter="20" class="content-row">
      <el-col :xs="24" :lg="16">
        <el-card class="content-card">
          <template #header>
            <div class="card-header">
              <span>系统警告</span>
              <el-button text type="primary" @click="$router.push('/warnLog')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="alertStore.alerts" stripe v-loading="alertStore.loading">
            <el-table-column prop="timestamp" label="时间" width="180" />
            <el-table-column prop="alertType" label="类型" width="100" />
            <el-table-column prop="plugin" label="插件" width="120" />
            <el-table-column prop="description" label="描述" show-overflow-tooltip />
          </el-table>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="8">
        <el-card class="content-card">
          <template #header>
            <div class="card-header">
              <span>已载入插件</span>
              <el-button text type="primary" :loading="reloading" @click="handleReload">
                <el-icon><Refresh /></el-icon>
                重载
              </el-button>
            </div>
          </template>
          <el-table :data="plugins" stripe>
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="version" label="版本" width="100" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <div class="last-update">
      最后更新: {{ lastUpdate }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import StatCard from '@/components/StatCard.vue'
import { useWebSocketStore, useAlertStore } from '@/stores'
import { getInfo, getPlugins, reloadPlugins } from '@/api/apis'
import { useWebSocket } from '@/composables/useWebSocket'
import type { Plugin, ServerInfo } from '@/types'

const wsStore = useWebSocketStore()
const alertStore = useAlertStore()
const plugins = ref<Plugin[]>([])
const info = ref<ServerInfo>({ uptime: '-', alertCount: 0, pluginCount: 0 })
const reloading = ref(false)
const lastUpdate = ref(new Date().toLocaleString('zh-CN'))

const fetchData = async () => {
  try {
    const [infoRes, pluginsRes] = await Promise.all([
      getInfo(),
      getPlugins()
    ])
    info.value = infoRes.data
    plugins.value = pluginsRes.data
    lastUpdate.value = new Date().toLocaleString('zh-CN')
  } catch (e) {
    console.error('Fetch data error:', e)
  }
}

const handleReload = async () => {
  reloading.value = true
  try {
    await reloadPlugins()
    await fetchData()
    ElMessage.success('插件重载成功')
  } catch {
    ElMessage.error('插件重载失败')
  } finally {
    reloading.value = false
  }
}

const handleMessage = (type: string) => {
  if (type === 'alert') {
    alertStore.fetchAlerts()
  }
  fetchData()
}

useWebSocket(handleMessage)

onMounted(() => {
  fetchData()
  alertStore.fetchAlerts()
})
</script>

<style scoped>
.main-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.stat-row .el-col {
  margin-bottom: 12px;
}

.content-card {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.last-update {
  text-align: right;
  font-size: 12px;
  color: #909399;
}
</style>
