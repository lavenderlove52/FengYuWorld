<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { NButton } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { activePageTheme } from '@/stores/pageTheme'
import { appBase } from '@/config/app'

const auth = useAuthStore()
const route = useRoute()
const logoUrl = `${appBase}logo.png`

const hasToken = computed(() => auth.tokenExists)
const showMyHomeLink = computed(() => route.name === 'post-detail' && !!auth.user?.username)
const avatarUrl = computed(() => auth.user?.avatar_url || '')
const avatarInitial = computed(() => auth.user?.nickname?.[0]?.toUpperCase() || 'U')

const headerBg = computed(() => activePageTheme.enabled ? activePageTheme.bg : '#f8fafc')
const headerBorder = computed(() => activePageTheme.enabled ? activePageTheme.border : 'transparent')
const headerText = computed(() => activePageTheme.enabled ? activePageTheme.text : '#0f172a')
</script>

<template>
  <header
    class="sticky top-0 z-10"
    :style="{
      backgroundColor: headerBg,
      borderBottom: `1px solid ${headerBorder}`,
      color: headerText,
    }"
  >
    <div class="w-full px-4 sm:px-6 h-14 flex items-center justify-between">
      <RouterLink to="/">
        <img :src="logoUrl" alt="风雨世界" class="h-9 w-auto object-contain" />
      </RouterLink>
      <div class="flex items-center gap-1 sm:gap-2">
        <RouterLink
          v-if="showMyHomeLink"
          :to="`/u/${auth.user!.username}`"
          class="w-9 h-9 rounded-full flex items-center justify-center hover:opacity-80 transition-opacity"
          :style="{ color: activePageTheme.themeColorActive ? activePageTheme.accent : headerText }"
          title="我的主页"
          aria-label="我的主页"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
            <path d="M10 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM3.465 14.493a1.23 1.23 0 0 0 .41 1.412A9.957 9.957 0 0 0 10 18c2.31 0 4.438-.784 6.131-2.1.43-.333.604-.903.408-1.41a7.002 7.002 0 0 0-13.074.003Z" />
          </svg>
        </RouterLink>
        <RouterLink v-if="hasToken" to="/dashboard" class="p-1.5 rounded-full">
          <div
            class="w-9 h-9 rounded-full overflow-hidden border flex items-center justify-center hover:opacity-80 transition-opacity cursor-pointer"
            :style="{ borderColor: headerBorder || '#d1d5db', backgroundColor: '#e5e7eb' }"
          >
            <img
              v-if="avatarUrl"
              :src="avatarUrl"
              :alt="auth.user?.nickname"
              class="w-full h-full object-cover"
            />
            <span v-else class="text-sm font-semibold select-none" :style="{ color: headerText }">
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
