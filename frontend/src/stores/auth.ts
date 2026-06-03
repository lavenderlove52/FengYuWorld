import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '@/api/auth'
import { clearTokens, getAccessToken, getRefreshToken, setTokens } from '@/api/http'
import type { UserInfo } from '@/types/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserInfo | null>(null)
  const loading = ref(false)
  // 响应式追踪 token 是否存在，确保 computed/watch 可感知登录/登出状态变化
  const tokenExists = ref<boolean>(!!getAccessToken())

  const isLoggedIn = computed(() => tokenExists.value && !!user.value)

  const fetchMe = async () => {
    if (!getAccessToken()) return
    loading.value = true
    try {
      const res = await authApi.getMe()
      if (res.code === 0 && res.data) {
        user.value = res.data
      }
    } finally {
      loading.value = false
    }
  }

  const login = async (login: string, password: string, captchaId: string, captchaAnswer: string) => {
    // 清除过期 token，避免请求携带无效 Authorization 或 guest 路由误判
    clearTokens()
    tokenExists.value = false
    user.value = null

    const res = await authApi.login({
      login: login.trim(),
      password,
      captcha_id: captchaId,
      captcha_answer: captchaAnswer.trim(),
    })
    if (res.code !== 0 || !res.data) {
      throw new Error(res.message || '登录失败')
    }
    setTokens(res.data.access_token, res.data.refresh_token)
    tokenExists.value = true
    await fetchMe()
    if (!user.value) {
      clearTokens()
      tokenExists.value = false
      throw new Error('登录成功但获取用户信息失败，请检查网络后重试')
    }
    return res.data
  }

  const register = async (payload: {
    username: string
    email: string
    password: string
    nickname: string
    email_code: string
    invite_code: string
    captcha_id: string
    captcha_answer: string
  }) => {
    clearTokens()
    tokenExists.value = false
    user.value = null

    const res = await authApi.register({
      ...payload,
      captcha_answer: payload.captcha_answer.trim(),
    })
    if (res.code !== 0 || !res.data) {
      throw new Error(res.message || '注册失败')
    }
    setTokens(res.data.access_token, res.data.refresh_token)
    tokenExists.value = true
    // 注册接口返回的 user 可能缺少 role 等字段，直接请求 /users/me 获取完整数据
    await fetchMe()
    return res.data
  }

  const logout = async () => {
    const refreshToken = getRefreshToken()
    try {
      await authApi.logout(refreshToken)
    } finally {
      clearTokens()
      tokenExists.value = false
      user.value = null
    }
  }

  const updateProfile = async (payload: { nickname?: string; avatar_url?: string; bio?: string }) => {
    const res = await authApi.updateMe(payload)
    if (res.code !== 0 || !res.data) {
      throw new Error(res.message || '更新失败')
    }
    user.value = res.data
  }

  return { user, loading, isLoggedIn, tokenExists, fetchMe, login, register, logout, updateProfile }
})
