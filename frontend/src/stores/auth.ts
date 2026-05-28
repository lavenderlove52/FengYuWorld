import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '@/api/auth'
import { clearTokens, getAccessToken, setTokens } from '@/api/http'
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

  const login = async (login: string, password: string, geetest: { lot_number: string; captcha_output: string; pass_token: string; gen_time: string }) => {
    const res = await authApi.login({ login, password, ...geetest })
    if (res.code !== 0 || !res.data) {
      throw new Error(res.message || '登录失败')
    }
    setTokens(res.data.access_token, res.data.refresh_token)
    tokenExists.value = true
    user.value = res.data.user
    return res.data
  }

  const register = async (payload: {
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
  }) => {
    const res = await authApi.register(payload)
    if (res.code !== 0 || !res.data) {
      throw new Error(res.message || '注册失败')
    }
    setTokens(res.data.access_token, res.data.refresh_token)
    tokenExists.value = true
    user.value = res.data.user
    return res.data
  }

  const logout = async () => {
    try {
      await authApi.logout()
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
