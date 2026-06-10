<script setup lang="ts">
import { computed } from 'vue'
import type { PostTreeNode } from '@/types/postTree'

const props = defineProps<{
  node: PostTreeNode
  depth: number
  expandedIds: Set<number>
  selectedPostId?: number | null
}>()

const emit = defineEmits<{
  toggle: [node: PostTreeNode]
  selectPost: [node: PostTreeNode]
  editPost: [node: PostTreeNode]
}>()

const isFolder = computed(() => props.node.node_type === 'folder')
const isExpanded = computed(() => props.expandedIds.has(props.node.id))
const isSelected = computed(() => props.node.post_id && props.node.post_id === props.selectedPostId)
const title = computed(() => props.node.post?.title || props.node.name)
</script>

<template>
  <div>
    <div
      class="tree-row"
      :class="{ 'tree-row-selected': isSelected }"
      :style="{ paddingLeft: `${depth * 18 + 10}px` }"
      @click="isFolder ? emit('toggle', node) : emit('selectPost', node)"
    >
      <button
        type="button"
        class="tree-toggle"
        :class="{ invisible: !isFolder }"
        :title="isExpanded ? '收起' : '展开'"
        @click.stop="isFolder && emit('toggle', node)"
      >
        <span>{{ isExpanded ? '⌄' : '›' }}</span>
      </button>

      <span class="tree-title" :title="title">{{ title }}</span>

      <button
        v-if="!isFolder"
        type="button"
        class="tree-edit"
        title="编辑文章"
        @click.stop="emit('editPost', node)"
      >
        <svg viewBox="0 0 20 20" aria-hidden="true">
          <path d="M13.86 3.57a1.95 1.95 0 0 1 2.76 2.76l-.78.78-2.76-2.76.78-.78Zm-1.84 1.84 2.76 2.76-7.7 7.7-3.15.39.39-3.15 7.7-7.7Z" />
        </svg>
      </button>
    </div>

    <div v-if="isFolder && isExpanded">
      <ArticleDirectoryTreeNode
        v-for="child in node.children || []"
        :key="child.id"
        :node="child"
        :depth="depth + 1"
        :expanded-ids="expandedIds"
        :selected-post-id="selectedPostId"
        @toggle="emit('toggle', $event)"
        @select-post="emit('selectPost', $event)"
        @edit-post="emit('editPost', $event)"
      />
    </div>
  </div>
</template>

<style scoped>
.tree-row {
  display: flex;
  align-items: center;
  gap: 4px;
  min-height: 34px;
  border-radius: 6px;
  color: var(--reader-text, #d4d4d4);
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.tree-row:hover {
  background: var(--reader-hover, #2a2a2a);
  color: var(--reader-text, #f5f5f5);
}

.tree-row-selected {
  background: color-mix(in srgb, var(--theme-accent, #8ab4f8) 24%, transparent);
  color: var(--reader-text, #ffffff);
  font-weight: 700;
}

.tree-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 24px;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: var(--reader-muted, #a3a3a3);
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
}

.tree-toggle:hover {
  background: var(--reader-hover, #3f3f3f);
  color: var(--theme-accent, #ffffff);
}

.tree-toggle.invisible {
  pointer-events: none;
  opacity: 0;
}

.tree-title {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
  line-height: 1.4;
}

.tree-edit {
  display: none;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  margin-right: 4px;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: var(--reader-muted, #a3a3a3);
  cursor: pointer;
}

.tree-edit svg {
  width: 14px;
  height: 14px;
  fill: currentColor;
}

.tree-row:hover .tree-edit,
.tree-row-selected .tree-edit {
  display: inline-flex;
}

.tree-edit:hover {
  background: var(--reader-hover, #4a4a4a);
  color: var(--theme-accent, #ffffff);
}
</style>
