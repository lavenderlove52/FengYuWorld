<script setup lang="ts">
import { computed } from 'vue'
import { extractTocItems } from '@/utils/markdownToc'

const props = defineProps<{ contentMd: string }>()

const items = computed(() => extractTocItems(props.contentMd))
</script>

<template>
  <nav v-if="items.length" class="article-toc">
    <div class="text-xs font-semibold tracking-widest uppercase mb-3" style="color: var(--theme-text, #64748b); opacity: 0.55">
      目录
    </div>
    <a
      v-for="item in items"
      :key="item.id"
      :href="`#${item.id}`"
      class="toc-link"
      :style="{ paddingLeft: `${Math.max(0, item.level - 1) * 12}px` }"
    >
      {{ item.text }}
    </a>
  </nav>
</template>

<style scoped>
.article-toc {
  position: sticky;
  top: 84px;
  max-height: calc(100vh - 112px);
  overflow: auto;
  padding: 14px 0;
}

.toc-link {
  display: block;
  padding-top: 6px;
  padding-bottom: 6px;
  font-size: 13px;
  line-height: 1.35;
  color: var(--theme-text, #475569);
  opacity: 0.7;
  border-left: 2px solid transparent;
  transition: color 0.15s ease, opacity 0.15s ease, border-color 0.15s ease;
}

.toc-link:hover {
  color: var(--theme-accent, #6366f1);
  opacity: 1;
  border-left-color: var(--theme-accent, #6366f1);
}
</style>
