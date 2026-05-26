<script setup lang="ts">
import { computed } from 'vue'
import type { PostSummary } from '@/types/post'
import PostCard from './PostCard.vue'
import EmptyState from '@/components/common/EmptyState.vue'

const props = defineProps<{ posts: PostSummary[]; loading?: boolean }>()

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
  <div v-else>
    <template v-for="(row, rowIndex) in rows" :key="rowIndex">
      <hr v-if="rowIndex > 0" class="border-0 border-t border-slate-200" />
      <div class="grid grid-cols-4 gap-x-10">
        <PostCard v-for="post in row" :key="post.id" :post="post" />
      </div>
    </template>
  </div>
</template>
