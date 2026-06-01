import { ref, reactive } from 'vue'
import { apiBase } from '@/config/app'
import type { Message, UnreadCounts, WsOutMsg } from '@/types/chat'

const WS_RECONNECT_MAX = 30000

function buildWsUrl(): string {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  // apiBase 形如 "/api" 或 "http://xxx/api"
  const base = apiBase.startsWith('http')
    ? apiBase.replace(/^http/, 'ws')
    : `${proto}://${location.host}${apiBase}`
  return `${base}/ws/chat`
}

export function useChat() {
  const connected = ref(false)
  // key: friendId
  const messages = reactive<Record<number, Message[]>>({})
  const unreadCounts = reactive<UnreadCounts>({})
  // 当前正在聊天的好友 id（null 表示未打开聊天面板）
  const activeFriendId = ref<number | null>(null)

  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectDelay = 1000
  let myUserId = 0
  let stopped = false

  function connect(userId: number, token: string) {
    stopped = false
    myUserId = userId
    _connect(token)
  }

  function _connect(token: string) {
    if (stopped) return
    const url = `${buildWsUrl()}?token=${encodeURIComponent(token)}`
    ws = new WebSocket(url)

    ws.onopen = () => {
      connected.value = true
      reconnectDelay = 1000
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as WsOutMsg
        _handleServerMsg(msg)
      } catch {
        // ignore
      }
    }

    ws.onclose = () => {
      connected.value = false
      ws = null
      if (!stopped) {
        reconnectTimer = setTimeout(() => {
          reconnectDelay = Math.min(reconnectDelay * 2, WS_RECONNECT_MAX)
          _connect(token)
        }, reconnectDelay)
      }
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  function disconnect() {
    stopped = true
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    ws?.close()
    ws = null
    connected.value = false
  }

  function sendMessage(to: number, content: string, type: 'text' | 'image' | 'emoji' = 'text') {
    if (!ws || ws.readyState !== WebSocket.OPEN) return
    ws.send(JSON.stringify({ type: 'msg', to, content, msg_type: type }))
  }

  function markRead(friendId: number) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return
    ws.send(JSON.stringify({ type: 'read', friend_id: friendId }))
    unreadCounts[friendId] = 0
  }

  function _handleServerMsg(msg: WsOutMsg) {
    switch (msg.type) {
      case 'msg': {
        const m = msg.payload as Message
        // 好友 ID = 对方
        const friendId = m.from_user === myUserId ? m.to_user : m.from_user
        if (!messages[friendId]) messages[friendId] = []
        if (!messages[friendId].find((x) => x.id === m.id)) {
          messages[friendId].push(m)
        }
        // 新消息到达，仅在以下条件同时满足时累加未读：
        // 1. 不是自己发出的消息
        // 2. 消息未被读
        // 3. 当前没有打开与该好友的聊天窗口
        if (m.from_user !== myUserId && !m.is_read && activeFriendId.value !== friendId) {
          unreadCounts[friendId] = (unreadCounts[friendId] ?? 0) + 1
        }
        break
      }
      case 'unread_counts': {
        const counts = msg.payload as Record<string, number>
        for (const [k, v] of Object.entries(counts)) {
          unreadCounts[Number(k)] = v
        }
        break
      }
      case 'unread_update': {
        const p = msg.payload as { friend_id: number; count: number }
        unreadCounts[p.friend_id] = p.count
        break
      }
    }
  }

  /** 外部注入消息（从 REST API 加载历史） */
  function injectMessages(friendId: number, msgs: Message[]) {
    if (!messages[friendId]) {
      messages[friendId] = msgs
    } else {
      // 去重合并（历史消息在前）
      const existingIds = new Set(messages[friendId].map((m) => m.id))
      const newMsgs = msgs.filter((m) => !existingIds.has(m.id))
      messages[friendId] = [...newMsgs, ...messages[friendId]]
    }
  }

  return {
    connected,
    messages,
    unreadCounts,
    activeFriendId,
    connect,
    disconnect,
    sendMessage,
    markRead,
    injectMessages,
  }
}
