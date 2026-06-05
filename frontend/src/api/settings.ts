import http from './http'
import type { ApiResponse } from '@/types/auth'

export interface PublicSettings {
  friend_chat_enabled: boolean
}

export const getPublicSettings = () =>
  http.get<ApiResponse<PublicSettings>>('/settings/public').then((r) => r.data)

export const updateFriendChatSetting = (enabled: boolean) =>
  http.put<ApiResponse<PublicSettings>>('/super-admin/settings/friend-chat', { enabled }).then((r) => r.data)
