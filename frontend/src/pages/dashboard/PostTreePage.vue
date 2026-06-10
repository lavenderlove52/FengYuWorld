<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NEmpty, NInput, NModal, NSelect, NSpin, useDialog, useMessage } from 'naive-ui'
import { listMyPosts } from '@/api/posts'
import {
  addPostTreePost,
  createPostTreeFolder,
  deletePostTreeNode,
  listPostTree,
  updatePostTreeNode,
} from '@/api/postTree'
import type { PostSummary } from '@/types/post'
import type { PostTreeNode } from '@/types/postTree'
import PostTreeNodeView from '@/components/post-tree/PostTreeNode.vue'
import { getPostPath } from '@/utils/postPath'

const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const ROOT_PARENT_ID = 0

const tree = ref<PostTreeNode[]>([])
const posts = ref<PostSummary[]>([])
const loading = ref(false)
const selectedId = ref<number | null>(null)
const expandedIds = ref<Set<number>>(new Set())

const folderModalVisible = ref(false)
const folderName = ref('')
const folderParentId = ref<number | null>(null)
const editingFolder = ref<PostTreeNode | null>(null)

const postModalVisible = ref(false)
const postParentId = ref<number | null>(null)
const selectedPostId = ref<number | null>(null)

const moveParentId = ref<number | null>(null)
const draggingNode = ref<PostTreeNode | null>(null)
const dragTargetKey = ref('')
const rootDropActive = ref(false)

const flattenNodes = (nodes: PostTreeNode[]): PostTreeNode[] =>
  nodes.flatMap((node) => [node, ...flattenNodes(node.children || [])])

const allNodes = computed(() => flattenNodes(tree.value))
const selectedNode = computed(() => allNodes.value.find((node) => node.id === selectedId.value) || null)
const selectedFolder = computed(() => selectedNode.value?.node_type === 'folder' ? selectedNode.value : null)
const selectedFolderHasPosts = computed(() => (selectedFolder.value?.children || []).some((child) => child.node_type === 'post'))
const usedPostIds = computed(() => new Set(allNodes.value.filter((node) => node.node_type === 'post' && node.post_id).map((node) => node.post_id!)))
const availablePosts = computed(() => posts.value.filter((post) => !usedPostIds.value.has(post.id)))

const folderOptions = computed<any[]>(() => [
  { label: '根目录', value: ROOT_PARENT_ID },
  ...allNodes.value
    .filter((node) => node.node_type === 'folder')
    .map((node) => ({
      label: buildFolderLabel(node),
      value: node.id,
      disabled:
        (selectedNode.value?.node_type === 'folder' && isDescendantOrSelf(node.id, selectedNode.value.id)) ||
        (node.children || []).some((child) => child.node_type === 'post'),
    })),
])

const postOptions = computed(() =>
  availablePosts.value.map((post) => ({
    label: `${post.title}${post.status === 'draft' ? '（草稿）' : ''}`,
    value: post.id,
  })),
)

const currentTargetLabel = computed(() => selectedFolder.value?.name || '根目录')

const load = async () => {
  loading.value = true
  try {
    const [treeRes, postsRes] = await Promise.all([
      listPostTree(),
      listMyPosts({ page: 1, page_size: 200 }),
    ])
    if (treeRes.code === 0 && treeRes.data) {
      tree.value = treeRes.data
      expandAllFolders(tree.value)
    }
    if (postsRes.code === 0 && postsRes.data) {
      posts.value = postsRes.data.items
    }
  } finally {
    loading.value = false
  }
}

const expandAllFolders = (nodes: PostTreeNode[]) => {
  const next = new Set(expandedIds.value)
  for (const node of nodes) {
    if (node.node_type === 'folder') {
      next.add(node.id)
      expandAllFolders(node.children || [])
    }
  }
  expandedIds.value = next
}

