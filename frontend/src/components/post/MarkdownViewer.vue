<script setup lang="ts">
import 'katex/dist/katex.min.css'

defineProps<{ html: string }>()
</script>

<template>
  <article class="prose max-w-none" v-html="html" />
</template>

<style scoped>
.prose :deep(h1),
.prose :deep(h2),
.prose :deep(h3) {
  font-weight: 700;
  margin-top: 1.5em;
  margin-bottom: 0.5em;
}
.prose :deep(p) {
  margin: 0.75em 0;
  line-height: 1.75;
}
.prose :deep(pre) {
  background: #1e293b;
  color: #f8fafc;
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
}
.prose :deep(code) {
  background: #f1f5f9;
  padding: 0.125rem 0.25rem;
  border-radius: 0.25rem;
}
.prose :deep(.katex-display) {
  margin: 1em 0;
  overflow-x: auto;
  overflow-y: hidden;
}
</style>
