import api from '../index'
import type { FileLog } from '@/types'

export const getFileLogs = (page = 0, count = 20) =>
  api.get<FileLog[]>(`/listfilesystem?page=${page}&count=${count}`)

export const downloadFile = (id: string, token: string) =>
  api.get(`/downloadfile?id=${id}`, {
    responseType: 'blob',
    headers: { Token: token }
  })
