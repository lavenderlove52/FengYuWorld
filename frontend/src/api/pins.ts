import http from './http'
import type { ApiResponse } from '@/types/auth'
import type { PinnedPost } from '@/types/post'

// 首页置顶
export const getSitePins = () =>
  http.get<ApiResponse<PinnedPost[]>>('/posts/site-pins').then((r) => r.data)

export const addSitePin = (postId: number) =>
  http.post<ApiResponse<null>>(`/pins/site/${postId}`).then((r) => r.data)

export const removeSitePin = (postId: number) =>
  http.delete<ApiResponse<null>>(`/pins/site/${postId}`).then((r) => r.data)

// 个人置顶
export const getMyPins = () =>
  http.get<ApiResponse<PinnedPost[]>>('/me/pins').then((r) => r.data)

export const addUserPin = (postId: number) =>
  http.post<ApiResponse<null>>(`/me/posts/${postId}/pin`).then((r) => r.data)

export const removeUserPin = (postId: number) =>
  http.delete<ApiResponse<null>>(`/me/posts/${postId}/pin`).then((r) => r.data)

// 管理员授权
export const grantSitePinPermission = (userId: number) =>
  http.post<ApiResponse<null>>(`/admin/users/${userId}/site-pin`).then((r) => r.data)

export const revokeSitePinPermission = (userId: number) =>
  http.delete<ApiResponse<null>>(`/admin/users/${userId}/site-pin`).then((r) => r.data)
