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
      <el-select v-model="operationFilter" placeholder="操作类型" clearable size="small">
        <el-option v-for="op in operations" :key="op" :label="op" :value="op" />
      </el-select>
    </template>

    <template #columns>
      <el-table-column prop="timestamp" label="时间" width="180" sortable />
      <el-table-column prop="operation" label="操作" width="120">
        <template #default="{ row }">
          <el-tag :type="getOperationType(row.operation)" size="small">{{ row.operation }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="path" label="路径" show-overflow-tooltip>
        <template #default="{ row }">
          <span :class="{ 'is-dir': row.path.endsWith('/') }">{{ row.path }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="handleDownload(row)">下载</el-button>
        </template>
      </el-table-column>
    </template>
  </LogTable>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import LogTable from '@/components/LogTable.vue'
import { getFileLogs, downloadFile } from '@/api/apis'
import { downloadBlob } from '@/utils'
import type { FileLog } from '@/types'

const logs = ref<FileLog[]>([])
const total = ref(0)
const loading = ref(false)
const currentPage = ref(0)
const pageSize = ref(20)
const operationFilter = ref('')

const operations = ['CREATE', 'MODIFY', 'CLOSE_WRITE', 'ATTRIB', 'DELETE']

const getOperationType = (op: string) => {
  const map: Record<string, string> = {
    CREATE: 'success',
    MODIFY: 'warning',
    DELETE: 'danger',
    ATTRIB: 'info',
    CLOSE_WRITE: ''
  }
  return map[op] || ''
}

const fetchLogs = async (page: number) => {
  loading.value = true
  try {
    const { data } = await getFileLogs(page, pageSize.value)
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

const handleDownload = async (row: FileLog) => {
  const token = sessionStorage.getItem('token') || ''
  const { data } = await downloadFile(row.id, token)
  downloadBlob(data, row.path.split('/').pop() || 'file')
}

watch(operationFilter, () => {
  fetchLogs(0)
})

onMounted(() => fetchLogs(0))
</script>

<style scoped>
.is-dir {
  color: #67C23A;
}
</style>
