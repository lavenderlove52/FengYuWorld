<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { NButton } from 'naive-ui'
import { getAccessToken } from '@/api/http'
import { appBase } from '@/config/app'

const loggedIn = !!getAccessToken()
const logoUrl = `${appBase}logo.png`
</script>

<template>
  <header class="bg-[#f8fafc] sticky top-0 z-10">
    <div class="w-full px-6 h-14 flex items-center justify-between">
      <RouterLink to="/">
        <img :src="logoUrl" alt="风雨世界" class="h-9 w-auto object-contain" />
      </RouterLink>
      <div class="flex items-center gap-2">
        <RouterLink v-if="loggedIn" to="/dashboard">
          <button class="w-9 h-9 rounded-full border border-gray-300 bg-white flex items-center justify-center hover:bg-gray-100 transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#333" width="20" height="20">
              <path d="M12 12c2.7 0 4.8-2.1 4.8-4.8S14.7 2.4 12 2.4 7.2 4.5 7.2 7.2 9.3 12 12 12zm0 2.4c-3.2 0-9.6 1.6-9.6 4.8v2.4h19.2v-2.4c0-3.2-6.4-4.8-9.6-4.8z"/>
            </svg>
          </button>
        </RouterLink>
        <template v-else>
          <RouterLink to="/login"><NButton size="small">登录</NButton></RouterLink>
          <RouterLink to="/register"><NButton type="primary" size="small">注册</NButton></RouterLink>
        </template>
      </div>
    </div>
  </header>
</template>
