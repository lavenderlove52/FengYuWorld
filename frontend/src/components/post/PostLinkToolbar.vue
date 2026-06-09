<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { NormalToolbar } from 'md-editor-v3'
import { NButton, NEmpty, NInput, NModal, NSpin, useMessage } from 'naive-ui'
import { searchLinkablePosts } from '@/api/posts'
import type { PostSummary } from '@/types/post'
import { getPostPath } from '@/utils/postPath'

type Insert = (
  generate: (selectedText: string) => {
    targetValue: string
    select?: boolean
    deviationStart?: number
    deviationEnd?: number
  },
) => void

const props = defineProps<{
  insert?: Insert
}>()

const message = useMessage()
const visible = ref(false)
const keyword = ref('')
const posts = ref<PostSummary[]>([])
const loading = ref(false)
let searchTimer: ReturnType<typeof setTimeout> | null = null

const hasPosts = computed(() => posts.value.length > 0)

const escapeMarkdownText = (text: string) => text.replace(/([\\[\]])/g, '\\$1')

const loadPosts = async () => {
  loading.value = true
  try {
    const res = await searchLinkablePosts(keyword.value.trim(), 1, 20)
    if (res.code === 0 && res.data) {
      posts.value = res.data.items
    } else {
      message.error(res.message || '文章搜索失败')
    }
  } finally {
    loading.value = false
  }
}

const openPicker = () => {
  visible.value = true
  if (!posts.value.length) {
    void loadPosts()
  }
}

const onSearchInput = () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(loadPosts, 300)
}

const insertPostLink = (post: PostSummary) => {
  if (!props.insert) return
  const url = getPostPath(post)
  props.insert((selectedText) => {
    const text = escapeMarkdownText((selectedText || post.title).trim() || post.title)
    return {
      targetValue: `[${text}](${url})`,
      select: false,
    }
  })
  visible.value = false
}

onBeforeUnmount(() => {
  if (searchTimer) clearTimeout(searchTimer)
})
</script>

<template>
  <NormalToolbar title="插入文章链接" @on-click="openPicker">
    <span class="post-link-toolbar-icon">↗</span>
  </NormalToolbar>

  <NModal
    v-model:show="visible"
    preset="card"
    title="插入文章链接"
    class="post-link-modal"
    :bordered="false"
    :segmented="{ content: true }"
  >
    <div class="post-link-picker">
      <NInput
        v-model:value="keyword"
        clearable
        placeholder="搜索标题、摘要或作者"
        @input="onSearchInput"
        @keyup.enter="loadPosts"
      />

      <NSpin :show="loading">
        <div v-if="hasPosts" class="post-link-list">
          <button
            v-for="post in posts"
            :key="post.id"
            type="button"
            class="post-link-item"
            @click="insertPostLink(post)"
          >
            <span class="post-link-title">{{ post.title }}</span>
            <span class="post-link-meta">
              {{ post.author.nickname || post.author.username }} · {{ post.status === 'draft' ? '草稿' : (post.published_at || post.created_at).slice(0, 10) }}
            </span>
          </button>
        </div>
        <NEmpty v-else-if="!loading" description="没有找到可链接的文章" />
      </NSpin>
    </div>

    <template #footer>
      <div class="post-link-footer">
        <NButton @click="visible = false">取消</NButton>
      </div>
    </template>
  </NModal>
</template>

<style scoped>
.post-link-toolbar-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  font-size: 15px;
  line-height: 1;
}

.post-link-picker {
  display: grid;
  gap: 12px;
}

.post-link-list {
  display: grid;
  gap: 8px;
  max-height: 360px;
  overflow: auto;
}

.post-link-item {
  display: grid;
  gap: 4px;
  width: 100%;
  padding: 10px 12px;
  text-align: left;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: #fff;
  color: #0f172a;
  cursor: pointer;
  transition:
    border-color 0.2s,
    background-color 0.2s;
}

.post-link-item:hover {
  border-color: #6366f1;
  background: #f8fafc;
}

.post-link-title {
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
}

.post-link-meta {
  font-size: 12px;
  color: #64748b;
  line-height: 1.4;
}

.post-link-footer {
  display: flex;
  justify-content: flex-end;
}
</style>

<style>
.post-link-modal {
  width: min(620px, calc(100vw - 32px));
}
</style>
