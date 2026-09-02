<template>
  <LogTable
    :data="logs"
    :total="total"
    :loading="loading"
    :page-size="pageSize"
    @refresh="fetchLogs(currentPage)"
    @page-change="handlePageChange"
  >
    <template #filters>
      <el-switch v-model="showActive" active-text="只显示运行中" @change="fetchLogs(0)" />
    </template>

    <template #columns>
      <el-table-column prop="timestamp" label="时间" width="180" sortable />
      <el-table-column prop="uid" label="UID" width="80" />
      <el-table-column prop="ppid" label="父进程" width="100" />
      <el-table-column prop="pid" label="进程号" width="100" />
      <el-table-column prop="name" label="进程名" width="150" />
      <el-table-column prop="cmdline" label="启动参数" show-overflow-tooltip />
    </template>
  </LogTable>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import LogTable from '@/components/LogTable.vue'
import { getProcessLogs, getActiveProcess } from '@/api/apis'
import type { ProcessLog } from '@/types'

const logs = ref<ProcessLog[]>([])
const total = ref(0)
const loading = ref(false)
const currentPage = ref(0)
const pageSize = ref(20)
const showActive = ref(false)

const fetchLogs = async (page: number) => {
  loading.value = true
  try {
    if (showActive.value) {
      const { data } = await getActiveProcess()
      logs.value = data
    } else {
      const { data } = await getProcessLogs(page, pageSize.value)
      logs.value = data
    }
    total.value = logs.value.length
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchLogs(page)
}

onMounted(() => fetchLogs(0))
</script>
