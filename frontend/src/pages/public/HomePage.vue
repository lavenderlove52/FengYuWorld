<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listPublicPosts } from '@/api/posts'
import type { PostSummary } from '@/types/post'
import PostList from '@/components/post/PostList.vue'
import Pagination from '@/components/common/Pagination.vue'

const posts = ref<PostSummary[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = 12
const total = ref(0)

const load = async () => {
  loading.value = true
  try {
    const res = await listPublicPosts(page.value, pageSize)
    if (res.code === 0 && res.data) {
      posts.value = res.data.items
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

onMounted(load)

const onPageChange = (p: number) => {
  page.value = p
  load()
}
</script>

<template>
  <div>
    <h1 class="text-3xl font-bold mb-2">公开文章</h1>
    <p class="text-slate-600 mb-8">全站用户公开发布的博客与日记</p>
    <PostList :posts="posts" :loading="loading" />
    <Pagination v-if="total > pageSize" :page="page" :page-size="pageSize" :total="total" @change="onPageChange" />
  </div>
</template>
