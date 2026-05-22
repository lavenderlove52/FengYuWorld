import http from './http'
import type { ApiResponse } from '@/types/auth'

export const uploadImage = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return http
    .post<ApiResponse<{ url: string }>>('/uploads/image', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then((r) => r.data)
}

export const uploadAvatar = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return http
    .post<ApiResponse<{ url: string }>>('/uploads/avatar', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then((r) => r.data)
}
