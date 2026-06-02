import { clearTokens } from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

export type LogoutReason = 'deleted' | 'locked' | 'inactivity' | 'session_expired'

export function forceLogout(reason?: LogoutReason) {
  clearTokens()
  try {
    const auth = useAuthStore()
    auth.user = null
    auth.tokenExists = false
  } catch {
    // Pinia 可能在 Vue app 初始化前调用，忽略
  }
  const query = reason ? `?reason=${reason}` : ''
  // deleted/locked 属于账号级别异常，任意页面均需跳转登录；
  // inactivity/session_expired/无原因 只在受保护路由上跳转，
  // 避免公开页面因过期 token 或其他 Tab 登出而被强制重定向
  const alwaysRedirect = reason === 'deleted' || reason === 'locked'
  if (alwaysRedirect || router.currentRoute.value.meta.requiresAuth) {
    router.push(`/login${query}`)
  }
}
