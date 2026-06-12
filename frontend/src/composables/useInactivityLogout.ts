import { onMounted, onUnmounted } from 'vue'
import { getAccessToken } from '@/api/http'
import { forceLogout } from '@/utils/auth'

const INACTIVITY_MS = 7 * 24 * 60 * 60 * 1000 // 7 天

export function useInactivityLogout() {
  let timer: ReturnType<typeof setTimeout> | null = null

  const reset = () => {
    if (!getAccessToken()) return
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      if (getAccessToken()) {
        forceLogout('inactivity')
      }
    }, INACTIVITY_MS)
  }

  const onStorageChange = (e: StorageEvent) => {
    if (e.key === 'access_token' && !e.newValue && getAccessToken() === null) {
      // token 被其他 Tab 清除，当前 Tab 同步退出
      forceLogout('session_expired')
    }
  }

  const EVENTS = ['mousemove', 'mousedown', 'keydown', 'touchstart', 'scroll', 'click'] as const

  onMounted(() => {
    reset()
    EVENTS.forEach((e) => window.addEventListener(e, reset, { passive: true }))
    window.addEventListener('storage', onStorageChange)
  })

  onUnmounted(() => {
    if (timer) clearTimeout(timer)
    EVENTS.forEach((e) => window.removeEventListener(e, reset))
    window.removeEventListener('storage', onStorageChange)
  })

  return { reset }
}
