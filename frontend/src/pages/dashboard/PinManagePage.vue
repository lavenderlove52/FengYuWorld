<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NButton, useMessage } from 'naive-ui'
import { listMyPosts } from '@/api/posts'
import { getMyPins, addUserPin, removeUserPin } from '@/api/pins'
import type { PostSummary, PinnedPost } from '@/types/post'

const message = useMessage()

const posts = ref<PostSummary[]>([])
const pins = ref<PinnedPost[]>([])
const loadingPosts = ref(false)
const pinning = ref<number | null>(null)

const pinnedPostIds = () => new Set(pins.value.map((p) => p.post_id))

const load = async () => {
  loadingPosts.value = true
  try {
    const [postsRes, pinsRes] = await Promise.all([
      listMyPosts({ status: 'published', visibility: 'public', page: 1, page_size: 50 }),
      getMyPins(),
    ])
    if (postsRes.code === 0 && postsRes.data) posts.value = postsRes.data.items
    if (pinsRes.code === 0 && pinsRes.data) pins.value = pinsRes.data
  } finally {
    loadingPosts.value = false
  }
}

onMounted(load)

const togglePin = async (post: PostSummary) => {
  const ids = pinnedPostIds()
  pinning.value = post.id
  try {
    if (ids.has(post.id)) {
      const res = await removeUserPin(post.id)
      if (res.code === 0) {
        pins.value = pins.value.filter((p) => p.post_id !== post.id)
        message.success('已取消置顶')
      } else {
        message.error(res.message || '操作失败')
      }
    } else {
      const res = await addUserPin(post.id)
      if (res.code === 0) {
        await load()
        message.success('已置顶')
      } else {
        message.error(res.message || '操作失败')
      }
    }
  } finally {
    pinning.value = null
  }
}

const formatDate = (d: string) => d?.slice(0, 10) ?? ''
</script>

<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold">置顶管理</h1>
      <p class="text-slate-500 text-sm mt-1">置顶的文章会显示在你的个人主页顶部轮播</p>
    </div>

    <!-- 当前置顶 -->
    <div v-if="pins.length" class="mb-8">
      <h2 class="text-base font-semibold text-slate-700 mb-3">当前置顶（{{ pins.length }} 篇）</h2>
      <div class="space-y-2">
        <div
          v-for="pin in pins"
          :key="pin.id"
          class="flex items-center justify-between px-4 py-3 bg-indigo-50 border border-indigo-100 rounded-lg"
        >
          <div class="flex items-center gap-3">
            <span class="text-indigo-400 font-bold text-sm w-4">{{ pin.display_order }}</span>
            <span class="font-medium text-slate-800 text-sm">{{ pin.post.title }}</span>
            <span class="text-xs text-slate-400">{{ formatDate(pin.post.published_at || pin.post.created_at) }}</span>
          </div>
          <NButton
            size="tiny"
            type="error"
            tertiary
            :loading="pinning === pin.post_id"
            @click="togglePin(pin.post)"
          >
            取消置顶
          </NButton>
        </div>
      </div>
    </div>

    <!-- 所有已发布公开文章 -->
    <div>
      <h2 class="text-base font-semibold text-slate-700 mb-3">已发布的公开文章</h2>
      <div v-if="loadingPosts" class="text-center py-8 text-slate-400">加载中...</div>
      <div v-else-if="!posts.length" class="text-center py-8 text-slate-400">暂无已发布的公开文章</div>
      <div v-else class="space-y-2">
        <div
          v-for="post in posts"
          :key="post.id"
          class="flex items-center justify-between px-4 py-3 bg-white border border-slate-100 rounded-lg hover:border-slate-200 transition"
        >
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-10 h-10 rounded overflow-hidden bg-slate-100 flex-shrink-0">
              <img
                v-if="post.cover_image_url"
                :src="post.cover_image_url"
                class="w-full h-full object-cover"
              />
              <div v-else class="w-full h-full flex items-center justify-center text-slate-300 text-xl">📄</div>
            </div>
            <div class="min-w-0">
              <div class="font-medium text-slate-800 text-sm truncate">{{ post.title }}</div>
              <div class="text-xs text-slate-400 mt-0.5">{{ formatDate(post.published_at || post.created_at) }} · {{ post.view_count }} 阅读</div>
            </div>
          </div>
          <NButton
            :type="pinnedPostIds().has(post.id) ? 'default' : 'primary'"
            :tertiary="pinnedPostIds().has(post.id)"
            size="small"
            :loading="pinning === post.id"
            @click="togglePin(post)"
          >
            {{ pinnedPostIds().has(post.id) ? '已置顶' : '置顶' }}
          </NButton>
        </div>
      </div>
    </div>
  </div>
</template>
