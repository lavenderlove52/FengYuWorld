<script setup lang="ts">
import DOMPurify from 'dompurify'
import { nextTick, ref, watch } from 'vue'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/preview.css'
import { extractTocItems } from '@/utils/markdownToc'
import type { TocItem } from '@/utils/markdownToc'
import { postMobileImageURL } from '@/utils/responsiveImage'

const props = defineProps<{ contentMd: string; tocItems?: TocItem[] }>()
const rootRef = ref<HTMLElement | null>(null)

const sanitizeHtml = (html: string) => DOMPurify.sanitize(html)

const enhanceImages = () => {
  const images = Array.from(rootRef.value?.querySelectorAll<HTMLImageElement>('.md-editor-preview img') ?? [])
  for (const image of images) {
    image.loading = 'lazy'
    image.decoding = 'async'
    const mobileURL = postMobileImageURL(image.getAttribute('src') || '')
    if (!mobileURL || image.parentElement?.dataset.responsivePostImage === 'true') continue

    const picture = document.createElement('picture')
    picture.dataset.responsivePostImage = 'true'
    const source = document.createElement('source')
    source.media = '(max-width: 767px)'
    source.srcset = mobileURL
    image.before(picture)
    picture.append(source, image)
  }
}

watch(
  () => props.contentMd,
  async (content) => {
    await nextTick()
    const headings = Array.from(rootRef.value?.querySelectorAll<HTMLElement>('.md-editor-preview h1, .md-editor-preview h2, .md-editor-preview h3, .md-editor-preview h4, .md-editor-preview h5, .md-editor-preview h6') ?? [])
    const items = props.tocItems ?? extractTocItems(content)
    headings.forEach((heading, index) => {
      const item = items[index]
      if (item) heading.id = item.id
    })
    enhanceImages()
  },
  { immediate: true, flush: 'post' },
)
</script>

<template>
  <div ref="rootRef">
    <MdPreview
      :model-value="contentMd"
      language="zh-CN"
      preview-theme="default"
      class="blog-markdown-viewer"
      :sanitize="sanitizeHtml"
    />
  </div>
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
  color: var(--theme-text, #3f4a54);
  background: transparent;
}

.blog-markdown-viewer :deep(.md-editor-preview p),
.blog-markdown-viewer :deep(.md-editor-preview li),
.blog-markdown-viewer :deep(.md-editor-preview td),
.blog-markdown-viewer :deep(.md-editor-preview th) {
  color: var(--theme-text, #3f4a54);
}

.blog-markdown-viewer :deep(.md-editor-preview table) {
  display: table;
  width: max-content;
  max-width: 100%;
  margin: 1.2em 0;
  border-collapse: collapse;
  overflow-x: auto;
  color: var(--theme-text, #3f4a54);
  background: transparent;
}

.blog-markdown-viewer :deep(.md-editor-preview table tr),
.blog-markdown-viewer :deep(.md-editor-preview table tr:nth-child(2n)) {
  background: transparent;
}

.blog-markdown-viewer :deep(.md-editor-preview table th),
.blog-markdown-viewer :deep(.md-editor-preview table td) {
  border: 1px solid var(--theme-border, #e6e6e6);
  background: transparent;
}

.blog-markdown-viewer :deep(.md-editor-preview table th) {
  font-weight: 700;
  color: var(--theme-text, #3f4a54);
}

.blog-markdown-viewer :deep(.md-editor-preview blockquote) {
  margin: 1.2em 0;
  color: var(--theme-text, #3f4a54);
  background: color-mix(in srgb, var(--theme-accent, #2d8cf0) 10%, transparent);
  border-left: 4px solid var(--theme-accent, #2d8cf0);
}

.blog-markdown-viewer :deep(.md-editor-preview blockquote p) {
  color: var(--theme-text, #3f4a54);
}

.blog-markdown-viewer :deep(.md-editor-preview code:not(pre code)) {
  background: color-mix(in srgb, var(--theme-accent, #2d8cf0) 12%, transparent);
  color: var(--theme-accent, #2d8cf0);
}
</style>
