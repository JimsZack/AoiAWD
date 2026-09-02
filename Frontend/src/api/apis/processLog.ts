import api from '../index'
import type { ProcessLog } from '@/types'

export const getProcessLogs = (page = 0, count = 20) =>
  api.get<ProcessLog[]>(`/listprocess?page=${page}&count=${count}`)

export const getCurrentProcess = () =>
  api.get<ProcessLog[]>('/listcurrentprocess')

export const getActiveProcess = () =>
  api.get<ProcessLog[]>('/currentprocess')
