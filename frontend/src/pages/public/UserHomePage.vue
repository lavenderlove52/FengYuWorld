<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getUserProfile } from '@/api/users'
import { listUserPosts } from '@/api/posts'
import type { UserInfo } from '@/types/auth'
import type { PostSummary } from '@/types/post'
import PostList from '@/components/post/PostList.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'

const route = useRoute()
const user = ref<UserInfo | null>(null)
const posts = ref<PostSummary[]>([])
const loading = ref(true)
const page = ref(1)
const total = ref(0)

const load = async () => {
  loading.value = true
  const username = String(route.params.username)
  try {
    const [profile, list] = await Promise.all([
      getUserProfile(username),
      listUserPosts(username, page.value),
    ])
    if (profile.code === 0) user.value = profile.data || null
    if (list.code === 0 && list.data) {
      posts.value = list.data.items
      total.value = list.data.total
    }
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <EmptyState v-if="!loading && !user" message="用户不存在" />
    <template v-else-if="user">
      <div class="flex items-center gap-4 mb-8">
        <img
          v-if="user.avatar_url"
          :src="user.avatar_url"
          class="w-16 h-16 rounded-full object-cover border"
        />
        <div>
          <h1 class="text-2xl font-bold">{{ user.nickname }}</h1>
          <p class="text-slate-500">@{{ user.username }}</p>
          <p v-if="user.bio" class="mt-2 text-slate-600">{{ user.bio }}</p>
        </div>
      </div>
      <PostList :posts="posts" :loading="loading" />
      <Pagination v-if="total > 12" :page="page" :page-size="12" :total="total" @change="(p) => { page = p; load() }" />
    </template>
  </div>
</template>
