import { ref, onMounted, onUnmounted } from 'vue'
import config from '@/config'
import { useWebSocketStore } from '@/stores'
import { ElMessage } from 'element-plus'

export function useWebSocket(onMessage?: (type: string, data: unknown) => void) {
  const store = useWebSocketStore()
  const ws = ref<WebSocket | null>(null)
  const reconnectTimer = ref<number | null>(null)
  const reconnectDelay = ref(1000)
  const destroyed = ref(false)

  const connect = () => {
    if (ws.value || destroyed.value) return

    try {
      ws.value = new WebSocket(config.ws_addr)

      ws.value.onopen = () => {
        store.setConnected(true)
        reconnectDelay.value = 1000
        console.log('WebSocket connected')
      }

      ws.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          store.setMessage(data)
          onMessage?.(data.type, data)
        } catch (e) {
          console.error('WebSocket message parse error:', e)
        }
      }

      ws.value.onclose = () => {
        store.setConnected(false)
        ws.value = null
        if (!destroyed.value) {
          scheduleReconnect()
        }
      }

      ws.value.onerror = (error) => {
        console.error('WebSocket error:', error)
      }
    } catch (e) {
      console.error('WebSocket connection failed:', e)
      scheduleReconnect()
    }
  }

  const scheduleReconnect = () => {
    if (destroyed.value || reconnectTimer.value) return

    ElMessage.info('WebSocket连接丢失，正在尝试重连...')

    reconnectTimer.value = window.setTimeout(() => {
      reconnectTimer.value = null
      reconnectDelay.value = Math.min(reconnectDelay.value * 2, 10000)
      connect()
    }, reconnectDelay.value)
  }

  const disconnect = () => {
    destroyed.value = true
    if (reconnectTimer.value) {
      clearTimeout(reconnectTimer.value)
      reconnectTimer.value = null
    }
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
  }

  onMounted(connect)
  onUnmounted(disconnect)

  return { connect, disconnect }
}
