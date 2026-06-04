<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import type { PostSummary } from '@/types/post'
import type { ThemeLayout } from '@/types/theme'
import { getPostPath } from '@/utils/postPath'
import PostCard from './PostCard.vue'
import ResponsivePostImage from './ResponsivePostImage.vue'
import EmptyState from '@/components/common/EmptyState.vue'

const props = defineProps<{ posts: PostSummary[]; loading?: boolean; layout?: ThemeLayout }>()

const currentLayout = computed(() => props.layout || 'grid')

const rows = computed(() => {
  const result: PostSummary[][] = []
  for (let i = 0; i < props.posts.length; i += 4) {
    result.push(props.posts.slice(i, i + 4))
  }
  return result
})
</script>

<template>
  <div v-if="loading" class="text-center py-12 text-slate-500">加载中...</div>
  <EmptyState v-else-if="!posts.length" />

  <!-- grid 布局：移动 1 列 → 平板 2 列 → 桌面 4 列 -->
  <div v-else-if="currentLayout === 'grid'">
    <template v-for="(row, rowIndex) in rows" :key="rowIndex">
      <hr v-if="rowIndex > 0" class="border-0 border-t" style="border-color: var(--theme-border, #e2e8f0)" />
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-x-8 lg:gap-x-10">
        <PostCard v-for="post in row" :key="post.id" :post="post" />
      </div>
    </template>
  </div>

  <!-- list 布局：图文横排 -->
  <div v-else-if="currentLayout === 'list'" class="space-y-0">
    <div
      v-for="post in posts"
      :key="post.id"
      class="border-b py-5"
      style="border-color: var(--theme-border, #e2e8f0)"
    >
      <RouterLink :to="getPostPath(post)" class="flex gap-4 items-start group">
        <div v-if="post.cover_image_url" class="w-24 h-18 sm:w-32 sm:h-24 rounded-lg overflow-hidden flex-shrink-0 bg-slate-100">
          <ResponsivePostImage
            :src="post.cover_image_url"
            :alt="post.title"
            class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
          />
        </div>
        <div class="flex-1 min-w-0">
          <h2 class="post-list-title text-lg font-bold mb-2 transition-colors" style="color: var(--theme-text, #0f172a)">{{ post.title }}</h2>
          <p v-if="post.summary" class="text-sm leading-relaxed line-clamp-2 mb-3" style="color: var(--theme-text, #64748b); opacity: 0.7">{{ post.summary }}</p>
          <div class="flex items-center gap-2 text-xs" style="color: var(--theme-text, #94a3b8); opacity: 0.6">
            <span>{{ post.author.nickname }}</span>
            <span>·</span>
            <span>{{ (post.published_at || post.created_at)?.slice(0, 10) }}</span>
            <span>·</span>
            <span>{{ post.view_count }} 阅读</span>
          </div>
        </div>
      </RouterLink>
    </div>
  </div>

  <!-- magazine 布局：移动 1 列 → 桌面 2 列 -->
  <div v-else-if="currentLayout === 'magazine'" class="grid grid-cols-1 md:grid-cols-2 gap-6 sm:gap-8">
    <RouterLink
      v-for="post in posts"
      :key="post.id"
      :to="getPostPath(post)"
      class="group border rounded-xl overflow-hidden"
      style="border-color: var(--theme-border, #e2e8f0); background: var(--theme-card-bg, #fff)"
    >
      <div v-if="post.cover_image_url" class="h-44 overflow-hidden">
        <ResponsivePostImage
          :src="post.cover_image_url"
          :alt="post.title"
          class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
        />
      </div>
      <div class="p-5">
        <h2 class="post-list-title text-base font-bold mb-2 transition-colors line-clamp-2" style="color: var(--theme-text, #0f172a)">{{ post.title }}</h2>
        <p v-if="post.summary" class="text-sm line-clamp-2 mb-3" style="color: var(--theme-text, #64748b); opacity: 0.7">{{ post.summary }}</p>
        <div class="text-xs" style="color: var(--theme-text, #94a3b8); opacity: 0.6">
          {{ post.author.nickname }} · {{ (post.published_at || post.created_at)?.slice(0, 10) }} · {{ post.view_count }} 阅读
        </div>
      </div>
    </RouterLink>
  </div>
</template>