const selectNode = (node: PostTreeNode) => {
  selectedId.value = node.id
  moveParentId.value = node.parent_id ?? ROOT_PARENT_ID
}

const toggleNode = (node: PostTreeNode) => {
  const next = new Set(expandedIds.value)
  if (next.has(node.id)) next.delete(node.id)
  else next.add(node.id)
  expandedIds.value = next
}

const normalizeParentId = (value: number | null) => value === ROOT_PARENT_ID ? null : value

const openCreateFolder = (parent?: PostTreeNode | null) => {
  editingFolder.value = null
  folderName.value = ''
  folderParentId.value = parent?.id ?? selectedFolder.value?.id ?? ROOT_PARENT_ID
  folderModalVisible.value = true
}

const openRenameFolder = (node: PostTreeNode) => {
  editingFolder.value = node
  folderName.value = node.name
  folderParentId.value = node.parent_id ?? ROOT_PARENT_ID
  folderModalVisible.value = true
}

const submitFolder = async () => {
  const name = folderName.value.trim()
  if (!name) {
    message.warning('目录名称不能为空')
    return
  }
  const parentId = normalizeParentId(folderParentId.value)
  const res = editingFolder.value
    ? await updatePostTreeNode(editingFolder.value.id, {
        name,
        parent_id: parentId,
        sort_order: editingFolder.value.sort_order,
      })
    : await createPostTreeFolder({
        name,
        parent_id: parentId,
        sort_order: nextSortOrder(parentId),
      })
  if (res.code === 0) {
    message.success(editingFolder.value ? '目录已重命名' : '目录已创建')
    folderModalVisible.value = false
    await load()
  } else {
    message.error(res.message || '保存失败')
  }
}

const openAddPost = (parent?: PostTreeNode | null) => {
  postParentId.value = parent?.id ?? selectedFolder.value?.id ?? ROOT_PARENT_ID
  selectedPostId.value = null
  postModalVisible.value = true
}

const submitPost = async () => {
  if (!selectedPostId.value) {
    message.warning('请选择文章')
    return
  }
  const parentId = normalizeParentId(postParentId.value)
  const res = await addPostTreePost({
    parent_id: parentId,
    post_id: selectedPostId.value,
    sort_order: nextSortOrder(parentId),
  })
  if (res.code === 0) {
    message.success('文章已加入目录')
    postModalVisible.value = false
    await load()
  } else {
    message.error(res.message || '加入失败')
  }
}

const removeNode = (node: PostTreeNode) => {
  const isFolder = node.node_type === 'folder'
  dialog.warning({
    title: isFolder ? '移除目录' : '移除文章引用',
    content: isFolder
      ? `确定移除「${node.name}」及其下所有目录引用吗？文章本体不会被删除。`
      : `确定从目录中移除「${node.post?.title || node.name}」吗？文章本体不会被删除。`,
    positiveText: '移除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res = await deletePostTreeNode(node.id)
      if (res.code === 0) {
        if (selectedId.value === node.id) selectedId.value = null
        message.success('已移除')
        await load()
      } else {
        message.error(res.message || '移除失败')
      }
    },
  })
}

const moveNode = async () => {
  if (!selectedNode.value) return
  const node = selectedNode.value
  const parentId = normalizeParentId(moveParentId.value)
  const res = await updatePostTreeNode(node.id, {
    parent_id: parentId,
    name: node.name,
    sort_order: nextSortOrder(parentId),
  })
  if (res.code === 0) {
    message.success('节点已移动')
    await load()
  } else {
    message.error(res.message || '移动失败')
  }
}

const openPostNode = (node: PostTreeNode) => {
  if (!node.post) return
  router.push(getPostPath(node.post))
}

const editPostNode = (node: PostTreeNode) => {
  if (!node.post_id) return
  router.push(`/dashboard/posts/${node.post_id}/edit`)
}

