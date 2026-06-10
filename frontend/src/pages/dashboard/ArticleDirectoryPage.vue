<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { NButton, NEmpty, NInput, NSpin, useMessage } from 'naive-ui'
import { getPost, updatePostContent } from '@/api/posts'
import { listPostTree } from '@/api/postTree'
import { useAuthStore } from '@/stores/auth'
import { useUserTheme } from '@/composables/useUserTheme'
import type { PostDetail } from '@/types/post'
import type { PostTreeNode } from '@/types/postTree'
import MarkdownViewer from '@/components/post/MarkdownViewer.vue'
import ArticleDirectoryTreeNode from '@/components/post-tree/ArticleDirectoryTreeNode.vue'
import { getPostPath } from '@/utils/postPath'

const message = useMessage()
const auth = useAuthStore()
const containerRef = ref<HTMLElement | null>(null)
const { load: loadTheme } = useUserTheme(containerRef)

const tree = ref<PostTreeNode[]>([])
const expandedIds = ref<Set<number>>(new Set())
const selectedPostId = ref<number | null>(null)
const selectedPost = ref<PostDetail | null>(null)
const draftContent = ref('')
const loadingTree = ref(false)
const loadingPost = ref(false)
const saving = ref(false)
const editing = ref(false)

const flattenNodes = (nodes: PostTreeNode[]): PostTreeNode[] =>
  nodes.flatMap((node) => [node, ...flattenNodes(node.children || [])])

const allNodes = computed(() => flattenNodes(tree.value))
const firstPostNode = computed(() => allNodes.value.find((node) => node.node_type === 'post' && node.post_id) || null)
const folderCount = computed(() => allNodes.value.filter((node) => node.node_type === 'folder').length)
const postCount = computed(() => allNodes.value.filter((node) => node.node_type === 'post').length)

const loadTree = async () => {
  loadingTree.value = true
  try {
    const res = await listPostTree()
    if (res.code === 0 && res.data) {
      tree.value = res.data
      expandAll(res.data)
      if (!selectedPostId.value && firstPostNode.value?.post_id) {
        await selectPost(firstPostNode.value)
      }
    }
  } finally {
    loadingTree.value = false
  }
}

const expandAll = (nodes: PostTreeNode[]) => {
  const next = new Set(expandedIds.value)
  for (const node of nodes) {
    if (node.node_type === 'folder') {
      next.add(node.id)
      expandAll(node.children || [])
    }
  }
  expandedIds.value = next
}

const toggleFolder = (node: PostTreeNode) => {
  const next = new Set(expandedIds.value)
  if (next.has(node.id)) next.delete(node.id)
  else next.add(node.id)
  expandedIds.value = next
}

const selectPost = async (node: PostTreeNode) => {
  if (!node.post_id) return
  selectedPostId.value = node.post_id
  editing.value = false
  loadingPost.value = true
  try {
    const res = await getPost(node.post_id)
    if (res.code === 0 && res.data) {
      selectedPost.value = res.data
      draftContent.value = res.data.content_md || ''
    } else {
      message.error(res.message || '文章加载失败')
    }
  } finally {
    loadingPost.value = false
  }
}

const editPost = async (node: PostTreeNode) => {
  await selectPost(node)
  editing.value = true
}

const saveContent = async () => {
  if (!selectedPost.value) return
  saving.value = true
  try {
    const post = selectedPost.value
    const res = await updatePostContent(post.id, draftContent.value)
    if (res.code === 0) {
      selectedPost.value = res.data ? res.data : { ...post, content_md: draftContent.value }
      draftContent.value = selectedPost.value.content_md || ''
      editing.value = false
      message.success('文章已保存')
    } else {
      message.error(res.message || '保存失败')
    }
  } catch (err) {
    message.error(err instanceof Error ? err.message : '保存失败')
  } finally {
    saving.value = false
  }
}

const openPublicPost = () => {
  if (!selectedPost.value) return
  const base = import.meta.env.BASE_URL.replace(/\/$/, '')
  window.open(`${base}${getPostPath(selectedPost.value)}`, '_blank')
}

onMounted(async () => {
  if (auth.user?.username) {
    await loadTheme(auth.user.username)
  }
  await loadTree()
})
</script>

