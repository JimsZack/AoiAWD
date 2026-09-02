<template>
  <div class="log-table" v-loading="loading">
    <div class="table-toolbar">
      <slot name="filters" />
      <div class="toolbar-right">
        <el-button size="small" @click="$emit('refresh')">前往最新</el-button>
        <el-switch
          v-model="autoSync"
          active-text="实时同步"
          @change="$emit('autoSync', $event)"
        />
      </div>
    </div>

    <el-table
      :data="data"
      stripe
      highlight-current-row
      @row-dblclick="$emit('rowClick', $event)"
    >
      <slot name="columns" />
    </el-table>

    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        layout="prev, pager, next, jumper"
        @current-change="$emit('pageChange', $event - 1)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  data: unknown[]
  total: number
  loading: boolean
  pageSize?: number
}>()

defineEmits<{
  refresh: []
  autoSync: [value: boolean]
  pageChange: [page: number]
  rowClick: [row: any]
}>()

const currentPage = ref(1)
const autoSync = ref(false)
</script>

<style scoped>
.log-table {
  padding: 16px;
}

.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}
</style>