const onDragStart = (node: PostTreeNode, event: DragEvent) => {
  draggingNode.value = node
  selectedId.value = node.id
  event.dataTransfer?.setData('text/plain', String(node.id))
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
  }
}

const clearDragState = () => {
  draggingNode.value = null
  dragTargetKey.value = ''
  rootDropActive.value = false
}

const onDragOverNode = (target: PostTreeNode, position: 'before' | 'inside' | 'after', event: DragEvent) => {
  if (!canDrop(target, position)) return
  event.preventDefault()
  dragTargetKey.value = `${target.id}:${position}`
  if (event.dataTransfer) event.dataTransfer.dropEffect = 'move'
}

const onDropNode = async (target: PostTreeNode, position: 'before' | 'inside' | 'after', event: DragEvent) => {
  event.preventDefault()
  if (!canDrop(target, position) || !draggingNode.value) return
  const dragged = draggingNode.value
  clearDragState()

  if (position === 'inside') {
    await persistOrder(target.id, [...getChildren(target.id).filter((node) => node.id !== dragged.id), dragged])
    expandedIds.value = new Set(expandedIds.value).add(target.id)
    return
  }

  const parentId = target.parent_id ?? null
  const siblings = getChildren(parentId).filter((node) => node.id !== dragged.id)
  const targetIndex = siblings.findIndex((node) => node.id === target.id)
  const insertIndex = position === 'before' ? targetIndex : targetIndex + 1
  siblings.splice(insertIndex, 0, dragged)
  await persistOrder(parentId, siblings)
}

const onRootDragOver = (event: DragEvent) => {
  if (!draggingNode.value || wouldMoveInto(draggingNode.value, null)) return
  event.preventDefault()
  rootDropActive.value = true
  dragTargetKey.value = 'root'
  if (event.dataTransfer) event.dataTransfer.dropEffect = 'move'
}

const onRootDrop = async (event: DragEvent) => {
  event.preventDefault()
  if (!draggingNode.value) return
  const dragged = draggingNode.value
  clearDragState()
  await persistOrder(null, [...tree.value.filter((node) => node.id !== dragged.id), dragged])
}

const canDrop = (target: PostTreeNode, position: 'before' | 'inside' | 'after') => {
  const dragged = draggingNode.value
  if (!dragged || dragged.id === target.id) return false
  const nextParentId = position === 'inside' ? target.id : target.parent_id ?? null
  if (position === 'inside' && target.node_type !== 'folder') return false
  return !wouldMoveInto(dragged, nextParentId)
}

const wouldMoveInto = (dragged: PostTreeNode, parentId: number | null) => {
  return dragged.node_type === 'folder' && parentId !== null && isDescendantOrSelf(parentId, dragged.id)
}

const persistOrder = async (parentId: number | null, orderedNodes: PostTreeNode[]) => {
  const updates = orderedNodes.map((node, index) =>
    updatePostTreeNode(node.id, {
      parent_id: parentId,
      name: node.name,
      sort_order: (index + 1) * 10,
    }),
  )
  const results = await Promise.all(updates)
  const failed = results.find((res) => res.code !== 0)
  if (failed) {
    message.error(failed.message || '移动失败')
    await load()
    return
  }
  message.success('目录已更新')
  await load()
}

const getChildren = (parentId: number | null) => {
  if (parentId === null) return tree.value
  return allNodes.value.find((node) => node.id === parentId)?.children || []
}

const nextSortOrder = (parentId: number | null) => {
  const children = getChildren(parentId)
  return children.reduce((max, child) => Math.max(max, child.sort_order), 0) + 10
}

function buildFolderLabel(node: PostTreeNode) {
  const names = [node.name]
  let parentId = node.parent_id ?? null
  while (parentId) {
    const parent = allNodes.value.find((item) => item.id === parentId)
    if (!parent) break
    names.unshift(parent.name)
    parentId = parent.parent_id ?? null
  }
  return names.join(' / ')
}

