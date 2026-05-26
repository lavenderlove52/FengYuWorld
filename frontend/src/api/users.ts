export { getMe, updateMe, getUserProfile } from './auth'

import http from './http'
import type { ApiResponse } from '@/types/auth'
import type { UserInfo } from '@/types/auth'
import type { PageResult } from '@/types/post'

export const listUsers = (keyword = '', page = 1, pageSize = 20) =>
  http
    .get<ApiResponse<PageResult<UserInfo>>>('/admin/users', { params: { keyword, page, page_size: pageSize } })
    .then((r) => r.data)
