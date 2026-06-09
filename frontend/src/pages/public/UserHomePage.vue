<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getUserProfile } from '@/api/users'
import { listUserPosts } from '@/api/posts'
import { getMyPins, getUserPins } from '@/api/pins'
import { useAuthStore } from '@/stores/auth'
import { useUserTheme } from '@/composables/useUserTheme'
import type { PublicUserInfo } from '@/types/auth'
import type { PostSummary, PinnedPost } from '@/types/post'
import type { ThemeImage } from '@/types/theme'
import PostList from '@/components/post/PostList.vue'
import PinCarousel from '@/components/post/PinCarousel.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ThemeEffects from '@/components/theme/ThemeEffects.vue'
import ThemeHomeHero from '@/components/theme/ThemeHomeHero.vue'

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
const contentRef = ref<HTMLElement | null>(null)
const { theme, load: loadTheme } = useUserTheme(containerRef)

const isOwner = computed(() => auth.user?.username === String(route.params.username))
const showPins = computed(() => isOwner.value || theme.value.show_pins_public)
const heroImages = computed(() => {
  const records = theme.value.hero_images ?? []
  const byId = new Map(records.map((image) => [image.id, image]))
  return theme.value.hero_config.image_ids
    .slice(0, 5)
    .map((id) => byId.get(id))
    .filter((image): image is ThemeImage => Boolean(image?.url))
})
const showFullscreenHero = computed(() => theme.value.hero_config.enabled && heroImages.value.length > 0)
const articleStyle = computed(() => ({
  maxWidth: `${Math.min(1800, Math.max(600, theme.value.hero_config.page_width_px))}px`,
}))
const articleLayout = computed(() => showFullscreenHero.value ? theme.value.hero_config.article_layout : theme.value.layout)

const scrollToContent = () => {
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  contentRef.value?.scrollIntoView({ behavior: reduceMotion ? 'auto' : 'smooth', block: 'start' })
}

const load = async () => {
  loading.value = true
  const username = String(route.params.username)
  try {
    const pinsRequest = isOwner.value || theme.value.show_pins_public
      ? (isOwner.value ? getMyPins() : getUserPins(username))
          .then((res) => {
            if (res.code === 0 && res.data) {
              pins.value = res.data
            }
          })
          .catch(() => {})
      : null

    void pinsRequest
    const requests: Promise<any>[] = [getUserProfile(username), listUserPosts(username, page.value, pageSize)]
    const [profileResult, listResult] = await Promise.allSettled(requests)
    const profile = profileResult.status === 'fulfilled' ? profileResult.value : null
    const list = listResult.status === 'fulfilled' ? listResult.value : null
    if (profile?.code === 0) user.value = profile.data || null
    if (list?.code === 0 && list.data) {
      posts.value = list.data.items
      total.value = list.data.total
    }
  } finally {
    loading.value = false
  }
}

const loadPage = async () => {
  user.value = null
  posts.value = []
  pins.value = []
  total.value = 0
  page.value = 1
  const username = String(route.params.username)
  await loadTheme(username)
  await load()
}

onMounted(loadPage)
watch(() => route.params.username, loadPage)
</script>

<template>
  <div
    ref="containerRef"
    :data-theme-user="String(route.params.username)"
    data-theme-page="user-home"
    class="relative isolate min-h-screen"
    style="color: var(--theme-text)"
  >
    <ThemeEffects :config="theme.effect_config" page-scope="home" />
    <div v-if="loading && !user" class="relative z-[1] text-center py-20 text-slate-400">加载中...</div>
    <EmptyState v-else-if="!loading && !user" class="relative z-[1]" message="用户不存在" />
    <div v-else-if="user" class="relative z-[1]">
      <ThemeHomeHero
        v-if="showFullscreenHero"
        :config="theme.hero_config"
        :images="heroImages"
        :user="user"
        @scroll-down="scrollToContent"
      />

      <div class="home-content-shell">
        <div ref="contentRef" class="home-content-inner">
          <div
            v-if="!showFullscreenHero && theme.banner_url"
            class="theme-hero-banner w-full h-36 sm:h-48 overflow-hidden mb-6 sm:mb-8 rounded-xl"
          >
            <img :src="theme.banner_url" class="w-full h-full object-cover" alt="banner" />
          </div>

          <div
            v-if="!showFullscreenHero"
            class="theme-hero-profile flex items-center gap-4 sm:gap-5 mb-6 sm:mb-8 pb-6 sm:pb-8"
            style="border-bottom: 1px solid var(--theme-border, #e2e8f0)"
          >
            <div
              class="w-12 h-12 sm:w-16 sm:h-16 rounded-full overflow-hidden flex-shrink-0"
              style="background: var(--theme-card-bg, #f1f5f9)"
            >
              <img
                v-if="user.avatar_url"
                :src="user.avatar_url"
                class="w-full h-full object-cover"
                :alt="user.nickname"
              />
              <div
                v-else
                class="theme-color-accent w-full h-full flex items-center justify-center text-xl sm:text-2xl font-bold"
                style="color: var(--theme-accent, #6366f1)"
              >
                {{ user.nickname?.[0]?.toUpperCase() }}
              </div>
            </div>
            <div>
              <h1 class="text-xl sm:text-2xl font-bold" style="color: var(--theme-text, #0f172a)">
                {{ user.nickname }}
              </h1>
              <p class="text-sm mt-0.5" style="color: var(--theme-text, #64748b); opacity: 0.6">
                @{{ user.username }}
              </p>
              <p
                v-if="user.bio"
                class="mt-2 text-sm leading-relaxed"
                style="color: var(--theme-text, #64748b); opacity: 0.8"
              >
                {{ user.bio }}
              </p>
            </div>
          </div>

          <div class="theme-main-below" :style="articleStyle">
            <div v-if="showPins && pins.length" class="mb-10">
              <div class="flex items-baseline gap-2 mb-3">
                <span class="text-sm font-semibold" style="color: var(--theme-text, #475569)">置顶</span>
                <span class="text-xs" style="color: var(--theme-text, #94a3b8); opacity: 0.7">
                  {{ pins.length }} 篇
                </span>
              </div>
              <PinCarousel :pins="pins" />
            </div>

            <div class="flex items-baseline justify-between mb-6">
              <h2 class="text-lg sm:text-xl font-bold" style="color: var(--theme-text, #0f172a)">
                {{ isOwner ? '我的文章' : `${user.nickname} 的文章` }}
              </h2>
              <span class="text-sm" style="color: var(--theme-text, #94a3b8); opacity: 0.7">共 {{ total }} 篇</span>
            </div>

            <PostList :posts="posts" :loading="loading" :layout="articleLayout" />
            <Pagination
              v-if="total > pageSize"
              :page="page"
              :page-size="pageSize"
              :total="total"
              class="mt-8"
              @change="(p) => { page = p; load() }"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-content-shell {
  width: 100%;
  background: var(--theme-bg);
}

@media (min-width: 640px) {
  .home-content-shell {
    width: 100vw;
    margin-left: calc(50% - 50vw);
  }
}

.home-content-inner {
  width: calc(100% - 2rem);
  max-width: 1800px;
  margin: 0 auto;
  padding: 2rem 0 3rem;
  scroll-margin-top: 3.5rem;
  content-visibility: auto;
}

.theme-main-below {
  width: 100%;
  margin: 0 auto;
}

@media (min-width: 640px) {
  .home-content-inner {
    width: calc(100% - 3rem);
    padding-top: 2.5rem;
  }
}
</style>
