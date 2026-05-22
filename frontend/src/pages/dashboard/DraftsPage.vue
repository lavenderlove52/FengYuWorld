<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listMyPosts } from '@/api/posts'
import type { PostSummary } from '@/types/post'
import PostList from '@/components/post/PostList.vue'

const router = useRouter()
const posts = ref<PostSummary[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const res = await listMyPosts({ status: 'draft', page_size: 50 })
    if (res.code === 0 && res.data) posts.value = res.data.items
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">草稿箱</h1>
    <PostList :posts="posts" :loading="loading" />
    <div class="mt-4">
      <button
        v-for="post in posts"
        :key="post.id"
        class="block text-indigo-600 text-sm mb-2"
        @click="router.push(`/dashboard/posts/${post.id}/edit`)"
      >
        编辑：{{ post.title }}
      </button>
    </div>
  </div>
</template>
