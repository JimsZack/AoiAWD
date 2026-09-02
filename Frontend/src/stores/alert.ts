import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getAlertLogs } from '@/api/apis'
import type { AlertLog } from '@/types'

export const useAlertStore = defineStore('alert', () => {
  const alerts = ref<AlertLog[]>([])
  const total = ref(0)
  const loading = ref(false)

  const fetchAlerts = async (page = 0, count = 8) => {
    loading.value = true
    try {
      const { data } = await getAlertLogs(page, count)
      alerts.value = data
      total.value = data.length
    } finally {
      loading.value = false
    }
  }

  return { alerts, total, loading, fetchAlerts }
})
