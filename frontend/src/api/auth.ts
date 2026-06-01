import http from './http'
import type { ApiResponse, AuthResponse, UserInfo } from '@/types/auth'

export interface RegisterPayload {
  username: string
  email: string
  password: string
  nickname: string
  email_code: string
  invite_code: string
  captcha_id: string
  captcha_answer: string
}

export interface LoginPayload {
  login: string
  password: string
  captcha_id: string
  captcha_answer: string
}

export interface CaptchaData {
  id: string
  image: string
}

export const register = (payload: RegisterPayload) =>
  http.post<ApiResponse<AuthResponse>>('/auth/register', payload).then((r) => r.data)

export const login = (payload: LoginPayload) =>
  http.post<ApiResponse<AuthResponse>>('/auth/login', payload).then((r) => r.data)

export const sendEmailCode = (email: string) =>
  http.post<ApiResponse<null>>('/auth/send-email-code', { email }).then((r) => r.data)

export const getCaptcha = () =>
  http.get<ApiResponse<CaptchaData>>('/auth/captcha').then((r) => r.data)

export interface ResetPasswordPayload {
  email: string
  email_code: string
  new_password: string
  captcha_id: string
  captcha_answer: string
}

export const resetPassword = (payload: ResetPasswordPayload) =>
  http.post<ApiResponse<null>>('/auth/reset-password', payload).then((r) => r.data)

export const logout = () => http.post<ApiResponse<null>>('/auth/logout').then((r) => r.data)

export const getMe = () => http.get<ApiResponse<UserInfo>>('/users/me').then((r) => r.data)

export const updateMe = (payload: { nickname?: string; avatar_url?: string; bio?: string }) =>
  http.put<ApiResponse<UserInfo>>('/users/me', payload).then((r) => r.data)

export const getUserProfile = (username: string) =>
  http.get<ApiResponse<UserInfo>>(`/users/${username}`).then((r) => r.data)
