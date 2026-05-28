import { clearTokens } from '@/api/http'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

export type LogoutReason = 'deleted' | 'locked' | 'inactivity' | 'session_expired'

export function forceLogout(reason?: LogoutReason) {
  clearTokens()
  try {
    const auth = useAuthStore()
    auth.user = null
  } catch {
    // Pinia 可能在 Vue app 初始化前调用，忽略
  }
  const query = reason ? `?reason=${reason}` : ''
  router.push(`/login${query}`)
}
