import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useWebSocketStore = defineStore('websocket', () => {
  const connected = ref(false)
  const lastMessage = ref<{ type: string; data?: unknown } | null>(null)

  const setConnected = (value: boolean) => {
    connected.value = value
  }

  const setMessage = (message: { type: string; data?: unknown }) => {
    lastMessage.value = message
  }

  return { connected, lastMessage, setConnected, setMessage }
})
