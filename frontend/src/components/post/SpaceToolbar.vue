<script setup lang="ts">
import { ref } from 'vue'
import { NormalToolbar } from 'md-editor-v3'
import { NPopover } from 'naive-ui'

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

const visible = ref(false)

const SPACE_OPTIONS = [
  { label: '&emsp;&emsp;', desc: '段首缩进（两个全角空格）', value: '&emsp;&emsp;' },
  { label: '&emsp;', desc: '全角空格（em）', value: '&emsp;' },
  { label: '&nbsp;&nbsp;', desc: '两个半角不换行空格', value: '&nbsp;&nbsp;' },
  { label: '&nbsp;', desc: '半角不换行空格', value: '&nbsp;' },
  { label: '&ensp;', desc: '半宽空格（en）', value: '&ensp;' },
  { label: '　', desc: 'Unicode 全角空格', value: '　' },
]

const insertSpace = (value: string) => {
  if (!props.insert) return
  props.insert(() => ({ targetValue: value, select: false }))
  visible.value = false
}
</script>

<template>
  <NPopover v-model:show="visible" trigger="click" placement="bottom" :show-arrow="false">
    <template #trigger>
      <NormalToolbar title="插入空格">
        <span class="space-toolbar-icon">⎵</span>
      </NormalToolbar>
    </template>

    <div class="space-menu">
      <p class="space-menu-title">插入空格</p>
      <button
        v-for="opt in SPACE_OPTIONS"
        :key="opt.value"
        type="button"
        class="space-menu-item"
        @click="insertSpace(opt.value)"
      >
        <code class="space-menu-code">{{ opt.label }}</code>
        <span class="space-menu-desc">{{ opt.desc }}</span>
      </button>
    </div>
  </NPopover>
</template>

<style scoped>
.space-toolbar-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  font-size: 15px;
  line-height: 1;
}

.space-menu {
  display: grid;
  gap: 4px;
  min-width: 220px;
  padding: 4px 0;
}

.space-menu-title {
  padding: 2px 10px 6px;
  font-size: 12px;
  color: #94a3b8;
  font-weight: 500;
}

.space-menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 6px 10px;
  background: transparent;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  text-align: left;
  transition: background-color 0.15s;
}

.space-menu-item:hover {
  background: #f1f5f9;
}

.space-menu-code {
  flex-shrink: 0;
  min-width: 90px;
  padding: 2px 6px;
  font-size: 12px;
  font-family: ui-monospace, 'Cascadia Code', monospace;
  background: #f1f5f9;
  border-radius: 3px;
  color: #6366f1;
}

.space-menu-desc {
  font-size: 12px;
  color: #64748b;
}
</style>