<template>
  <div
    ref="containerRef"
    class="article-directory-page"
    :data-theme-user="auth.user?.username"
    data-theme-page="article-directory"
  >
    <aside class="directory-sidebar">
      <div class="sidebar-head">
        <div class="head-title">
          <span class="head-icon">
            <svg viewBox="0 0 20 20" aria-hidden="true">
              <path d="M3 5h14v1.5H3V5Zm0 4.25h14v1.5H3v-1.5ZM3 13.5h14V15H3v-1.5Z" />
            </svg>
          </span>
          <span>目录</span>
        </div>
        <NButton size="tiny" quaternary @click="loadTree">刷新</NButton>
      </div>

      <div class="sidebar-stats">
        <span>{{ folderCount }} 个目录</span>
        <span>{{ postCount }} 篇文章</span>
      </div>

      <NSpin :show="loadingTree">
        <div v-if="tree.length" class="directory-tree">
          <template v-for="node in tree" :key="node.id">
            <ArticleDirectoryTreeNode
              :node="node"
              :depth="0"
              :expanded-ids="expandedIds"
              :selected-post-id="selectedPostId"
              @toggle="toggleFolder"
              @select-post="selectPost"
              @edit-post="editPost"
            />
          </template>
        </div>
        <NEmpty v-else-if="!loadingTree" description="还没有文章目录" class="empty-state theme-empty" />
      </NSpin>
    </aside>

    <main class="article-main">
      <NSpin :show="loadingPost">
        <template v-if="selectedPost">
          <div class="article-toolbar">
            <div class="article-meta">
              <h1>{{ selectedPost.title }}</h1>
              <p>
                <span>{{ selectedPost.status === 'draft' ? '草稿' : '已发布' }}</span>
                <span>{{ selectedPost.visibility }}</span>
              </p>
            </div>
            <div class="toolbar-actions">
              <RouterLink to="/" class="reader-link">首页</RouterLink>
              <RouterLink to="/dashboard" class="reader-link">用户中心</RouterLink>
              <NButton v-if="!editing" secondary @click="editing = true">编辑</NButton>
              <NButton v-if="!editing" type="primary" secondary @click="openPublicPost">打开主页</NButton>
              <NButton v-if="editing" secondary @click="editing = false">取消</NButton>
              <NButton v-if="editing" type="primary" :loading="saving" @click="saveContent">保存</NButton>
            </div>
          </div>

          <div v-if="editing" class="editor-split">
            <section class="raw-panel">
              <div class="panel-label">原文档</div>
              <NInput v-model:value="draftContent" type="textarea" class="raw-input" />
            </section>
            <section class="preview-panel">
              <div class="panel-label">展示格式</div>
              <div class="preview-body">
                <MarkdownViewer :content-md="draftContent" />
              </div>
            </section>
          </div>

          <article v-else class="reader-panel">
            <p v-if="selectedPost.summary" class="summary">{{ selectedPost.summary }}</p>
            <MarkdownViewer :content-md="selectedPost.content_md || ''" />
          </article>
        </template>

        <NEmpty v-else-if="!loadingPost" description="从左侧选择一篇文章" class="empty-state theme-empty" />
      </NSpin>
    </main>
  </div>
</template>

