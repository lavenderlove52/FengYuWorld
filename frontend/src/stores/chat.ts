import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useChat } from '@/composables/useChat'
import * as chatApi from '@/api/chat'
import type { Message } from '@/types/chat'

const chat = useChat()

// 每个好友的游标（最旧消息的 id），用于"加载更多"
const cursors = ref<Record<number, number | null>>({})

export const useChatStore = defineStore('chat', () => {
  const { connected, messages, unreadCounts, activeFriendId, connect, disconnect, sendMessage, markRead, injectMessages } = chat

  /** 初始化 WebSocket 连接 */
  function initWs(userId: number, token: string) {
    connect(userId, token)
  }

  /** 关闭连接 */
  function closeWs() {
    disconnect()
  }

  /** 设置当前活跃聊天好友 */
  function setActiveFriend(id: number | null) {
    activeFriendId.value = id
  }

  /** 加载历史消息（首次） */
  async function loadHistory(friendId: number) {
    const res = await chatApi.getChatHistory(friendId)
    if (res.code === 0 && res.data) {
      injectMessages(friendId, res.data)
      // 记录游标：最旧消息的 id
      cursors.value[friendId] = res.data.length > 0 ? res.data[0].id : null
    }
  }

  /** 加载更多历史消息（向上翻页） */
  async function loadMoreHistory(friendId: number): Promise<boolean> {
    const cursor = cursors.value[friendId]
    if (cursor === null) return false

    const res = await chatApi.getChatHistory(friendId, cursor)
    if (res.code === 0 && res.data && res.data.length > 0) {
      injectMessages(friendId, res.data)
      cursors.value[friendId] = res.data[0].id
      return res.data.length === 50
    }
    cursors.value[friendId] = null
    return false
  }

  /** 发送消息 */
  function send(to: number, content: string, type: Message['type'] = 'text') {
    sendMessage(to, content, type)
  }

  /** 标记已读 */
  function read(friendId: number) {
    markRead(friendId)
  }

  return {
    connected,
    messages,
    unreadCounts,
    activeFriendId,
    initWs,
    closeWs,
    setActiveFriend,
    loadHistory,
    loadMoreHistory,
    send,
    read,
  }
})
