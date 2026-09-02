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
    <template #filters>
      <el-select v-model="methodFilter" placeholder="方法筛选" clearable size="small">
        <el-option v-for="m in methods" :key="m" :label="m" :value="m" />
      </el-select>
    </template>

    <template #columns>
      <el-table-column prop="timestamp" label="时间" width="180" sortable />
      <el-table-column prop="method" label="方法" width="100">
        <template #default="{ row }">
          <el-tag :type="getMethodType(row.method)" size="small">{{ row.method }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="ip" label="IP" width="140" />
      <el-table-column prop="url" label="URL" show-overflow-tooltip />
    </template>
  </LogTable>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import LogTable from '@/components/LogTable.vue'
import { getWebLogs } from '@/api/apis'
import type { WebLog } from '@/types'

const router = useRouter()
const logs = ref<WebLog[]>([])
const total = ref(0)
const loading = ref(false)
const currentPage = ref(0)
const pageSize = ref(20)
const methodFilter = ref('')

const methods = ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'HEAD']

const getMethodType = (method: string) => {
  const map: Record<string, string> = {
    GET: 'success',
    POST: 'primary',
    PUT: 'warning',
    DELETE: 'danger',
    OPTIONS: 'info',
    HEAD: 'info'
  }
  return map[method] || ''
}

const fetchLogs = async (page: number) => {
  loading.value = true
  try {
    const { data } = await getWebLogs(page, pageSize.value)
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

const handleRowClick = (row: WebLog) => {
  router.push(`/webLog/detail/${row.id}`)
}

watch(methodFilter, () => {
  fetchLogs(0)
})

onMounted(() => fetchLogs(0))
</script>
