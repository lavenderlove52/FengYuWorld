import http from './http'
import type { ApiResponse } from '@/types/auth'
import type { PageResult } from '@/types/post'

export interface InviteCode {
  id: number
  code: string
  created_by: number
  is_used: boolean
  used_by?: number
  used_at?: string
  expires_at?: string
  created_at: string
  creator?: { nickname: string; username: string }
  used_by_user?: { nickname: string; username: string }
}

export const listInviteCodes = (page = 1, pageSize = 20) =>
  http
    .get<ApiResponse<PageResult<InviteCode>>>('/admin/invite-codes', { params: { page, page_size: pageSize } })
    .then((r) => r.data)

export const generateInviteCodes = (count: number, expiresAt?: string) =>
  http
    .post<ApiResponse<InviteCode[]>>('/admin/invite-codes', { count, expires_at: expiresAt || undefined })
    .then((r) => r.data)

export const deleteInviteCode = (id: number) =>
  http.delete<ApiResponse<null>>(`/admin/invite-codes/${id}`).then((r) => r.data)
