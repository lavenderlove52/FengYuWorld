<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { NButton } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { activePageTheme } from '@/stores/pageTheme'
import { appBase } from '@/config/app'

const auth = useAuthStore()
const logoUrl = `${appBase}logo.png`

const hasToken = computed(() => auth.tokenExists)
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
    <div class="w-full px-6 h-14 flex items-center justify-between">
      <RouterLink to="/">
        <img :src="logoUrl" alt="风雨世界" class="h-9 w-auto object-contain" />
      </RouterLink>
      <div class="flex items-center gap-2">
        <RouterLink v-if="hasToken" to="/dashboard">
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
