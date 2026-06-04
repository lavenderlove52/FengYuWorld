<script setup lang="ts">
import { computed, onMounted, nextTick, ref, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { getPost, getPostBySlug } from '@/api/posts'
import { useUserTheme } from '@/composables/useUserTheme'
import type { PostDetail } from '@/types/post'
import MarkdownViewer from '@/components/post/MarkdownViewer.vue'
import ArticleToc from '@/components/post/ArticleToc.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ThemeEffects from '@/components/theme/ThemeEffects.vue'
import { extractTocItems } from '@/utils/markdownToc'

const route = useRoute()
const post = ref<PostDetail | null>(null)
const loading = ref(true)
const notFound = ref(false)

const containerRef = ref<HTMLElement | null>(null)
const { theme, load: loadTheme } = useUserTheme(containerRef)
const showToc = computed(() => post.value?.show_toc !== false && extractTocItems(post.value?.content_md || '').length > 0)

async function loadPost(param: string) {
  // 纯数字路径既可能是 ID，也可能是 slug（如 slug=123 的旧链接 /p/123）
  if (/^\d+$/.test(param)) {
    const byId = await getPost(Number(param))
    if (byId.code === 0 && byId.data) return byId.data

    const bySlug = await getPostBySlug(param)
    if (bySlug.code === 0 && bySlug.data) return bySlug.data
    return null
  }

  const bySlug = await getPostBySlug(param)
  if (bySlug.code === 0 && bySlug.data) return bySlug.data
  return null
}

const loadPage = async () => {
  post.value = null
  loading.value = true
  notFound.value = false
  const param = String(route.params.id)

  try {
    const data = await loadPost(param)
    if (data) {
      post.value = data
    } else {
      notFound.value = true
    }
  } catch {
    notFound.value = true
  } finally {
    // 立即结束 loading，让文章内容先渲染出来
    loading.value = false
  }

  // 第二阶段：文章渲染完毕后再加载主题（containerRef 此时已指向真实 DOM）
  // 主题加载失败不影响文章内容的展示
  if (post.value) {
    await nextTick()
    try {
      await loadTheme(post.value.author.username)
    } catch {
      // 主题加载失败属于非致命错误，静默忽略
    }
  }
}

onMounted(loadPage)
watch(() => route.params.id, loadPage)
</script>

<template>
  <div v-if="loading" class="text-center py-12">加载中...</div>
  <EmptyState v-else-if="notFound || !post" message="文章不存在或无权查看" />
  <div
    v-else
    ref="containerRef"
    :data-theme-user="post.author.username"
    data-theme-page="post-detail"
    class="article-detail-root transition-colors duration-300"
  >
    <ThemeEffects :config="theme.effect_config" page-scope="post" />
    <div class="article-detail-layout grid gap-6" :class="showToc ? 'lg:grid-cols-[minmax(0,1fr)_220px]' : ''">
      <article class="article-detail-card rounded-xl border px-4 py-5 sm:px-6 sm:py-7 md:px-8 md:py-10" style="background: var(--theme-card-bg, #fff); border-color: var(--theme-border, #e2e8f0)">
        <h1 class="text-2xl sm:text-3xl font-bold mb-4" style="color: var(--theme-text, #0f172a)">{{ post.title }}</h1>
        <div class="text-sm mb-5 sm:mb-6 flex flex-wrap gap-2 sm:gap-4" style="color: var(--theme-text, #64748b); opacity: 0.7">
          <RouterLink
            :to="`/u/${post.author.username}`"
            class="font-medium hover:opacity-100 transition-opacity"
            style="color: var(--theme-accent, #6366f1)"
          >
            {{ post.author.nickname }}
          </RouterLink>
          <span>{{ (post.published_at || post.created_at)?.slice(0, 10) }}</span>
          <span>{{ post.view_count }} 阅读</span>
        </div>
        <p v-if="post.summary" class="text-base leading-relaxed mb-6" style="color: var(--theme-text, #64748b); opacity: 0.72">
          {{ post.summary }}
        </p>
        <img v-if="post.cover_image_url" :src="post.cover_image_url" class="w-full rounded-lg mb-6" />
        <div v-if="showToc" class="lg:hidden mb-6 rounded-lg border px-4 py-3" style="border-color: var(--theme-border, #e2e8f0)">
          <ArticleToc :content-md="post.content_md || ''" />
        </div>
        <MarkdownViewer :content-md="post.content_md || ''" />
        <div v-if="post.tags?.length" class="mt-8 flex gap-2 flex-wrap">
          <span
            v-for="tag in post.tags"
            :key="tag"
            class="px-2 py-1 rounded text-sm"
            style="background: var(--theme-card-bg, #f1f5f9); color: var(--theme-accent, #6366f1); border: 1px solid var(--theme-border, #e2e8f0)"
          >#{{ tag }}</span>
        </div>
      </article>
      <aside v-if="showToc" class="hidden lg:block">
        <ArticleToc :content-md="post.content_md || ''" />
      </aside>
    </div>
  </div>
</template>

<style scoped>
.article-detail-root.article-detail-root > .article-detail-layout.article-detail-layout {
  width: auto !important;
  max-width: none !important;
  height: auto !important;
  min-height: 0 !important;
  max-height: none !important;
  position: relative !important;
  left: auto !important;
  transform: none !important;
  margin: 0 !important;
  border-radius: 0 !important;
  overflow: visible !important;
}

.article-detail-root.article-detail-root > .article-detail-layout.article-detail-layout::after {
  content: none !important;
  display: none !important;
}

.article-detail-root.article-detail-root .article-detail-card.article-detail-card img {
  width: auto !important;
  max-width: 100% !important;
  height: auto !important;
  min-height: 0 !important;
  max-height: none !important;
  position: static !important;
  left: auto !important;
  transform: none !important;
  border-radius: 0.5rem !important;
  filter: none !important;
}
</style>
