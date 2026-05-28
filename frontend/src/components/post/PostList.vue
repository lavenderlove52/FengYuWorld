<script setup lang="ts">
import { computed } from 'vue'
import type { PostSummary } from '@/types/post'
import type { ThemeLayout } from '@/types/theme'
import PostCard from './PostCard.vue'
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

  <!-- grid 布局：4列 -->
  <div v-else-if="currentLayout === 'grid'">
    <template v-for="(row, rowIndex) in rows" :key="rowIndex">
      <hr v-if="rowIndex > 0" class="border-0 border-t" style="border-color: var(--theme-border, #e2e8f0)" />
      <div class="grid grid-cols-4 gap-x-10">
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
      <a :href="`/p/${post.id}`" class="flex gap-5 items-start group">
        <div v-if="post.cover_image_url" class="w-32 h-24 rounded-lg overflow-hidden flex-shrink-0 bg-slate-100">
          <img :src="post.cover_image_url" :alt="post.title" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
        </div>
        <div class="flex-1 min-w-0">
          <h2 class="text-lg font-bold mb-2 group-hover:opacity-70 transition-opacity" style="color: var(--theme-text, #0f172a)">{{ post.title }}</h2>
          <p class="text-sm leading-relaxed line-clamp-2 mb-3" style="color: var(--theme-text, #64748b); opacity: 0.7">{{ post.summary }}</p>
          <div class="flex items-center gap-2 text-xs" style="color: var(--theme-text, #94a3b8); opacity: 0.6">
            <span>{{ post.author.nickname }}</span>
            <span>·</span>
            <span>{{ (post.published_at || post.created_at)?.slice(0, 10) }}</span>
            <span>·</span>
            <span>{{ post.view_count }} 阅读</span>
          </div>
        </div>
      </a>
    </div>
  </div>

  <!-- magazine 布局：杂志双栏 -->
  <div v-else-if="currentLayout === 'magazine'" class="grid grid-cols-2 gap-8">
    <a
      v-for="post in posts"
      :key="post.id"
      :href="`/p/${post.id}`"
      class="group border rounded-xl overflow-hidden"
      style="border-color: var(--theme-border, #e2e8f0); background: var(--theme-card-bg, #fff)"
    >
      <div v-if="post.cover_image_url" class="h-44 overflow-hidden">
        <img :src="post.cover_image_url" :alt="post.title" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
      </div>
      <div class="p-5">
        <h2 class="text-base font-bold mb-2 group-hover:opacity-70 transition-opacity line-clamp-2" style="color: var(--theme-text, #0f172a)">{{ post.title }}</h2>
        <p class="text-sm line-clamp-2 mb-3" style="color: var(--theme-text, #64748b); opacity: 0.7">{{ post.summary }}</p>
        <div class="text-xs" style="color: var(--theme-text, #94a3b8); opacity: 0.6">
          {{ post.author.nickname }} · {{ (post.published_at || post.created_at)?.slice(0, 10) }} · {{ post.view_count }} 阅读
        </div>
      </div>
    </a>
  </div>
</template>
