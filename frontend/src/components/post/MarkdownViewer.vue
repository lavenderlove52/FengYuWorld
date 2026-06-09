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
  max-width: 100%;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.blog-markdown-viewer :deep(.md-editor-preview p),
.blog-markdown-viewer :deep(.md-editor-preview li),
.blog-markdown-viewer :deep(.md-editor-preview td),
.blog-markdown-viewer :deep(.md-editor-preview th) {
  color: var(--theme-text, #3f4a54);
}

.blog-markdown-viewer :deep(.md-editor-preview) {
  overflow-x: visible;
  -webkit-overflow-scrolling: touch;
}

.blog-markdown-viewer :deep(.md-editor-preview p),
.blog-markdown-viewer :deep(.md-editor-preview li) {
  max-width: 100%;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.blog-markdown-viewer :deep(.md-editor-preview .md-editor-code),
.blog-markdown-viewer :deep(.md-editor-preview pre) {
  max-width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  -webkit-overflow-scrolling: touch;
}

.blog-markdown-viewer :deep(.md-editor-preview pre code) {
  display: block;
  min-width: max-content;
  max-width: none;
  white-space: pre;
  overflow-wrap: normal;
  word-break: normal;
}

@media (max-width: 767px) {
  .blog-markdown-viewer :deep(.md-editor-preview) {
    overflow-x: hidden;
  }

  .blog-markdown-viewer :deep(.md-editor-preview p),
  .blog-markdown-viewer :deep(.md-editor-preview li),
  .blog-markdown-viewer :deep(.md-editor-preview h1),
  .blog-markdown-viewer :deep(.md-editor-preview h2),
  .blog-markdown-viewer :deep(.md-editor-preview h3),
  .blog-markdown-viewer :deep(.md-editor-preview h4),
  .blog-markdown-viewer :deep(.md-editor-preview h5),
  .blog-markdown-viewer :deep(.md-editor-preview h6) {
    white-space: normal !important;
    overflow-wrap: anywhere !important;
    word-break: break-word !important;
  }

  .blog-markdown-viewer :deep(.md-editor-preview .md-editor-code),
  .blog-markdown-viewer :deep(.md-editor-preview pre) {
    max-width: 100%;
    overflow-x: hidden;
  }

  .blog-markdown-viewer :deep(.md-editor-preview pre code),
  .blog-markdown-viewer :deep(.md-editor-preview pre code .md-editor-code-block) {
    min-width: 0 !important;
    width: 100% !important;
    max-width: 100% !important;
    white-space: pre-wrap !important;
    overflow-wrap: anywhere !important;
    word-break: break-word !important;
  }
}

.blog-markdown-viewer :deep(.md-editor-preview table) {
  display: table;
  width: max-content;
  max-width: 100%;
  margin: 1.2em 0;
  border-collapse: collapse;
  color: var(--theme-text, #3f4a54);
  background: transparent;
}

@media (max-width: 767px) {
  .blog-markdown-viewer :deep(.md-editor-preview table) {
    font-size: 0.8125rem;
  }

  .blog-markdown-viewer :deep(.md-editor-preview table th),
  .blog-markdown-viewer :deep(.md-editor-preview table td) {
    padding: 0.4rem 0.5rem;
    white-space: nowrap;
  }
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

.blog-markdown-viewer :deep(.md-editor-preview .mermaid),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid']) {
  color: #0f172a !important;
}

.blog-markdown-viewer :deep(.md-editor-preview .mermaid),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid) {
  box-sizing: border-box;
  width: 100%;
  margin: 1.5rem 0;
  padding: 1rem;
  overflow-x: auto;
  border: 1px solid #d7dee8;
  border-radius: 8px;
  background: #ffffff !important;
  -webkit-overflow-scrolling: touch;
}

.blog-markdown-viewer :deep(.md-editor-preview .mermaid svg),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid']) {
  display: block;
  max-width: 100%;
  min-width: 640px;
  height: auto;
  margin: 0 auto;
  background: #ffffff !important;
}

.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .node rect),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .node circle),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .node ellipse),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .node polygon),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .node path),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .node rect),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .node circle),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .node ellipse),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .node polygon),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .node path) {
  fill: #f8fafc !important;
  stroke: #64748b !important;
}

.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg text),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .label),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .nodeLabel),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .nodeLabel *),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .edgeLabel),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .edgeLabel *),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .cluster-label),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] text),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .label),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .nodeLabel),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .nodeLabel *),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .edgeLabel),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .edgeLabel *),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .cluster-label) {
  color: #0f172a !important;
  fill: #0f172a !important;
}

.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg foreignObject),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg foreignObject *),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .nodeLabel div),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .nodeLabel span),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .edgeLabel div),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .edgeLabel span),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] foreignObject),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] foreignObject *),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .nodeLabel div),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .nodeLabel span),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .edgeLabel div),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .edgeLabel span) {
  color: #0f172a !important;
  line-height: 1.35 !important;
}

.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .label-container),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .edgeLabel .label),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .label-container),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .edgeLabel .label) {
  color: #0f172a !important;
  background: transparent !important;
}

.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .edgeLabel),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .edgeLabel rect),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .labelBkg),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .edgeLabel),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .edgeLabel rect),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .labelBkg) {
  background-color: #ffffff !important;
  fill: #ffffff !important;
}

.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .edgePath path),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg .flowchart-link),
.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg marker path),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .edgePath path),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] .flowchart-link),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] marker path) {
  stroke: #475569 !important;
}

.blog-markdown-viewer :deep(.md-editor-preview .md-editor-mermaid svg marker path),
.blog-markdown-viewer :deep(.md-editor-preview svg[id^='mermaid'] marker path) {
  fill: #475569 !important;
}

.blog-markdown-viewer :deep(.md-editor-preview code:not(pre code)) {
  background: color-mix(in srgb, var(--theme-accent, #2d8cf0) 12%, transparent);
  color: var(--theme-accent, #2d8cf0);
}
</style>
