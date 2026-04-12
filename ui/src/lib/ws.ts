export type WsEvent = {
  type: string
  data: any
}

export type ConnectionState = 'connecting' | 'connected' | 'disconnected'

export function createWebSocket(
  onEvent: (event: WsEvent) => void,
  onStateChange: (state: ConnectionState) => void
) {
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null

  function connect() {
    onStateChange('connecting')

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    ws = new WebSocket(`${protocol}//${location.host}/ws`)

    ws.onopen = () => {
      onStateChange('connected')
    }

    ws.onmessage = (msg) => {
      try {
        const event: WsEvent = JSON.parse(msg.data)
        onEvent(event)
      } catch {}
    }

    ws.onclose = () => {
      onStateChange('disconnected')
      reconnectTimer = setTimeout(connect, 2000)
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  function disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
  }

  connect()

  return { disconnect }
}
