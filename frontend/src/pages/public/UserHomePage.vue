<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { getUserProfile } from '@/api/users'
import { listUserPosts } from '@/api/posts'
import { getMyPins } from '@/api/pins'
import { useAuthStore } from '@/stores/auth'
import type { UserInfo } from '@/types/auth'
import type { PostSummary, PinnedPost } from '@/types/post'
import PostList from '@/components/post/PostList.vue'
import PinCarousel from '@/components/post/PinCarousel.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'

const route = useRoute()
const auth = useAuthStore()
const user = ref<UserInfo | null>(null)
const posts = ref<PostSummary[]>([])
const pins = ref<PinnedPost[]>([])
const loading = ref(true)
const page = ref(1)
const pageSize = 12
const total = ref(0)

const isOwner = computed(() => auth.user?.username === String(route.params.username))

const load = async () => {
  loading.value = true
  const username = String(route.params.username)
  try {
    const requests: Promise<any>[] = [
      getUserProfile(username),
      listUserPosts(username, page.value, pageSize),
    ]
    // 只有本人才能看个人置顶（后端接口是 /me/pins）
    if (isOwner.value) {
      requests.push(getMyPins())
    }
    const [profile, list, pinsRes] = await Promise.all(requests)
    if (profile.code === 0) user.value = profile.data || null
    if (list.code === 0 && list.data) {
      posts.value = list.data.items
      total.value = list.data.total
    }
    if (pinsRes?.code === 0 && pinsRes.data) {
      pins.value = pinsRes.data
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
      <!-- 用户信息头 -->
      <div class="flex items-center gap-5 mb-8 pb-8 border-b border-slate-200">
        <div class="w-16 h-16 rounded-full overflow-hidden bg-slate-100 flex-shrink-0">
          <img
            v-if="user.avatar_url"
            :src="user.avatar_url"
            class="w-full h-full object-cover"
            :alt="user.nickname"
          />
          <div v-else class="w-full h-full flex items-center justify-center text-2xl font-bold text-slate-400">
            {{ user.nickname?.[0]?.toUpperCase() }}
          </div>
        </div>
        <div>
          <h1 class="text-2xl font-bold text-slate-800">{{ user.nickname }}</h1>
          <p class="text-slate-400 text-sm mt-0.5">@{{ user.username }}</p>
          <p v-if="user.bio" class="mt-2 text-slate-500 text-sm leading-relaxed">{{ user.bio }}</p>
        </div>
      </div>

      <!-- 个人置顶轮播（仅本人） -->
      <div v-if="isOwner && pins.length" class="mb-10">
        <div class="flex items-baseline gap-2 mb-3">
          <span class="text-sm font-semibold text-slate-600">我的置顶</span>
          <span class="text-xs text-slate-400">{{ pins.length }} 篇</span>
        </div>
        <PinCarousel :pins="pins" />
      </div>

      <!-- 文章列表 -->
      <div class="flex items-baseline justify-between mb-6">
        <h2 class="text-xl font-bold text-slate-800">
          {{ isOwner ? '我的文章' : `${user.nickname} 的文章` }}
        </h2>
        <span class="text-sm text-slate-400">共 {{ total }} 篇</span>
      </div>

      <PostList :posts="posts" :loading="loading" />
      <Pagination
        v-if="total > pageSize"
        :page="page"
        :page-size="pageSize"
        :total="total"
        class="mt-8"
        @change="(p) => { page = p; load() }"
      />
    </template>
  </div>
</template>