function isDescendantOrSelf(candidateId: number, nodeId: number): boolean {
  if (candidateId === nodeId) return true
  let current = allNodes.value.find((item) => item.id === candidateId)
  while (current?.parent_id) {
    if (current.parent_id === nodeId) return true
    current = allNodes.value.find((item) => item.id === current?.parent_id)
  }
  return false
}

onMounted(load)
</script>

<template>
  <div class="post-tree-page">
    <div class="page-head">
      <div>
        <h1>文章目录</h1>
        <p>把文章整理成只属于自己的路径树。拖动目录或文章即可调整层级和顺序。</p>
      </div>
      <div class="head-actions">
        <NButton @click="load">刷新</NButton>
        <NButton @click="openCreateFolder(null)">新建目录</NButton>
        <NButton type="primary" @click="openAddPost(null)">加入文章</NButton>
      </div>
    </div>

    <div class="planner-shell">
      <section class="tree-panel">
        <div class="panel-title">
          <span>文章目录</span>
          <span>{{ allNodes.length }} 个节点</span>
        </div>
        <NSpin :show="loading">
          <div
            v-if="tree.length"
            class="tree-list"
            :class="{ 'tree-list-root-drop': rootDropActive }"
            @dragover.prevent="onRootDragOver"
            @dragleave="rootDropActive = false"
            @drop.prevent="onRootDrop"
          >
            <PostTreeNodeView
              v-for="node in tree"
              :key="node.id"
              :node="node"
              :depth="0"
              :selected-id="selectedId"
              :expanded-ids="expandedIds"
              :drag-target-key="dragTargetKey"
              @select="selectNode"
              @toggle="toggleNode"
              @add-folder="openCreateFolder"
              @add-post="openAddPost"
              @rename="openRenameFolder"
              @remove="removeNode"
              @open-post="openPostNode"
              @edit-post="editPostNode"
              @drag-start="onDragStart"
              @drag-end="clearDragState"
              @drag-over="onDragOverNode"
              @drop-node="onDropNode"
            />
          </div>
          <div v-else-if="!loading" class="empty-tree">
            <NEmpty description="还没有目录">
              <template #extra>
                <NButton type="primary" @click="openCreateFolder(null)">创建第一个目录</NButton>
              </template>
            </NEmpty>
          </div>
        </NSpin>
      </section>

      <aside class="detail-panel">
        <template v-if="selectedNode">
          <div class="panel-title">
            <span>{{ selectedNode.node_type === 'folder' ? '目录' : '文章' }}</span>
            <span>#{{ selectedNode.id }}</span>
          </div>
          <h2>{{ selectedNode.post?.title || selectedNode.name }}</h2>
          <p v-if="selectedNode.post" class="muted">
            {{ selectedNode.post.status === 'draft' ? '草稿' : '已发布' }} · {{ selectedNode.post.visibility }}
          </p>
          <p v-else class="muted">当前目录下可以继续创建目录或加入文章。</p>

          <div class="detail-actions">
            <NButton v-if="selectedNode.node_type === 'folder' && !selectedFolderHasPosts" @click="openCreateFolder(selectedNode)">新建子目录</NButton>
            <NButton v-if="selectedNode.node_type === 'folder'" @click="openAddPost(selectedNode)">加入文章</NButton>
            <NButton v-if="selectedNode.node_type === 'folder'" @click="openRenameFolder(selectedNode)">重命名</NButton>
            <NButton v-if="selectedNode.node_type === 'post'" type="primary" secondary @click="openPostNode(selectedNode)">查看文章</NButton>
            <NButton v-if="selectedNode.node_type === 'post'" @click="editPostNode(selectedNode)">编辑文章</NButton>
            <NButton type="error" ghost @click="removeNode(selectedNode)">移除</NButton>
          </div>

          <div class="move-box">
            <label>移动到</label>
            <NSelect v-model:value="moveParentId" :options="folderOptions" />
            <NButton type="primary" secondary @click="moveNode">移动</NButton>
          </div>
        </template>
        <template v-else>
          <div class="panel-title">
            <span>当前目标</span>
            <span>根目录</span>
          </div>
          <h2>{{ currentTargetLabel }}</h2>
          <p class="muted">选择左侧目录后，可以在这里继续整理文章结构。</p>
          <div class="detail-actions">
            <NButton @click="openCreateFolder(null)">新建根目录</NButton>
            <NButton type="primary" @click="openAddPost(null)">加入文章</NButton>
          </div>
        </template>
      </aside>
    </div>

    <NModal v-model:show="folderModalVisible" preset="card" class="tree-modal" :title="editingFolder ? '重命名目录' : '新建目录'">
      <div class="modal-form">
        <label>目录名称</label>
        <NInput v-model:value="folderName" placeholder="例如：研究生 / 简历相关面试题" @keyup.enter="submitFolder" />
        <label>上级目录</label>
        <NSelect v-model:value="folderParentId" :options="folderOptions" />
      </div>
      <template #footer>
        <div class="modal-footer">
          <NButton @click="folderModalVisible = false">取消</NButton>
          <NButton type="primary" @click="submitFolder">保存</NButton>
        </div>
      </template>
    </NModal>

    <NModal v-model:show="postModalVisible" preset="card" class="tree-modal" title="加入文章">
      <div class="modal-form">
        <label>目标目录</label>
        <NSelect v-model:value="postParentId" :options="folderOptions" />
        <label>文章</label>
        <NSelect
          v-model:value="selectedPostId"
          filterable
          clearable
          :options="postOptions"
          placeholder="选择还没有加入目录树的文章"
        />
        <p class="muted">已加入目录树的文章不会重复显示。</p>
      </div>
      <template #footer>
        <div class="modal-footer">
          <NButton @click="postModalVisible = false">取消</NButton>
          <NButton type="primary" @click="submitPost">加入</NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.post-tree-page {
  display: grid;
  gap: 18px;
}

