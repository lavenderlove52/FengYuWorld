import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '@/api/auth'
import { clearTokens, getAccessToken, setTokens } from '@/api/http'
import type { UserInfo } from '@/types/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserInfo | null>(null)
  const loading = ref(false)

  const isLoggedIn = computed(() => !!getAccessToken() && !!user.value)

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

  const login = async (login: string, password: string) => {
    const res = await authApi.login({ login, password })
    if (res.code !== 0 || !res.data) {
      throw new Error(res.message || '登录失败')
    }
    setTokens(res.data.access_token, res.data.refresh_token)
    user.value = res.data.user
    return res.data
  }

  const register = async (payload: {
    username: string
    email: string
    password: string
    nickname: string
  }) => {
    const res = await authApi.register(payload)
    if (res.code !== 0 || !res.data) {
      throw new Error(res.message || '注册失败')
    }
    setTokens(res.data.access_token, res.data.refresh_token)
    user.value = res.data.user
    return res.data
  }

  const logout = async () => {
    try {
      await authApi.logout()
    } finally {
      clearTokens()
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

  return { user, loading, isLoggedIn, fetchMe, login, register, logout, updateProfile }
})
