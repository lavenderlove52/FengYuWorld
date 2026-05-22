import http from './http'
import type { ApiResponse, AuthResponse, UserInfo } from '@/types/auth'

export const register = (payload: {
  username: string
  email: string
  password: string
  nickname: string
}) => http.post<ApiResponse<AuthResponse>>('/auth/register', payload).then((r) => r.data)

export const login = (payload: { login: string; password: string }) =>
  http.post<ApiResponse<AuthResponse>>('/auth/login', payload).then((r) => r.data)

export const logout = () => http.post<ApiResponse<null>>('/auth/logout').then((r) => r.data)

export const getMe = () => http.get<ApiResponse<UserInfo>>('/users/me').then((r) => r.data)

export const updateMe = (payload: { nickname?: string; avatar_url?: string; bio?: string }) =>
  http.put<ApiResponse<UserInfo>>('/users/me', payload).then((r) => r.data)

export const getUserProfile = (username: string) =>
  http.get<ApiResponse<UserInfo>>(`/users/${username}`).then((r) => r.data)
