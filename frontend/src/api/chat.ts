import http from './http'
import type { ApiResponse } from '@/types/auth'
import type { Message, UnreadCounts } from '@/types/chat'

export const getChatHistory = (friendId: number, before?: number) =>
  http
    .get<ApiResponse<Message[]>>(`/chat/${friendId}`, {
      params: before ? { before } : {},
    })
    .then((r) => r.data)

export const markRead = (friendId: number) =>
  http.post<ApiResponse<null>>(`/chat/${friendId}/read`).then((r) => r.data)

export const getUnreadCounts = () =>
  http.get<ApiResponse<UnreadCounts>>('/chat/unread').then((r) => r.data)
