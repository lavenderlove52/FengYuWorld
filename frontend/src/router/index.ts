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

  const loggedIn = !!getAccessToken()
  if (to.meta.requiresAuth && !loggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.guest && loggedIn) {
    return { name: 'dashboard' }
  }

  // 已登录但用户数据尚未加载（如刷新后），需要获取完整的用户信息
  if (loggedIn && !useAuthStore().user) {
    if (to.meta.requiresAuth) {
      // 受保护路由：必须等待用户数据加载完成后再放行，失败则跳转登录
      // 使用 try/catch 保护：若 access token 过期且无 refresh token，fetchMe 会 reject
      try {
        await useAuthStore().fetchMe()
      } catch {
        const { clearTokens } = await import('@/api/http')
        clearTokens()
        return { name: 'login', query: { redirect: to.fullPath } }
      }
    } else {
      // 公开路由：后台静默加载用户信息，不阻塞导航
      // 避免手机等弱网环境因等待 fetchMe 导致白屏（RouterView 在导航完成前不渲染）
      useAuthStore().fetchMe().catch(() => {})
    }
  }
})

export default router
