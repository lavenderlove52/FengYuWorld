import { createRouter, createWebHistory } from 'vue-router'
import { getAccessToken } from '@/api/http'

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
        { path: 'admin/pins', name: 'admin-pins', component: () => import('@/pages/dashboard/AdminPinPage.vue') },
        { path: 'settings', name: 'settings', component: () => import('@/pages/dashboard/SettingsPage.vue') },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const loggedIn = !!getAccessToken()
  if (to.meta.requiresAuth && !loggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.guest && loggedIn) {
    return { name: 'dashboard' }
  }
})

export default router
