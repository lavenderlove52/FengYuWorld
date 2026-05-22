import http from './http'
import type { ApiResponse } from '@/types/auth'
import type { Category, Tag } from '@/types/category'

export const listCategories = () =>
  http.get<ApiResponse<Category[]>>('/categories').then((r) => r.data)

export const createCategory = (payload: { name: string; icon?: string; sort_order?: number }) =>
  http.post<ApiResponse<Category>>('/categories', payload).then((r) => r.data)

export const listTags = () => http.get<ApiResponse<Tag[]>>('/tags').then((r) => r.data)

export const createTag = (name: string) =>
  http.post<ApiResponse<Tag>>('/tags', { name }).then((r) => r.data)
