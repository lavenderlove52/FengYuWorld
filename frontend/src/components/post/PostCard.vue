<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { PostSummary } from '@/types/post'

defineProps<{ post: PostSummary }>()

const visibilityLabel: Record<string, string> = {
  public: '公开',
  private: '私密',
  friends: '好友',
  specified: '指定好友',
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  return dateStr.slice(0, 10)
}
</script>

<template>
  <article class="py-7 group">
    <RouterLink :to="`/p/${post.id}`" class="block">
      <div v-if="post.cover_image_url" class="h-44 overflow-hidden rounded-lg mb-4 bg-slate-100">
        <img
          :src="post.cover_image_url"
          :alt="post.title"
          class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
        />
      </div>

      <h2 class="text-xl font-bold text-slate-800 leading-snug mb-3 line-clamp-2 group-hover:text-blue-600 transition-colors duration-200">
        {{ post.title }}
      </h2>

      <p class="text-slate-500 text-[15px] leading-relaxed line-clamp-2 mb-4">
        {{ post.summary }}
      </p>

      <div class="flex items-center gap-1.5 text-sm text-slate-400 flex-wrap">
        <span class="text-slate-500 font-medium">{{ post.author.nickname }}</span>
        <span class="text-slate-300">·</span>
        <span>{{ formatDate(post.published_at || post.created_at) }}</span>
        <span class="text-slate-300">·</span>
        <span>{{ post.view_count }} 阅读</span>
        <span v-if="post.status !== 'published'" class="ml-1 text-amber-500">
          {{ visibilityLabel[post.visibility] || post.visibility }}
        </span>
      </div>
    </RouterLink>
  </article>
</template>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
