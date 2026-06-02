import { createRouter, createWebHistory } from 'vue-router'
import { getAccessToken } from '@/api/http'
import { setActivePageTheme } from '@/stores/pageTheme'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/PublicLayout.vue'),
      children: [
        { path: '', name: 'home', component: () => import('@/pages/public/HomePage.vue') },
        { path: 'p/:id', name: 'post-detail', component: () => import('@/pages/public/PostDetailPage.vue') },
        { path: 'u/:username', name: 'user-home', component: () => import('@/pages/public/UserHomePage.vue') },
      ],
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/auth/LoginPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/pages/auth/RegisterPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('@/pages/auth/ForgotPasswordPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/dashboard',
      component: () => import('@/layouts/DashboardLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', name: 'dashboard', component: () => import('@/pages/dashboard/DashboardHomePage.vue') },
        { path: 'posts', name: 'my-posts', component: () => import('@/pages/dashboard/MyPostsPage.vue') },
        { path: 'posts/new', name: 'post-new', component: () => import('@/pages/dashboard/PostEditorPage.vue') },
        { path: 'posts/:id/edit', name: 'post-edit', component: () => import('@/pages/dashboard/PostEditorPage.vue') },
        { path: 'drafts', name: 'drafts', component: () => import('@/pages/dashboard/DraftsPage.vue') },
        { path: 'friends', name: 'friends', component: () => import('@/pages/dashboard/FriendsPage.vue') },
        { path: 'pins', name: 'pin-manage', component: () => import('@/pages/dashboard/PinManagePage.vue') },
        { path: 'theme', name: 'theme-settings', component: () => import('@/pages/dashboard/ThemeSettingsPage.vue') },
        { path: 'admin/pins', name: 'admin-pins', component: () => import('@/pages/dashboard/AdminPinPage.vue') },
        { path: 'admin/invite-codes', name: 'admin-invite-codes', component: () => import('@/pages/dashboard/InviteCodePage.vue') },
        { path: 'settings', name: 'settings', component: () => import('@/pages/dashboard/SettingsPage.vue') },
      ],
    },
  ],
})

router.beforeEach(async (to, from) => {
  // 离开用户主页或文章页时立即重置主题，不等组件卸载
  const themeRoutes = ['user-home', 'post-detail']
  if (themeRoutes.includes(String(from.name)) && !themeRoutes.includes(String(to.name))) {
    setActivePageTheme(null)
    document.body.style.backgroundColor = ''
    document.body.style.color = ''
    document.body.style.fontFamily = ''
  }

  const auth = useAuthStore()
  const hasToken = !!getAccessToken()

  if (to.meta.requiresAuth && !hasToken) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  // 仅 token + user 均有效时重定向，避免过期 token 导致无法打开登录页（手机常见）
  if (to.meta.guest && auth.isLoggedIn) {
    return { name: 'dashboard' }
  }

  // 有 token 但 user 未加载（刷新/冷启动）：后台拉取，绝不 await，避免弱网下整页白屏
  if (hasToken && !auth.user) {
    auth.fetchMe().then(() => {
      if (auth.user && to.meta.guest) {
        router.replace({ name: 'dashboard' })
      }
    }).catch(() => {})
  }
})

export default router
