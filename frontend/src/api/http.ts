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

export const FORCE_LOGOUT_EVENT = 'my-blog:force-logout'

// 鍐呰仈 forceLogout锛岄伩鍏嶄粠 utils/auth 瀵煎叆瀵艰嚧寰幆渚濊禆
// 锛坔ttp 鈫?utils/auth 鈫?stores/auth 鈫?api/auth 鈫?http锛?
function forceLogout(reason?: 'deleted' | 'locked' | 'inactivity') {
  clearTokens()
  window.dispatchEvent(new CustomEvent(FORCE_LOGOUT_EVENT, { detail: { reason } }))
  const query = reason ? `?reason=${reason}` : ''
  // deleted/locked 灞炰簬璐﹀彿绾у埆寮傚父锛屼换鎰忛〉闈㈤兘闇€瑕佽烦杞櫥褰曪紱
  // 鍏朵粬鎯呭喌锛堝 refresh token 澶辨晥锛夊彧鍦ㄥ彈淇濇姢璺敱涓婃墠璺宠浆锛?
  // 閬垮厤鍦ㄥ叕寮€椤甸潰鍥犺繃鏈?token 瑙﹀彂寮哄埗閲嶅畾鍚戯紙鎵嬫満鐧藉睆鏍瑰洜锛?
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
let queue: Array<{
  resolve: (token: string) => void
  reject: (error: unknown) => void
}> = []

const flushQueue = (token: string) => {
  queue.forEach(({ resolve }) => resolve(token))
  queue = []
}

const rejectQueue = (error: unknown) => {
  queue.forEach(({ reject }) => reject(error))
  queue = []
}

http.interceptors.response.use(
  (res) => res,
  async (error: AxiosError<ApiResponse<unknown>>) => {
    const original = error.config
    const status = error.response?.status
    const code = (error.response?.data as any)?.code

    // JWT 鏈夋晥浣嗙敤鎴峰凡琚垹闄?鍚婇攢
    if (status === 401 && code === 40102) {
      forceLogout('deleted')
      return Promise.reject(new Error('会话已失效，请重新登录'))
    }

    // 璐﹀彿宸茶閿佸畾
    if (status === 403 && code === 40301) {
      forceLogout('locked')
      return Promise.reject(new Error('璐﹀彿宸茶閿佸畾'))
    }

    // 鏃犲搷搴旓細寮辩綉/鏂綉锛堟墜鏈哄父瑙侊級
    if (!error.response) {
      return Promise.reject(new Error('缃戠粶杩炴帴澶辫触锛岃妫€鏌ョ綉缁滃悗閲嶈瘯'))
    }

    // 闈?401 閿欒锛氭彁鍙栧悗绔?message 瀛楁鐩存帴鎶涘嚭锛岄伩鍏嶅墠绔樉绀?"Request failed with status code 4xx"
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
    // 鐧诲綍/娉ㄥ唽绛?auth 鎺ュ彛銆佹垨銆岀櫥褰曞け璐ャ€嶄笟鍔＄爜锛氫笉灏濊瘯 refresh锛岀洿鎺ヨ繑鍥炲悗绔枃妗?
    if (!refreshToken || isAuthApi || isLoginFailed) {
      if (backendMsg) return Promise.reject(new Error(backendMsg))
      return Promise.reject(error)
    }

    if (refreshing) {
      return new Promise((resolve, reject) => {
        queue.push({
          resolve: (token) => {
            original.headers.Authorization = `Bearer ${token}`
            resolve(http(original))
          },
          reject,
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
      rejectQueue(backendMsg ? new Error(backendMsg) : error)
      if (backendMsg) return Promise.reject(new Error(backendMsg))
      return Promise.reject(error)
    } finally {
      refreshing = false
    }
  },
)

export default http
