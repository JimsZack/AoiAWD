<template>
  <div class="detail-page" v-loading="loading">
    <el-page-header @back="$router.back()" title="返回列表">
      <template #content>
        <span class="page-title">Web日志详情</span>
      </template>
    </el-page-header>

    <template v-if="detail">
      <el-descriptions :column="2" border class="info-section">
        <el-descriptions-item label="URL">{{ detail.url }}</el-descriptions-item>
        <el-descriptions-item label="Remote">{{ detail.ip }}</el-descriptions-item>
        <el-descriptions-item label="Method">
          <el-tag>{{ detail.method }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="时间">{{ detail.timestamp }}</el-descriptions-item>
      </el-descriptions>

      <el-collapse class="detail-section">
        <el-collapse-item title="Header" name="header">
          <pre class="content-block">{{ detail.header || '无数据' }}</pre>
        </el-collapse-item>
        <el-collapse-item title="GET参数" name="get">
          <pre class="content-block">{{ detail.get || '无数据' }}</pre>
        </el-collapse-item>
        <el-collapse-item title="POST参数" name="post">
          <pre class="content-block">{{ detail.post || '无数据' }}</pre>
        </el-collapse-item>
        <el-collapse-item title="Cookie" name="cookie">
          <pre class="content-block">{{ detail.cookie || '无数据' }}</pre>
        </el-collapse-item>
        <el-collapse-item title="File" name="file">
          <pre class="content-block">{{ detail.file || '无数据' }}</pre>
        </el-collapse-item>
        <el-collapse-item title="Buffer" name="buffer">
          <pre class="content-block" v-html="bufferHexdump"></pre>
        </el-collapse-item>
      </el-collapse>

      <div class="action-bar">
        <el-button type="primary" @click="handleDownload">
          <el-icon><Download /></el-icon>
          一键重放
        </el-button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getWebDetail, downloadAutoScript } from '@/api/apis'
import { hexdump, downloadBlob } from '@/utils'
import type { WebLog } from '@/types'

const route = useRoute()
const detail = ref<WebLog | null>(null)
const loading = ref(true)

const bufferHexdump = computed(() => {
  if (!detail.value?.buffer) return '无数据'
  return hexdump(atob(detail.value.buffer))
})

const fetchData = async () => {
  try {
    const { data } = await getWebDetail(route.params.id as string)
    detail.value = data
  } finally {
    loading.value = false
  }
}

const handleDownload = async () => {
  const token = sessionStorage.getItem('token') || ''
  const response = await downloadAutoScript(route.params.id as string, token)
  downloadBlob(response.data, `replay_${route.params.id}.php`)
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
