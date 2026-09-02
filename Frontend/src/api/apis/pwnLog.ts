import api from '../index'
import type { PwnLog } from '@/types'

export const getPwnLogs = (page = 0, count = 20) =>
  api.get<PwnLog[]>(`/listpwn?page=${page}&count=${count}`)

export const getPwnDetail = (id: string) =>
  api.get<PwnLog>(`/pwndetail?id=${id}`)

export const downloadPwn = (id: string, type: 'packet' | 'stream' | 'map', token: string) => {
  const endpoints: Record<string, string> = {
    packet: '/downloadpwnautoscript',
    stream: '/downloadpwn',
    map: '/downloadpwn'
  }
  const url = `${endpoints[type]}?id=${id}${type === 'map' ? '&type=map' : ''}`
  return api.get(url, {
    responseType: 'blob',
    headers: { Token: token }
  })
}
