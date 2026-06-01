<script setup lang="ts">
import { onMounted, nextTick, ref } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { getPost, getPostBySlug } from '@/api/posts'
import { useUserTheme } from '@/composables/useUserTheme'
import type { PostDetail } from '@/types/post'
import MarkdownViewer from '@/components/post/MarkdownViewer.vue'
import EmptyState from '@/components/common/EmptyState.vue'

const route = useRoute()
const post = ref<PostDetail | null>(null)
const loading = ref(true)
const notFound = ref(false)

const containerRef = ref<HTMLElement | null>(null)
const { load: loadTheme } = useUserTheme(containerRef)

onMounted(async () => {
  const param = String(route.params.id)
  const isNumericId = /^\d+$/.test(param)

  // 第一阶段：加载文章内容，加载结果决定是否显示文章
  try {
    const res = isNumericId
      ? await getPost(Number(param))
      : await getPostBySlug(param)
    if (res.code === 0 && res.data) {
      post.value = res.data
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
})
</script>

<template>
  <div v-if="loading" class="text-center py-12">加载中...</div>
  <EmptyState v-else-if="notFound || !post" message="文章不存在或无权查看" />
  <div
    v-else
    ref="containerRef"
    :data-theme-user="post.author.username"
    class="transition-colors duration-300"
  >
    <article class="rounded-xl border px-4 py-5 sm:px-6 sm:py-7 md:px-8 md:py-10" style="background: var(--theme-card-bg, #fff); border-color: var(--theme-border, #e2e8f0)">
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
      <img v-if="post.cover_image_url" :src="post.cover_image_url" class="w-full rounded-lg mb-6" />
      <MarkdownViewer :html="post.content_html" />
      <div v-if="post.tags?.length" class="mt-8 flex gap-2 flex-wrap">
        <span
          v-for="tag in post.tags"
          :key="tag"
          class="px-2 py-1 rounded text-sm"
          style="background: var(--theme-card-bg, #f1f5f9); color: var(--theme-accent, #6366f1); border: 1px solid var(--theme-border, #e2e8f0)"
        >#{{ tag }}</span>
      </div>
    </article>
  </div>
</template>
