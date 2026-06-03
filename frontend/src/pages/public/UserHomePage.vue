<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { getUserProfile } from '@/api/users'
import { listUserPosts } from '@/api/posts'
import { getMyPins, getUserPins } from '@/api/pins'
import { useAuthStore } from '@/stores/auth'
import { useUserTheme } from '@/composables/useUserTheme'
import type { PublicUserInfo } from '@/types/auth'
import type { PostSummary, PinnedPost } from '@/types/post'
import PostList from '@/components/post/PostList.vue'
import PinCarousel from '@/components/post/PinCarousel.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'

const route = useRoute()
const auth = useAuthStore()
const user = ref<PublicUserInfo | null>(null)
const posts = ref<PostSummary[]>([])
const pins = ref<PinnedPost[]>([])
const loading = ref(true)
const page = ref(1)
const pageSize = 12
const total = ref(0)

const containerRef = ref<HTMLElement | null>(null)
const { theme, load: loadTheme } = useUserTheme(containerRef)

const isOwner = computed(() => auth.user?.username === String(route.params.username))
const showPins = computed(() => isOwner.value || theme.value.show_pins_public)

const load = async () => {
  loading.value = true
  const username = String(route.params.username)
  try {
    const requests: Promise<any>[] = [
      getUserProfile(username),
      listUserPosts(username, page.value, pageSize),
    ]
    if (isOwner.value || theme.value.show_pins_public) {
      requests.push(isOwner.value ? getMyPins() : getUserPins(username))
    }
    const [profile, list, pinsRes] = await Promise.all(requests)
    if (profile.code === 0) user.value = profile.data || null
    if (list.code === 0 && list.data) {
      posts.value = list.data.items
      total.value = list.data.total
    }
    if (pinsRes?.code === 0 && pinsRes.data) {
      pins.value = pinsRes.data
    }
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const username = String(route.params.username)
  await loadTheme(username)
  await load()
})
</script>

<template>
  <div ref="containerRef" :data-theme-user="String(route.params.username)" data-theme-page="user-home" class="min-h-screen transition-colors duration-300" style="color: var(--theme-text)">
    <div v-if="loading && !user" class="text-center py-20 text-slate-400">加载中...</div>
    <EmptyState v-else-if="!loading && !user" message="用户不存在" />
    <template v-else-if="user">

      <!-- Banner -->
      <div v-if="theme.banner_url" class="theme-hero-banner w-full h-36 sm:h-48 overflow-hidden mb-6 sm:mb-8 rounded-xl -mx-0">
        <img :src="theme.banner_url" class="w-full h-full object-cover" alt="banner" />
      </div>

      <!-- 用户信息头 -->
      <div class="theme-hero-profile flex items-center gap-4 sm:gap-5 mb-6 sm:mb-8 pb-6 sm:pb-8" style="border-bottom: 1px solid var(--theme-border, #e2e8f0)">
        <div class="w-12 h-12 sm:w-16 sm:h-16 rounded-full overflow-hidden flex-shrink-0" style="background: var(--theme-card-bg, #f1f5f9)">
          <img
            v-if="user.avatar_url"
            :src="user.avatar_url"
            class="w-full h-full object-cover"
            :alt="user.nickname"
          />
          <div v-else class="w-full h-full flex items-center justify-center text-xl sm:text-2xl font-bold" style="color: var(--theme-accent, #6366f1)">
            {{ user.nickname?.[0]?.toUpperCase() }}
          </div>
        </div>
        <div>
          <h1 class="text-xl sm:text-2xl font-bold" style="color: var(--theme-text, #0f172a)">{{ user.nickname }}</h1>
          <p class="text-sm mt-0.5" style="color: var(--theme-text, #64748b); opacity: 0.6">@{{ user.username }}</p>
          <p v-if="user.bio" class="mt-2 text-sm leading-relaxed" style="color: var(--theme-text, #64748b); opacity: 0.8">{{ user.bio }}</p>
        </div>
      </div>

      <!-- 首屏图下方的正文区（置顶 + 文章列表） -->
      <div class="theme-main-below">
      <!-- 置顶轮播 -->
      <div v-if="showPins && pins.length" class="mb-10">
        <div class="flex items-baseline gap-2 mb-3">
          <span class="text-sm font-semibold" style="color: var(--theme-text, #475569)">置顶</span>
          <span class="text-xs" style="color: var(--theme-text, #94a3b8); opacity: 0.7">{{ pins.length }} 篇</span>
        </div>
        <PinCarousel :pins="pins" />
      </div>

      <!-- 文章列表 -->
      <div class="flex items-baseline justify-between mb-6">
        <h2 class="text-lg sm:text-xl font-bold" style="color: var(--theme-text, #0f172a)">
          {{ isOwner ? '我的文章' : `${user.nickname} 的文章` }}
        </h2>
        <span class="text-sm" style="color: var(--theme-text, #94a3b8); opacity: 0.7">共 {{ total }} 篇</span>
      </div>

      <PostList :posts="posts" :loading="loading" :layout="theme.layout" />
      <Pagination
        v-if="total > pageSize"
        :page="page"
        :page-size="pageSize"
        :total="total"
        class="mt-8"
        @change="(p) => { page = p; load() }"
      />
      </div>
    </template>
  </div>
</template>
