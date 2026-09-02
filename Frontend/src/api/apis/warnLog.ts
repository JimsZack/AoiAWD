import api from '../index'
import type { AlertLog } from '@/types'

export const getAlertLogs = (page = 0, count = 20) =>
  api.get<AlertLog[]>(`/listalert?page=${page}&count=${count}`)
