<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { listPublicPosts } from '@/api/posts'
import { getSitePins } from '@/api/pins'
import { useAuthStore } from '@/stores/auth'
import type { PostSummary, PinnedPost } from '@/types/post'
import PostList from '@/components/post/PostList.vue'
import PinCarousel from '@/components/post/PinCarousel.vue'
import Pagination from '@/components/common/Pagination.vue'

const posts = ref<PostSummary[]>([])
const pins = ref<PinnedPost[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = 12
const total = ref(0)

const auth = useAuthStore()

const load = async () => {
  loading.value = true
  try {
    const [postsRes, pinsRes] = await Promise.all([
      listPublicPosts(page.value, pageSize),
      getSitePins(),
    ])
    if (postsRes.code === 0 && postsRes.data) {
      posts.value = postsRes.data.items
      total.value = postsRes.data.total
    }
    if (pinsRes.code === 0 && pinsRes.data) {
      pins.value = pinsRes.data
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
    <!-- 我的主页入口（已登录） -->
    <div v-if="auth.isLoggedIn && auth.user" class="flex items-center justify-between mb-6">
      <div />
      <RouterLink
        :to="`/u/${auth.user.username}`"
        class="flex items-center gap-2 text-sm text-slate-500 hover:text-slate-800 transition-colors"
      >
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
          <path d="M10 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM3.465 14.493a1.23 1.23 0 0 0 .41 1.412A9.957 9.957 0 0 0 10 18c2.31 0 4.438-.784 6.131-2.1.43-.333.604-.903.408-1.41a7.002 7.002 0 0 0-13.074.003Z" />
        </svg>
        我的主页
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="w-3 h-3 opacity-50">
          <path fill-rule="evenodd" d="M6.22 4.22a.75.75 0 0 1 1.06 0l3.25 3.25a.75.75 0 0 1 0 1.06l-3.25 3.25a.75.75 0 0 1-1.06-1.06L9.19 8 6.22 5.03a.75.75 0 0 1 0-1.06Z" />
        </svg>
      </RouterLink>
    </div>

    <!-- 置顶轮播 -->
    <div v-if="pins.length" class="mb-10">
      <PinCarousel :pins="pins" />
    </div>

    <!-- 文章列表 -->
    <div class="flex items-baseline justify-between mb-6">
      <h1 class="text-xl sm:text-2xl font-bold text-slate-800">公开文章</h1>
      <span class="hidden sm:inline text-sm text-slate-400">全站用户公开发布的博客与日记</span>
    </div>

    <PostList :posts="posts" :loading="loading" />
    <Pagination
      v-if="total > pageSize"
      :page="page"
      :page-size="pageSize"
      :total="total"
      class="mt-8"
      @change="onPageChange"
    />
  </div>
</template>
