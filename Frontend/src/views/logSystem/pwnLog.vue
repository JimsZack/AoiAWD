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
      <el-table-column prop="binary" label="执行文件" width="200" show-overflow-tooltip />
      <el-table-column prop="stdin" label="STDIN" show-overflow-tooltip />
      <el-table-column prop="stdout" label="STDOUT" show-overflow-tooltip />
    </template>
  </LogTable>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import LogTable from '@/components/LogTable.vue'
import { getPwnLogs } from '@/api/apis'
import type { PwnLog } from '@/types'

const router = useRouter()
const logs = ref<PwnLog[]>([])
const total = ref(0)
const loading = ref(false)
const currentPage = ref(0)
const pageSize = ref(20)

const fetchLogs = async (page: number) => {
  loading.value = true
  try {
    const { data } = await getPwnLogs(page, pageSize.value)
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

const handleRowClick = (row: PwnLog) => {
  router.push(`/pwnLog/detail/${row.id}`)
}

onMounted(() => fetchLogs(0))
</script>
