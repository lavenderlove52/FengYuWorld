<script setup lang="ts">
import { computed } from 'vue'
import type { PostTreeNode } from '@/types/postTree'

const props = defineProps<{
  node: PostTreeNode
  depth: number
  selectedId?: number | null
  expandedIds: Set<number>
}>()

const emit = defineEmits<{
  select: [node: PostTreeNode]
  toggle: [node: PostTreeNode]
  addFolder: [node: PostTreeNode]
  addPost: [node: PostTreeNode]
  rename: [node: PostTreeNode]
  remove: [node: PostTreeNode]
  moveUp: [node: PostTreeNode]
  moveDown: [node: PostTreeNode]
}>()

const isFolder = computed(() => props.node.node_type === 'folder')
const isExpanded = computed(() => props.expandedIds.has(props.node.id))
const isSelected = computed(() => props.selectedId === props.node.id)
const title = computed(() => props.node.post?.title || props.node.name)
const statusText = computed(() => {
  if (!props.node.post) return ''
  return props.node.post.status === 'draft' ? '草稿' : '已发布'
})
</script>

<template>
  <div class="tree-node">
    <div
      class="tree-row group"
      :class="{ 'tree-row-selected': isSelected }"
      :style="{ paddingLeft: `${depth * 18 + 8}px` }"
      @click="emit('select', node)"
    >
      <button
        v-if="isFolder"
        type="button"
        class="tree-toggle"
        :aria-label="isExpanded ? '收起' : '展开'"
        @click.stop="emit('toggle', node)"
      >
        {{ isExpanded ? '⌄' : '›' }}
      </button>
      <span v-else class="tree-toggle-placeholder" />

      <span class="tree-icon">{{ isFolder ? '▣' : '◇' }}</span>
      <span class="tree-title" :title="title">{{ title }}</span>
      <span v-if="statusText" class="tree-badge">{{ statusText }}</span>

      <span class="tree-actions">
        <button type="button" title="上移" @click.stop="emit('moveUp', node)">↑</button>
        <button type="button" title="下移" @click.stop="emit('moveDown', node)">↓</button>
        <button v-if="isFolder" type="button" title="新建子目录" @click.stop="emit('addFolder', node)">＋</button>
        <button v-if="isFolder" type="button" title="加入文章" @click.stop="emit('addPost', node)">◇</button>
        <button v-if="isFolder" type="button" title="重命名" @click.stop="emit('rename', node)">✎</button>
        <button type="button" title="移除" @click.stop="emit('remove', node)">×</button>
      </span>
    </div>

    <div v-if="isFolder && isExpanded">
      <PostTreeNode
        v-for="child in node.children || []"
        :key="child.id"
        :node="child"
        :depth="depth + 1"
        :selected-id="selectedId"
        :expanded-ids="expandedIds"
        @select="emit('select', $event)"
        @toggle="emit('toggle', $event)"
        @add-folder="emit('addFolder', $event)"
        @add-post="emit('addPost', $event)"
        @rename="emit('rename', $event)"
        @remove="emit('remove', $event)"
        @move-up="emit('moveUp', $event)"
        @move-down="emit('moveDown', $event)"
      />
    </div>
  </div>
</template>

<style scoped>
.tree-row {
  display: flex;
  align-items: center;
  gap: 6px;
  min-height: 34px;
  padding-top: 4px;
  padding-bottom: 4px;
  padding-right: 8px;
  border-radius: 6px;
  color: #334155;
  cursor: pointer;
}

.tree-row:hover,
.tree-row-selected {
  background: #eef2ff;
  color: #0f172a;
}

.tree-toggle,
.tree-toggle-placeholder {
  width: 20px;
  height: 24px;
  flex: 0 0 20px;
}

.tree-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
}

.tree-toggle:hover {
  background: #dbeafe;
}

.tree-icon {
  width: 18px;
  flex: 0 0 18px;
  color: #64748b;
  font-size: 13px;
}

.tree-title {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
}

.tree-badge {
  flex: 0 0 auto;
  padding: 1px 6px;
  border-radius: 999px;
  background: #f1f5f9;
  color: #64748b;
  font-size: 11px;
}

.tree-actions {
  display: none;
  align-items: center;
  gap: 2px;
  flex: 0 0 auto;
}

.tree-row:hover .tree-actions,
.tree-row-selected .tree-actions {
  display: inline-flex;
}

.tree-actions button {
  width: 24px;
  height: 24px;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: #64748b;
  cursor: pointer;
  font-size: 13px;
}

.tree-actions button:hover {
  background: #fff;
  color: #1d4ed8;
}
</style>
