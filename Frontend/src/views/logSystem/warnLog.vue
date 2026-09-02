<template>
  <LogTable
    :data="logs"
    :total="total"
    :loading="loading"
    :page-size="pageSize"
    @refresh="fetchLogs(currentPage)"
    @page-change="handlePageChange"
    @row-click="handleRowClick"
  >
    <template #columns>
      <el-table-column prop="timestamp" label="时间" width="180" sortable />
      <el-table-column prop="alertType" label="类型" width="120">
        <template #default="{ row }">
          <el-tag type="danger" size="small">{{ row.alertType }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="plugin" label="插件" width="120" />
      <el-table-column prop="description" label="描述" show-overflow-tooltip />
    </template>
  </LogTable>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import LogTable from '@/components/LogTable.vue'
import { getAlertLogs } from '@/api/apis'
import type { AlertLog } from '@/types'

const router = useRouter()
const logs = ref<AlertLog[]>([])
const total = ref(0)
const loading = ref(false)
const currentPage = ref(0)
const pageSize = ref(20)

const fetchLogs = async (page: number) => {
  loading.value = true
  try {
    const { data } = await getAlertLogs(page, pageSize.value)
    logs.value = data
    total.value = data.length
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchLogs(page)
}

const handleRowClick = (row: AlertLog) => {
  if (row.alertType === 'web') {
    router.push('/webLog')
  } else if (row.alertType === 'file') {
    router.push('/fileLog')
  } else if (row.alertType === 'process') {
    router.push('/processLog')
  }
}

onMounted(() => fetchLogs(0))
</script>
