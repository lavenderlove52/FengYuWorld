<script setup lang="ts">
import type { PostSummary } from '@/types/post'
import PostCard from './PostCard.vue'
import EmptyState from '@/components/common/EmptyState.vue'

defineProps<{ posts: PostSummary[]; loading?: boolean }>()
</script>

<template>
  <div v-if="loading" class="text-center py-12 text-slate-500">加载中...</div>
  <EmptyState v-else-if="!posts.length" />
  <div v-else class="grid gap-6 md:grid-cols-2">
    <PostCard v-for="post in posts" :key="post.id" :post="post" />
  </div>
</template>