.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.page-head h1 {
  margin: 0;
  color: #0f172a;
  font-size: 24px;
  font-weight: 700;
}

.page-head p,
.muted {
  margin: 6px 0 0;
  color: #64748b;
  font-size: 14px;
  line-height: 1.6;
}

.head-actions,
.detail-actions,
.modal-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.planner-shell {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  gap: 16px;
}

.tree-panel,
.detail-panel {
  min-width: 0;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #fff;
}

.tree-panel {
  min-height: 540px;
  padding: 12px;
}

.detail-panel {
  align-self: start;
  padding: 16px;
}

.panel-title {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  color: #64748b;
  font-size: 12px;
}

.panel-title span:first-child {
  color: #334155;
  font-weight: 700;
}

.tree-list {
  display: grid;
  align-content: start;
  gap: 2px;
  min-height: 440px;
  padding: 4px;
  border: 1px dashed transparent;
  border-radius: 8px;
}

.tree-list-root-drop {
  border-color: #6366f1;
  background: #f8fafc;
}

.empty-tree {
  display: grid;
  min-height: 420px;
  place-items: center;
}

.detail-panel h2 {
  margin: 0;
  color: #0f172a;
  font-size: 18px;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

.detail-actions {
  margin-top: 16px;
}

.move-box {
  display: grid;
  gap: 8px;
  margin-top: 18px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

.move-box label,
.modal-form label {
  color: #334155;
  font-size: 13px;
  font-weight: 700;
}

.modal-form {
  display: grid;
  gap: 10px;
}

.modal-footer {
  justify-content: flex-end;
}

.tree-modal {
  width: min(520px, calc(100vw - 32px));
}

@media (max-width: 900px) {
  .page-head,
  .planner-shell {
    grid-template-columns: 1fr;
  }

  .page-head {
    display: grid;
  }
}
</style>
