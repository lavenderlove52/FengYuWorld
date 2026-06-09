import http from './http'
import type { ApiResponse } from '@/types/auth'
import type {
  AddPostTreePostPayload,
  CreatePostTreeFolderPayload,
  PostTreeNode,
  UpdatePostTreeNodePayload,
} from '@/types/postTree'

export const listPostTree = () =>
  http.get<ApiResponse<PostTreeNode[]>>('/me/post-tree').then((r) => r.data)

export const createPostTreeFolder = (payload: CreatePostTreeFolderPayload) =>
  http.post<ApiResponse<PostTreeNode>>('/me/post-tree/folders', payload).then((r) => r.data)

export const addPostTreePost = (payload: AddPostTreePostPayload) =>
  http.post<ApiResponse<PostTreeNode>>('/me/post-tree/posts', payload).then((r) => r.data)

export const updatePostTreeNode = (id: number, payload: UpdatePostTreeNodePayload) =>
  http.put<ApiResponse<PostTreeNode>>(`/me/post-tree/nodes/${id}`, payload).then((r) => r.data)

export const deletePostTreeNode = (id: number) =>
  http.delete<ApiResponse<null>>(`/me/post-tree/nodes/${id}`).then((r) => r.data)
