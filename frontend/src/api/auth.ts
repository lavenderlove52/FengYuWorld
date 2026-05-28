import http from './http'
import type { ApiResponse, AuthResponse, UserInfo } from '@/types/auth'
import type { GeetestResult } from '@/composables/useGeetest'

export interface RegisterPayload {
  username: string
  email: string
  password: string
  nickname: string
  email_code: string
  invite_code: string
  lot_number: string
  captcha_output: string
  pass_token: string
  gen_time: string
}

export interface LoginPayload {
  login: string
  password: string
  lot_number: string
  captcha_output: string
  pass_token: string
  gen_time: string
}

export const register = (payload: RegisterPayload) =>
  http.post<ApiResponse<AuthResponse>>('/auth/register', payload).then((r) => r.data)

export const login = (payload: LoginPayload) =>
  http.post<ApiResponse<AuthResponse>>('/auth/login', payload).then((r) => r.data)

export const sendEmailCode = (email: string, geetest?: Partial<GeetestResult>) =>
  http
    .post<ApiResponse<null>>('/auth/send-email-code', { email, ...geetest })
    .then((r) => r.data)

export const logout = () => http.post<ApiResponse<null>>('/auth/logout').then((r) => r.data)

export const getMe = () => http.get<ApiResponse<UserInfo>>('/users/me').then((r) => r.data)

export const updateMe = (payload: { nickname?: string; avatar_url?: string; bio?: string }) =>
  http.put<ApiResponse<UserInfo>>('/users/me', payload).then((r) => r.data)

export const getUserProfile = (username: string) =>
  http.get<ApiResponse<UserInfo>>(`/users/${username}`).then((r) => r.data)
