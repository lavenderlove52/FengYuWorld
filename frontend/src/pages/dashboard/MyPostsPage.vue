<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, useDialog, useMessage } from 'naive-ui'
import { listMyPosts, deletePost } from '@/api/posts'
import type { PostSummary } from '@/types/post'
import Pagination from '@/components/common/Pagination.vue'

const router = useRouter()
const dialog = useDialog()
const message = useMessage()

const posts = ref<PostSummary[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)
const statusFilter = ref<string>('all')
const visibilityFilter = ref<string>('all')

const statusOptions = [
  { value: 'all', label: '全部状态' },
  { value: 'published', label: '已发布' },
  { value: 'draft', label: '草稿' },
]
const visibilityOptions = [
  { value: 'all', label: '全部可见性' },
  { value: 'public', label: '公开' },
  { value: 'private', label: '私密' },
  { value: 'friends', label: '好友' },
  { value: 'specified', label: '指定好友' },
]

const load = async () => {
  loading.value = true
  try {
    const res = await listMyPosts({
      page: page.value,
      page_size: 12,
      status: statusFilter.value === 'all' ? undefined : statusFilter.value,
      visibility: visibilityFilter.value === 'all' ? undefined : visibilityFilter.value,
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

const onFilter = () => { page.value = 1; load() }

const remove = (post: PostSummary) => {
  dialog.warning({
    title: '确认删除',
    content: `「${post.title}」删除后不可恢复，确定继续？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await deletePost(post.id)
      message.success('已删除')
      load()
    },
  })
}

const formatDate = (d: string) => d?.slice(0, 10) ?? ''

const statusBadge: Record<string, { label: string; cls: string }> = {
  published: { label: '已发布', cls: 'bg-emerald-50 text-emerald-600 border-emerald-100' },
  draft: { label: '草稿', cls: 'bg-slate-100 text-slate-500 border-slate-200' },
}
const visibilityBadge: Record<string, { label: string; cls: string }> = {
  public: { label: '公开', cls: 'bg-blue-50 text-blue-500 border-blue-100' },
  private: { label: '私密', cls: 'bg-orange-50 text-orange-500 border-orange-100' },
  friends: { label: '好友', cls: 'bg-violet-50 text-violet-500 border-violet-100' },
  specified: { label: '指定', cls: 'bg-pink-50 text-pink-500 border-pink-100' },
}
</script>

<template>
  <div>
    <!-- 顶部操作栏 -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">我的文章</h1>
        <p class="text-sm text-slate-400 mt-0.5">共 {{ total }} 篇</p>
      </div>
      <NButton type="primary" @click="router.push('/dashboard/posts/new')">
        <template #icon>
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
            <path d="M10.75 4.75a.75.75 0 0 0-1.5 0v4.5h-4.5a.75.75 0 0 0 0 1.5h4.5v4.5a.75.75 0 0 0 1.5 0v-4.5h4.5a.75.75 0 0 0 0-1.5h-4.5v-4.5Z" />
          </svg>
        </template>
        写文章
      </NButton>
    </div>

    <!-- 筛选栏 -->
    <div class="flex gap-2 mb-5 flex-wrap">
      <button
        v-for="opt in statusOptions"
        :key="opt.value"
        class="px-3 py-1.5 rounded-full text-sm border transition-all"
        :class="statusFilter === opt.value
          ? 'bg-slate-800 text-white border-slate-800'
          : 'bg-white text-slate-500 border-slate-200 hover:border-slate-400'"
        @click="statusFilter = opt.value; onFilter()"
      >{{ opt.label }}</button>
      <div class="w-px bg-slate-200 mx-1" />
      <button
        v-for="opt in visibilityOptions"
        :key="opt.value"
        class="px-3 py-1.5 rounded-full text-sm border transition-all"
        :class="visibilityFilter === opt.value
          ? 'bg-slate-800 text-white border-slate-800'
          : 'bg-white text-slate-500 border-slate-200 hover:border-slate-400'"
        @click="visibilityFilter = opt.value; onFilter()"
      >{{ opt.label }}</button>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="text-center py-16 text-slate-400">加载中...</div>

    <!-- 空状态 -->
    <div v-else-if="!posts.length" class="text-center py-20 text-slate-400">
      <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 mx-auto mb-3 text-slate-200" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
      </svg>
      <p>暂无文章</p>
    </div>

    <!-- 文章列表 -->
    <div v-else class="space-y-0">
      <div
        v-for="post in posts"
        :key="post.id"
        class="group flex items-start gap-4 py-4 border-b border-slate-100 hover:bg-slate-50/60 transition-colors rounded-lg px-3 -mx-3"
      >
        <!-- 封面缩略图 -->
        <div class="w-14 h-14 rounded-lg overflow-hidden bg-slate-100 flex-shrink-0 mt-0.5">
          <img
            v-if="post.cover_image_url"
            :src="post.cover_image_url"
            :alt="post.title"
            class="w-full h-full object-cover"
          />
          <div v-else class="w-full h-full flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
            </svg>
          </div>
        </div>

        <!-- 主信息 + 操作按钮 -->
        <div class="flex-1 min-w-0">
          <!-- 第一行：标题 + 徽章 -->
          <div class="flex items-center gap-2 flex-wrap mb-1">
            <h2 class="text-[15px] font-semibold text-slate-800 leading-snug">{{ post.title }}</h2>
            <span
              v-if="statusBadge[post.status]"
              class="text-[11px] px-2 py-0.5 rounded-full border flex-shrink-0"
              :class="statusBadge[post.status].cls"
            >{{ statusBadge[post.status].label }}</span>
            <span
              v-if="visibilityBadge[post.visibility]"
              class="text-[11px] px-2 py-0.5 rounded-full border flex-shrink-0"
              :class="visibilityBadge[post.visibility].cls"
            >{{ visibilityBadge[post.visibility].label }}</span>
          </div>

          <!-- 第二行：摘要 -->
          <p class="text-sm text-slate-400 truncate mb-2">{{ post.summary || '暂无摘要' }}</p>

          <!-- 第三行：元信息 + 操作按钮 同行 -->
          <div class="flex items-center gap-3 text-xs text-slate-400">
            <span>{{ formatDate(post.published_at || post.created_at) }}</span>
            <span>·</span>
            <span>{{ post.view_count }} 阅读</span>
            <span v-if="post.tags?.length">· {{ post.tags.join('、') }}</span>

            <!-- 操作按钮 紧跟元信息后面 -->
            <div class="flex items-center gap-1.5 ml-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <button
                class="px-2.5 py-1 text-xs text-slate-600 bg-white border border-slate-200 rounded-md hover:border-slate-400 hover:text-slate-800 transition"
                @click="router.push(`/dashboard/posts/${post.id}/edit`)"
              >编辑</button>
              <button
                class="px-2.5 py-1 text-xs text-red-500 bg-white border border-red-100 rounded-md hover:bg-red-50 hover:border-red-300 transition"
                @click="remove(post)"
              >删除</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <Pagination
      v-if="total > 12"
      :page="page"
      :page-size="12"
      :total="total"
      class="mt-6"
      @change="(p) => { page = p; load() }"
    />
  </div>
</template>
