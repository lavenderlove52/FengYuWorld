import http from './http'
import type { ApiResponse } from '@/types/auth'
import type { ThemeImage, UserTheme, UpdateThemePayload } from '@/types/theme'

export interface ThemeImageList {
  items: ThemeImage[]
  total?: number
  max_count?: number
}

export const getUserTheme = (username: string) =>
  http.get<ApiResponse<UserTheme>>(`/users/${username}/theme`).then((r) => r.data)

export const updateMyTheme = (payload: UpdateThemePayload) =>
  http.put<ApiResponse<UserTheme>>('/me/theme', payload).then((r) => r.data)

export const listMyThemeImages = () =>
  http.get<ApiResponse<ThemeImage[] | ThemeImageList>>('/me/theme-images').then((r) => r.data)

export const uploadMyThemeImage = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return http
    .post<ApiResponse<ThemeImage>>('/me/theme-images', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then((r) => r.data)
}

export const deleteMyThemeImage = (id: number) =>
  http.delete<ApiResponse<null>>(`/me/theme-images/${id}`).then((r) => r.data)
