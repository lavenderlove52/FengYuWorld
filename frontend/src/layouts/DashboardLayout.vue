<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { NMenu, NButton, NSpace, NSpin } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'
import { useInactivityLogout } from '@/composables/useInactivityLogout'
import { setActivePageTheme } from '@/stores/pageTheme'
import { clearTokens } from '@/api/http'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const settingsStore = useSettingsStore()
useInactivityLogout()

/** 有 token 但尚未拿到 user（刷新后首屏） */
const authBooting = computed(() => auth.tokenExists && !auth.user)

const ensureUser = async () => {
  if (!auth.tokenExists || auth.user) return
  try {
    await auth.fetchMe()
  } catch {
    // fetchMe 内部请求失败时由拦截器处理 token
  }
  if (!auth.user) {
    clearTokens()
    auth.tokenExists = false
    router.replace({ name: 'login', query: { redirect: route.fullPath, reason: 'session_expired' } })
  }
}

watch(authBooting, (booting) => {
  if (booting) ensureUser()
})

const sidebarOpen = ref(false)

const menuOptions = computed(() => {
  const items = [
    { label: () => h(RouterLink, { to: '/dashboard' }, { default: () => '概览' }), key: 'dashboard' },
    { label: () => h(RouterLink, { to: '/dashboard/posts/new' }, { default: () => '写文章' }), key: 'post-new' },
    { label: () => h(RouterLink, { to: '/dashboard/posts' }, { default: () => '我的文章' }), key: 'my-posts' },
    { label: () => h(RouterLink, { to: '/dashboard/post-tree' }, { default: () => '文章目录' }), key: 'post-tree' },
    { label: () => h(RouterLink, { to: '/dashboard/drafts' }, { default: () => '草稿箱' }), key: 'drafts' },
    { label: () => h(RouterLink, { to: '/dashboard/pins' }, { default: () => '置顶管理' }), key: 'pin-manage' },
    { label: () => h(RouterLink, { to: '/dashboard/theme' }, { default: () => '主页定制' }), key: 'theme-settings' },
    { label: () => h(RouterLink, { to: '/dashboard/friends' }, { default: () => '好友' }), key: 'friends' },
    { label: () => h(RouterLink, { to: '/dashboard/settings' }, { default: () => '设置' }), key: 'settings' },
  ]
  const isAdmin = auth.user?.role === 'admin' || auth.user?.role === 'super_admin' || auth.user?.can_site_pin
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

// 路由变化时关闭移动端侧边栏
watch(() => route.path, () => {
  sidebarOpen.value = false
})

// 进入 dashboard 时强制清除个人主页可能遗留的深色主题
onMounted(() => {
  setActivePageTheme(null)
  document.body.style.backgroundColor = ''
  document.body.style.color = ''
  document.body.style.fontFamily = ''
  settingsStore.loadPublicSettings()
  if (authBooting.value) ensureUser()
})
</script>

<template>
  <!-- 鉴权加载中：立即显示骨架，避免手机弱网长时间白屏 -->
  <div v-if="authBooting" class="min-h-screen flex flex-col items-center justify-center bg-[#f8fafc] gap-3">
    <NSpin size="large" />
    <p class="text-sm text-slate-500">正在加载用户中心…</p>
  </div>

  <div v-else class="min-h-screen flex bg-[#f8fafc]">

    <!-- 移动端遮罩层 -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 bg-black/40 z-40 md:hidden"
      @click="sidebarOpen = false"
    />

    <!-- 侧边栏：移动端固定定位，桌面端静态 -->
    <aside
      class="fixed md:static inset-y-0 left-0 z-50 w-[220px] flex-shrink-0 bg-white border-r border-slate-200 flex flex-col transition-transform duration-300 md:translate-x-0"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <div class="p-4 border-b border-slate-200 flex items-center justify-between">
        <RouterLink to="/" class="font-bold text-lg text-indigo-600">首页</RouterLink>
        <!-- 移动端关闭按钮 -->
        <button
          class="md:hidden w-8 h-8 flex items-center justify-center text-slate-400 hover:text-slate-700 rounded"
          @click="sidebarOpen = false"
        >
          ✕
        </button>
      </div>
      <div class="flex-1 overflow-y-auto">
        <NMenu :value="activeKey" :options="menuOptions" />
      </div>
    </aside>

    <!-- 主内容区 -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- 顶部栏 -->
      <div class="flex items-center justify-between px-4 md:px-6 py-3 border-b bg-white sticky top-0 z-30">
        <div class="flex items-center gap-3">
          <!-- 移动端汉堡菜单按钮 -->
          <button
            class="md:hidden w-11 h-11 flex items-center justify-center text-slate-500 hover:text-slate-800 hover:bg-slate-100 rounded-lg transition-colors"
            aria-label="打开菜单"
            @click="sidebarOpen = true"
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.8" stroke="currentColor" class="w-6 h-6">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
            </svg>
          </button>
          <div class="font-medium text-slate-700">用户中心</div>
        </div>
        <NSpace align="center">
          <RouterLink
            to="/article-directory"
            class="w-9 h-9 rounded-full flex items-center justify-center text-slate-400 hover:text-indigo-500 hover:bg-slate-100 transition-colors"
            title="我的文章"
          >
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
              <path d="M4.75 3A1.75 1.75 0 0 0 3 4.75v10.5C3 16.216 3.784 17 4.75 17h10.5A1.75 1.75 0 0 0 17 15.25V4.75A1.75 1.75 0 0 0 15.25 3H4.75Zm.75 2.5h4.25v1.25H5.5V5.5Zm0 3.25h9v1.25h-9V8.75Zm0 3.25h9v1.25h-9V12Z" />
            </svg>
          </RouterLink>
          <RouterLink
            v-if="auth.user?.username"
            :to="`/u/${auth.user.username}`"
            class="w-9 h-9 rounded-full flex items-center justify-center text-slate-400 hover:text-indigo-500 hover:bg-slate-100 transition-colors"
            title="我的主页"
          >
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
              <path d="M10 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM3.465 14.493a1.23 1.23 0 0 0 .41 1.412A9.957 9.957 0 0 0 10 18c2.31 0 4.438-.784 6.131-2.1.43-.333.604-.903.408-1.41a7.002 7.002 0 0 0-13.074.003Z" />
            </svg>
          </RouterLink>
          <div class="w-8 h-8 rounded-full overflow-hidden bg-slate-200 flex-shrink-0 flex items-center justify-center text-sm font-bold text-slate-500">
            <img
              v-if="auth.user?.avatar_url"
              :src="auth.user.avatar_url"
              :alt="auth.user?.nickname"
              class="w-full h-full object-cover"
            />
            <span v-else>{{ auth.user?.nickname?.[0]?.toUpperCase() || 'U' }}</span>
          </div>
          <span class="hidden sm:inline text-sm text-slate-600">{{ auth.user?.nickname }}</span>
          <NButton size="small" @click="logout">退出</NButton>
        </NSpace>
      </div>

      <!-- 页面内容 -->
      <div class="flex-1 p-4 md:p-6">
        <RouterView />
      </div>
    </div>
  </div>
</template>
