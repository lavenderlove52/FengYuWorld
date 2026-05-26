import http from './http'
import type { ApiResponse } from '@/types/auth'
import type { CreatePostPayload, PageResult, PostDetail, PostSummary } from '@/types/post'

export const listPublicPosts = (page = 1, pageSize = 12) =>
  http
    .get<ApiResponse<PageResult<PostSummary>>>('/posts/public', { params: { page, page_size: pageSize } })
    .then((r) => r.data)

export const searchPublicPosts = (keyword: string, page = 1, pageSize = 20) =>
  http
    .get<ApiResponse<PageResult<PostSummary>>>('/posts/public', { params: { page, page_size: pageSize, keyword } })
    .then((r) => r.data)

export const getPost = (id: number) =>
  http.get<ApiResponse<PostDetail>>(`/posts/${id}`).then((r) => r.data)

export const createPost = (payload: CreatePostPayload) =>
  http.post<ApiResponse<PostDetail>>('/posts', payload).then((r) => r.data)

export const updatePost = (id: number, payload: CreatePostPayload) =>
  http.put<ApiResponse<PostDetail>>(`/posts/${id}`, payload).then((r) => r.data)

export const deletePost = (id: number) =>
  http.delete<ApiResponse<null>>(`/posts/${id}`).then((r) => r.data)

export const listMyPosts = (params: { status?: string; visibility?: string; page?: number; page_size?: number }) =>
  http.get<ApiResponse<PageResult<PostSummary>>>('/me/posts', { params }).then((r) => r.data)

export const listUserPosts = (username: string, page = 1, pageSize = 12) =>
  http
    .get<ApiResponse<PageResult<PostSummary>>>(`/users/${username}/posts`, {
      params: { page, page_size: pageSize },
    })
    .then((r) => r.data)
