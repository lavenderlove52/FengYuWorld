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
    if (!original || error.response?.status !== 401) {
      return Promise.reject(error)
    }

    const refreshToken = getRefreshToken()
    if (!refreshToken || original.url?.includes('/auth/refresh')) {
      clearTokens()
      if (router.currentRoute.value.path.startsWith('/dashboard')) {
        router.push('/login')
      }
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
      clearTokens()
      router.push('/login')
      return Promise.reject(error)
    } finally {
      refreshing = false
    }
  },
)

export default http
