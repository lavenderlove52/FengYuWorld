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
</script>

<template>
  <article class="bg-white rounded-xl shadow-sm border overflow-hidden hover:shadow-md transition">
    <RouterLink :to="`/p/${post.id}`">
      <div v-if="post.cover_image_url" class="h-44 bg-slate-100 overflow-hidden">
        <img :src="post.cover_image_url" :alt="post.title" class="w-full h-full object-cover" />
      </div>
      <div class="p-5">
        <h2 class="text-lg font-semibold mb-2">{{ post.title }}</h2>
        <p class="text-slate-600 text-sm line-clamp-2 mb-3">{{ post.summary }}</p>
        <div class="flex items-center justify-between text-xs text-slate-500">
          <span>{{ post.author.nickname }} · {{ post.published_at || post.created_at }}</span>
          <span>{{ post.view_count }} 阅读</span>
        </div>
        <div v-if="post.status !== 'published'" class="mt-2 text-xs text-amber-600">
          {{ post.status }} / {{ visibilityLabel[post.visibility] || post.visibility }}
        </div>
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
