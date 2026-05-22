<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NSelect, NButton, NSpace, useDialog, useMessage } from 'naive-ui'
import { listMyPosts, deletePost } from '@/api/posts'
import type { PostSummary } from '@/types/post'
import PostList from '@/components/post/PostList.vue'
import Pagination from '@/components/common/Pagination.vue'

const router = useRouter()
const dialog = useDialog()
const message = useMessage()

const posts = ref<PostSummary[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)
const status = ref<string | null>(null)
const visibility = ref<string | null>(null)

const load = async () => {
  loading.value = true
  try {
    const res = await listMyPosts({
      page: page.value,
      page_size: 12,
      status: status.value || undefined,
      visibility: visibility.value || undefined,
    })
    if (res.code === 0 && res.data) {
      posts.value = res.data.items
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

onMounted(load)

const remove = (id: number) => {
  dialog.warning({
    title: '确认删除',
    content: '删除后文章将不可见，确定继续？',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await deletePost(id)
      message.success('已删除')
      load()
    },
  })
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">我的文章</h1>
      <NButton type="primary" @click="router.push('/dashboard/posts/new')">写文章</NButton>
    </div>
    <NSpace class="mb-4">
      <NSelect
        v-model:value="status"
        clearable
        placeholder="状态"
        :options="[
          { label: '已发布', value: 'published' },
          { label: '草稿', value: 'draft' },
        ]"
        style="width: 140px"
        @update:value="load"
      />
      <NSelect
        v-model:value="visibility"
        clearable
        placeholder="可见性"
        :options="[
          { label: '公开', value: 'public' },
          { label: '私密', value: 'private' },
          { label: '好友', value: 'friends' },
          { label: '指定好友', value: 'specified' },
        ]"
        style="width: 140px"
        @update:value="load"
      />
    </NSpace>
    <PostList :posts="posts" :loading="loading" />
    <div v-for="post in posts" :key="post.id" class="hidden" />
    <div class="mt-4 space-y-2">
      <div v-for="post in posts" :key="'actions-' + post.id" class="flex gap-2 text-sm">
        <NButton size="tiny" @click="router.push(`/dashboard/posts/${post.id}/edit`)">编辑 #{{ post.id }}</NButton>
        <NButton size="tiny" type="error" tertiary @click="remove(post.id)">删除</NButton>
      </div>
    </div>
    <Pagination v-if="total > 12" :page="page" :page-size="12" :total="total" @change="(p) => { page = p; load() }" />
  </div>
</template>
