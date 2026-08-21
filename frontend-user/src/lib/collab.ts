import type { Op } from './types'

export interface CollabHandlers {
  onSnapshot: (text: string, siteId: number, users: any[], locks: any[]) => void
  onOp: (op: Op) => void
  onPresence: (users: any[]) => void
  onLock: (msg: any) => void
  onError: (msg: string) => void
}

const PING_INTERVAL_MS = 25000

export function connectCollab(documentId: string, token: string, h: CollabHandlers) {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const url = `${proto}://${location.host}/ws?token=${encodeURIComponent(token)}&documentId=${documentId}`
  const ws = new WebSocket(url)
  let closed = false
  let pingTimer: number | undefined

  function clearPing() {
    if (pingTimer !== undefined) {
      clearInterval(pingTimer)
      pingTimer = undefined
    }
  }

  // Client-side keepalive: send an app-level ping every 25s to keep the
  // WebSocket alive and to detect half-open connections.
  function startPing() {
    clearPing()
    pingTimer = window.setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'ping' }))
      }
    }, PING_INTERVAL_MS)
  }

  ws.onopen = () => {
    startPing()
  }
  ws.onmessage = (ev) => {
    const msg = JSON.parse(ev.data)
    if (msg.type === 'snapshot') h.onSnapshot(msg.text, msg.siteId, msg.users || [], msg.locks || [])
    else if (msg.type === 'op' && msg.op) h.onOp(msg.op)
    else if (msg.type === 'presence') h.onPresence(msg.users || [])
    else if (msg.type === 'lock') h.onLock(msg)
    else if (msg.type === 'error') h.onError(msg.message || '协同错误')
    // pong messages are expected keepalive responses; no action needed.
  }
  ws.onclose = () => {
    closed = true
    clearPing()
  }
  ws.onerror = () => {
    clearPing()
  }
  return {
    send(payload: Record<string, unknown>) {
      if (!closed && ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify(payload))
    },
    close() {
      closed = true
      clearPing()
      ws.close()
    },
    raw: ws,
  }
}
