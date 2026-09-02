import api from '../index'
import type { WebLog } from '@/types'

export const getWebLogs = (page = 0, count = 20) =>
  api.get<WebLog[]>(`/listweb?page=${page}&count=${count}`)

export const getWebDetail = (id: string) =>
  api.get<WebLog>(`/webdetail?id=${id}`)

export const downloadAutoScript = (id: string, token: string) => {
  const url = `/downloadwebautoscript?id=${id}`
  return api.get(url, {
    responseType: 'blob',
    headers: { Token: token }
  })
}
