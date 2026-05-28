import http from './http'
import type { ApiResponse } from '@/types/auth'
import type { UserTheme, UpdateThemePayload } from '@/types/theme'

export const getUserTheme = (username: string) =>
  http.get<ApiResponse<UserTheme>>(`/users/${username}/theme`).then((r) => r.data)

export const updateMyTheme = (payload: UpdateThemePayload) =>
  http.put<ApiResponse<UserTheme>>('/me/theme', payload).then((r) => r.data)
