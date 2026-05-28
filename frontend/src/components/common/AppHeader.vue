<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { NButton } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { appBase } from '@/config/app'

const auth = useAuthStore()
const logoUrl = `${appBase}logo.png`

// 使用 auth store 中的响应式 tokenExists，确保登录/登出后能立即更新
const hasToken = computed(() => auth.tokenExists)
// avatar_url 为空字符串时视为无头像，显示首字母占位
const avatarUrl = computed(() => auth.user?.avatar_url || '')
const avatarInitial = computed(() => auth.user?.nickname?.[0]?.toUpperCase() || 'U')
</script>

<template>
  <header class="bg-[#f8fafc] sticky top-0 z-10">
    <div class="w-full px-6 h-14 flex items-center justify-between">
      <RouterLink to="/">
        <img :src="logoUrl" alt="风雨世界" class="h-9 w-auto object-contain" />
      </RouterLink>
      <div class="flex items-center gap-2">
        <RouterLink v-if="hasToken" to="/dashboard">
          <div
            class="w-9 h-9 rounded-full overflow-hidden border border-gray-300 bg-gray-200 flex items-center justify-center hover:opacity-80 transition-opacity cursor-pointer"
          >
            <img
              v-if="avatarUrl"
              :src="avatarUrl"
              :alt="auth.user?.nickname"
              class="w-full h-full object-cover"
            />
            <span v-else class="text-sm font-semibold text-gray-600 select-none">
              {{ avatarInitial }}
            </span>
          </div>
        </RouterLink>
        <template v-else>
          <RouterLink to="/login"><NButton size="small">登录</NButton></RouterLink>
          <RouterLink to="/register"><NButton type="primary" size="small">注册</NButton></RouterLink>
        </template>
      </div>
    </div>
  </header>
</template>
