<script setup lang="ts">
import 'md-editor-v3/lib/preview.css'
import 'katex/dist/katex.min.css'

defineProps<{ html: string }>()
</script>

<template>
  <article class="md-editor md-editor-previewOnly blog-markdown-viewer">
    <div class="md-editor-preview default-theme" v-html="html" />
  </article>
</template>

<style scoped>
.blog-markdown-viewer {
  --md-color: var(--theme-text, #3f4a54);
  --md-bk-color: transparent;
  --md-border-color: transparent;
  --md-theme-bg-color: transparent;
  --md-theme-color: var(--theme-text, #3f4a54);
  --md-theme-border-color: var(--theme-border, #e6e6e6);
  --md-theme-link-color: var(--theme-accent, #2d8cf0);
  color: var(--theme-text, #3f4a54);
  background: transparent;
  font-family: inherit;
}

.blog-markdown-viewer :deep(.md-editor-preview) {
  padding: 0;
}

.blog-markdown-viewer :deep(.katex-display) {
  margin: 1em 0;
  overflow-x: auto;
  overflow-y: hidden;
}
</style>
