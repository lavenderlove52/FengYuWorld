import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import type { ApiResponse, AuthResponse } from '@/types/auth'
import { apiBase } from '@/config/app'
import router from '@/router'

const ACCESS_KEY = 'access_token'
const REFRESH_KEY = 'refresh_token'

export const getAccessToken = () => localStorage.getItem(ACCESS_KEY)
export const getRefreshToken = () => localStorage.getItem(REFRESH_KEY)

export const setTokens = (access: string, refresh: string) => {
  localStorage.setItem(ACCESS_KEY, access)
  localStorage.setItem(REFRESH_KEY, refresh)
}

export const clearTokens = () => {
  localStorage.removeItem(ACCESS_KEY)
  localStorage.removeItem(REFRESH_KEY)
}

// 内联 forceLogout，避免从 utils/auth 导入导致循环依赖
// （http → utils/auth → stores/auth → api/auth → http）
async function forceLogout(reason?: 'deleted' | 'locked' | 'inactivity') {
  clearTokens()
  try {
    const { useAuthStore } = await import('@/stores/auth')
    const auth = useAuthStore()
    auth.user = null
    auth.tokenExists = false
  } catch {
    // Pinia 尚未初始化时忽略
  }
  const query = reason ? `?reason=${reason}` : ''
  // deleted/locked 属于账号级别异常，任意页面都需要跳转登录；
  // 其他情况（如 refresh token 失效）只在受保护路由上才跳转，
  // 避免在公开页面因过期 token 触发强制重定向（手机白屏根因）
  const alwaysRedirect = reason === 'deleted' || reason === 'locked'
  if (alwaysRedirect || router.currentRoute.value.meta.requiresAuth) {
    router.push(`/login${query}`)
  }
}

const http = axios.create({
  baseURL: apiBase,
  timeout: 15000,
})

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let refreshing = false
let queue: Array<(token: string) => void> = []

const flushQueue = (token: string) => {
  queue.forEach((cb) => cb(token))
  queue = []
}

http.interceptors.response.use(
  (res) => res,
  async (error: AxiosError<ApiResponse<unknown>>) => {
    const original = error.config
    const status = error.response?.status
    const code = (error.response?.data as any)?.code

    // JWT 有效但用户已被删除/吊销
    if (status === 401 && code === 40102) {
      forceLogout('deleted')
      return Promise.reject(new Error('会话已失效，请重新登录'))
    }

    // 账号已被锁定
    if (status === 403 && code === 40301) {
      forceLogout('locked')
      return Promise.reject(new Error('账号已被锁定'))
    }

    // 无响应：弱网/断网（手机常见）
    if (!error.response) {
      return Promise.reject(new Error('网络连接失败，请检查网络后重试'))
    }

    // 非 401 错误：提取后端 message 字段直接抛出，避免前端显示 "Request failed with status code 4xx"
    if (status !== 401) {
      const backendMsg = (error.response?.data as any)?.message
      if (backendMsg) {
        return Promise.reject(new Error(backendMsg))
      }
      return Promise.reject(error)
    }

    if (!original) return Promise.reject(error)

    const backendMsg = (error.response?.data as any)?.message
    const refreshToken = getRefreshToken()
    const isAuthApi = /\/auth\//.test(original.url || '')
    const isLoginFailed = code === 40002
    // 登录/注册等 auth 接口、或「登录失败」业务码：不尝试 refresh，直接返回后端文案
    if (!refreshToken || isAuthApi || isLoginFailed) {
      if (backendMsg) return Promise.reject(new Error(backendMsg))
      return Promise.reject(error)
    }

    if (refreshing) {
      return new Promise((resolve) => {
        queue.push((token) => {
          original.headers.Authorization = `Bearer ${token}`
          resolve(http(original))
        })
      })
    }

    refreshing = true
    try {
      const { data } = await axios.post<ApiResponse<AuthResponse>>(`${apiBase}/auth/refresh`, {
        refresh_token: refreshToken,
      })
      if (data.code !== 0 || !data.data) {
        throw new Error('refresh failed')
      }
      setTokens(data.data.access_token, data.data.refresh_token)
      flushQueue(data.data.access_token)
      original.headers.Authorization = `Bearer ${data.data.access_token}`
      return http(original)
    } catch {
      forceLogout()
      if (backendMsg) return Promise.reject(new Error(backendMsg))
      return Promise.reject(error)
    } finally {
      refreshing = false
    }
  },
)

export default http
