<template>
  <div class="detail-page" v-loading="loading">
    <el-page-header @back="$router.back()" title="返回列表">
      <template #content>
        <span class="page-title">PWN日志详情</span>
      </template>
    </el-page-header>

    <template v-if="detail">
      <el-descriptions :column="2" border class="info-section">
        <el-descriptions-item label="Binary">{{ detail.binary }}</el-descriptions-item>
        <el-descriptions-item label="时间">{{ detail.timestamp }}</el-descriptions-item>
      </el-descriptions>

      <el-collapse class="detail-section">
        <el-collapse-item v-if="detail.stdout" title="STDOUT" name="stdout">
          <pre class="content-block" v-html="stdoutHexdump"></pre>
        </el-collapse-item>
        <el-collapse-item v-if="detail.stdin" title="STDIN" name="stdin">
          <pre class="content-block" v-html="stdinHexdump"></pre>
        </el-collapse-item>
      </el-collapse>

      <div class="action-bar">
        <el-button type="primary" @click="download('packet')">
          <el-icon><Download /></el-icon>
          一键重放
        </el-button>
        <el-button type="success" @click="download('stream')">
          <el-icon><Download /></el-icon>
          下载完整流
        </el-button>
        <el-button type="warning" @click="download('map')">
          <el-icon><Download /></el-icon>
          下载MAP
        </el-button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getPwnDetail, downloadPwn } from '@/api/apis'
import { hexdump, downloadBlob } from '@/utils'
import type { PwnLog } from '@/types'

const route = useRoute()
const detail = ref<PwnLog | null>(null)
const loading = ref(true)

const stdoutHexdump = computed(() => {
  if (!detail.value?.stdout) return '无数据'
  return hexdump(atob(detail.value.stdout))
})

const stdinHexdump = computed(() => {
  if (!detail.value?.stdin) return '无数据'
  return hexdump(atob(detail.value.stdin))
})

const fetchData = async () => {
  try {
    const { data } = await getPwnDetail(route.params.id as string)
    detail.value = data
  } finally {
    loading.value = false
  }
}

const download = async (type: 'packet' | 'stream' | 'map') => {
  const token = sessionStorage.getItem('token') || ''
  const { data } = await downloadPwn(route.params.id as string, type, token)
  const filenames: Record<string, string> = {
    packet: `replay_${route.params.id}.py`,
    stream: `stream_${route.params.id}.bin`,
    map: `map_${route.params.id}.txt`
  }
  downloadBlob(data, filenames[type])
}

onMounted(fetchData)
</script>

<style scoped>
.detail-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
}

.info-section {
  margin-top: 16px;
}

.detail-section {
  margin-top: 16px;
}

.content-block {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

.action-bar {
  margin-top: 20px;
  display: flex;
  gap: 12px;
}
</style>
