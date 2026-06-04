import http from './http'
import type { ApiResponse } from '@/types/auth'

export interface ImageUploadResult {
  url: string
  mobile_url?: string
}

export const uploadImage = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return http
    .post<ApiResponse<ImageUploadResult>>('/uploads/image', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then((r) => r.data)
}

/** 主页 Banner / Hero 背景（2K 级压缩，短边最高 1080px） */
export const uploadThemeBackground = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return http
    .post<ApiResponse<{ url: string }>>('/uploads/banner', form, {
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
