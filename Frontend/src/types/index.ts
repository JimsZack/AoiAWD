export interface LogEntry {
  id: string
  timestamp: string
  type: 'web' | 'file' | 'process' | 'alert' | 'pwn'
}

export interface WebLog extends LogEntry {
  method: string
  ip: string
  url: string
  header?: string
  get?: string
  post?: string
  cookie?: string
  file?: string
  buffer?: string
}

export interface PwnLog extends LogEntry {
  binary: string
  stdin?: string
  stdout?: string
  groups?: string[]
}

export interface FileLog extends LogEntry {
  operation: string
  path: string
  content?: string
}

export interface ProcessLog extends LogEntry {
  uid: number
  ppid: number
  pid: number
  name: string
  cmdline?: string
}

export interface AlertLog extends LogEntry {
  alertType: string
  plugin: string
  description: string
}

export interface Plugin {
  name: string
  version: string
  description: string
}

export interface ServerInfo {
  uptime: string
  alertCount: number
  pluginCount: number
}

export interface PaginationParams {
  page: number
  count: number
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  count: number
}