<style scoped>
.article-directory-page {
  --reader-bg: var(--theme-bg, #111111);
  --reader-text: var(--theme-text, #e8e8e8);
  --reader-muted: color-mix(in srgb, var(--theme-text, #e8e8e8) 58%, transparent);
  --reader-faint: color-mix(in srgb, var(--theme-text, #e8e8e8) 38%, transparent);
  --reader-border: var(--theme-border, #2f2f2f);
  --reader-panel: color-mix(in srgb, var(--theme-card-bg, #202020) 78%, var(--theme-bg, #111111));
  --reader-hover: color-mix(in srgb, var(--theme-accent, #8ab4f8) 16%, transparent);
  display: grid;
  grid-template-columns: 288px minmax(0, 1fr);
  min-height: 100vh;
  background: var(--reader-bg);
  color: var(--reader-text);
}

.directory-sidebar {
  min-width: 0;
  border-right: 1px solid var(--reader-border);
  background: var(--reader-panel);
  color: var(--reader-text);
}

.sidebar-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  height: 48px;
  padding: 0 12px;
  border-bottom: 1px solid var(--reader-border);
}

.head-title {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  font-size: 16px;
  font-weight: 700;
}

.head-icon {
  display: inline-flex;
  color: var(--reader-muted);
}

.head-icon svg {
  width: 17px;
  height: 17px;
  fill: currentColor;
}

.sidebar-stats {
  display: flex;
  gap: 12px;
  padding: 10px 14px 6px;
  color: var(--reader-faint);
  font-size: 12px;
}

.directory-tree {
  padding: 4px 8px 18px 4px;
}

.article-main {
  min-width: 0;
  background: var(--reader-bg);
}

.article-toolbar {
  position: sticky;
  top: 0;
  z-index: 1;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 22px 48px 12px;
  background: linear-gradient(
    180deg,
    var(--reader-bg) 0%,
    color-mix(in srgb, var(--reader-bg) 94%, transparent) 76%,
    color-mix(in srgb, var(--reader-bg) 0%, transparent) 100%
  );
}

.article-meta {
  min-width: 0;
}

.article-meta h1 {
  margin: 0;
  color: var(--reader-text);
  font-size: 26px;
  font-weight: 800;
  line-height: 1.25;
}

.article-meta p {
  display: flex;
  gap: 10px;
  margin: 8px 0 0;
  color: var(--reader-muted);
  font-size: 13px;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.reader-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 34px;
  padding: 0 10px;
  border-radius: 6px;
  color: var(--reader-muted);
  font-size: 13px;
  text-decoration: none;
}

.reader-link:hover {
  background: var(--reader-hover);
  color: var(--theme-accent, var(--reader-text));
}

.reader-panel {
  --theme-text: var(--reader-text);
  --theme-border: var(--reader-border);
  width: min(860px, calc(100% - 96px));
  margin: 18px auto 80px;
  color: var(--reader-text);
  font-size: 17px;
  line-height: 1.95;
}

.summary {
  margin: 0 0 22px;
  padding-bottom: 18px;
  border-bottom: 1px solid var(--reader-border);
  color: var(--reader-muted);
  font-size: 15px;
  line-height: 1.8;
}

.reader-panel :deep(.markdown-body),
.preview-body :deep(.markdown-body) {
  color: var(--reader-text);
}

.reader-panel :deep(.markdown-body h1),
.reader-panel :deep(.markdown-body h2),
.reader-panel :deep(.markdown-body h3),
.reader-panel :deep(.markdown-body h4),
.reader-panel :deep(.markdown-body h5),
.reader-panel :deep(.markdown-body h6),
.preview-body :deep(.markdown-body h1),
.preview-body :deep(.markdown-body h2),
.preview-body :deep(.markdown-body h3),
.preview-body :deep(.markdown-body h4),
.preview-body :deep(.markdown-body h5),
.preview-body :deep(.markdown-body h6) {
  color: var(--reader-text);
}

.reader-panel :deep(.markdown-body a) {
  color: var(--theme-accent, #8ab4f8);
}

.editor-split {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 1px;
  min-height: calc(100vh - 128px);
  margin-top: 4px;
  border-top: 1px solid var(--reader-border);
  background: var(--reader-border);
}

.raw-panel,
.preview-panel {
  min-width: 0;
  background: var(--reader-bg);
}

.panel-label {
  height: 38px;
  padding: 10px 18px;
  border-bottom: 1px solid var(--reader-border);
  color: var(--reader-muted);
  font-size: 13px;
  font-weight: 700;
}

.raw-input {
  height: calc(100vh - 166px);
}

.raw-input :deep(.n-input) {
  height: 100%;
  border-radius: 0;
  background: var(--reader-bg);
}

.raw-input :deep(.n-input-wrapper) {
  background: var(--reader-bg);
}

.raw-input :deep(textarea) {
  min-height: calc(100vh - 166px);
  padding: 20px;
  color: var(--reader-text);
  caret-color: var(--theme-accent, var(--reader-text));
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 14px;
  line-height: 1.75;
}

.preview-body {
  --theme-text: var(--reader-text);
  --theme-border: var(--reader-border);
  height: calc(100vh - 166px);
  overflow: auto;
  padding: 20px 28px 80px;
  color: var(--reader-text);
}

.empty-state {
  padding: 80px 0;
}

.theme-empty :deep(.n-empty__description) {
  color: var(--reader-muted);
}

@media (max-width: 1024px) {
  .article-directory-page {
    grid-template-columns: 1fr;
  }

  .directory-sidebar {
    max-height: 40vh;
    overflow: auto;
    border-right: 0;
    border-bottom: 1px solid var(--reader-border);
  }

  .article-toolbar {
    padding: 18px 18px 10px;
  }

  .reader-panel {
    width: auto;
    margin: 18px 18px 56px;
  }

  .editor-split {
    grid-template-columns: 1fr;
  }
}
</style>
