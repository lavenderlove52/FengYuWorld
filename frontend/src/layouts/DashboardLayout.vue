<script setup lang="ts">
import { computed, h } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { NLayout, NLayoutSider, NLayoutContent, NMenu, NButton, NSpace } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { useInactivityLogout } from '@/composables/useInactivityLogout'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
useInactivityLogout()

const menuOptions = computed(() => {
  const items = [
    { label: () => h(RouterLink, { to: '/dashboard' }, { default: () => '概览' }), key: 'dashboard' },
    { label: () => h(RouterLink, { to: '/dashboard/posts/new' }, { default: () => '写文章' }), key: 'post-new' },
    { label: () => h(RouterLink, { to: '/dashboard/posts' }, { default: () => '我的文章' }), key: 'my-posts' },
    { label: () => h(RouterLink, { to: '/dashboard/drafts' }, { default: () => '草稿箱' }), key: 'drafts' },
    { label: () => h(RouterLink, { to: '/dashboard/pins' }, { default: () => '置顶管理' }), key: 'pin-manage' },
    { label: () => h(RouterLink, { to: '/dashboard/theme' }, { default: () => '主页定制' }), key: 'theme-settings' },
    { label: () => h(RouterLink, { to: '/dashboard/friends' }, { default: () => '好友' }), key: 'friends' },
    { label: () => h(RouterLink, { to: '/dashboard/settings' }, { default: () => '设置' }), key: 'settings' },
  ]
  const isAdmin = auth.user?.role === 'admin' || auth.user?.can_site_pin
  if (isAdmin) {
    items.push({
      label: () => h(RouterLink, { to: '/dashboard/admin/pins' }, { default: () => '🔒 首页置顶' }),
      key: 'admin-pins',
    })
    items.push({
      label: () => h(RouterLink, { to: '/dashboard/admin/invite-codes' }, { default: () => '🔒 邀请码' }),
      key: 'admin-invite-codes',
    })
  }
  return items
})

const activeKey = computed(() => {
  if (route.name === 'post-edit' || route.name === 'post-new') return 'post-new'
  return String(route.name || 'dashboard')
})

const logout = async () => {
  await auth.logout()
  router.push('/login')
}
</script>

<template>
  <NLayout has-sider class="min-h-screen">
    <NLayoutSider bordered width="220" class="bg-white">
      <div class="p-4 border-b">
        <RouterLink to="/" class="font-bold text-lg text-indigo-600">首页</RouterLink>
      </div>
      <NMenu :value="activeKey" :options="menuOptions" />    </NLayoutSider>
    <NLayout>
      <div class="flex items-center justify-between px-6 py-3 border-b bg-white">
        <div class="font-medium">用户中心</div>
        <NSpace align="center">
          <div class="w-8 h-8 rounded-full overflow-hidden bg-slate-200 flex-shrink-0 flex items-center justify-center text-sm font-bold text-slate-500">
            <img
              v-if="auth.user?.avatar_url"
              :src="auth.user.avatar_url"
              :alt="auth.user?.nickname"
              class="w-full h-full object-cover"
            />
            <span v-else>{{ auth.user?.nickname?.[0]?.toUpperCase() || 'U' }}</span>
          </div>
          <span>{{ auth.user?.nickname }}</span>
          <NButton size="small" @click="logout">退出</NButton>
        </NSpace>
      </div>
      <NLayoutContent class="p-6">
        <RouterView />
      </NLayoutContent>
    </NLayout>
  </NLayout>
</template>
