<script setup lang="ts">
import { MdEditor, allToolbar } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { onMounted, ref } from 'vue'
import { uploadImage } from '@/api/uploads'
import { ensureMarkdownExtensions } from '@/utils/markdownExtensions'
import PostLinkToolbar from './PostLinkToolbar.vue'
import SpaceToolbar from './SpaceToolbar.vue'

const model = defineModel<string>({ required: true })
const extensionsReady = ref(false)
const toolbars = [...allToolbar] as any[]
const linkIndex = toolbars.indexOf('link')
// 在 link 按钮后插入：文章链接（slot 0）、分隔符、空格按钮（slot 1）
toolbars.splice(linkIndex >= 0 ? linkIndex + 1 : toolbars.length, 0, 0, '-', 1)

// 编辑器必须先出现：mermaid 那条依赖链接近 1MB，等它会让写作页白屏好几秒。
// 扩展加载完再把 markdownReady 打开，预览区的图表才会渲染。
onMounted(async () => {
  try {
    await ensureMarkdownExtensions()
  } catch {
    // 图表扩展加载失败不影响正常写字，正文照常可编辑
  } finally {
    extensionsReady.value = true
  }
})

const onUpload = async (files: File[], callback: (urls: string[]) => void) => {
  const urls: string[] = []
  for (const file of files) {
    const res = await uploadImage(file)
    if (res.code === 0 && res.data) {
      urls.push(res.data.url)
    }
  }
  callback(urls)
}
</script>

<template>
  <MdEditor
    v-model="model"
    language="zh-CN"
    :toolbars="toolbars"
    :on-upload-img="onUpload"
    :no-highlight="true"
    :no-prettier="true"
    style="height: 420px"
  >
    <template #defToolbars>
      <PostLinkToolbar />
      <SpaceToolbar />
    </template>
  </MdEditor>
  <div v-if="!extensionsReady" class="markdown-editor-loading">图表扩展加载中（可先正常写作，不影响保存）</div>
</template>

<style scoped>
.markdown-editor-loading {
  margin-top: 6px;
  font-size: 12px;
  color: #94a3b8;
}
</style>

<!-- 编辑器内部预览面板也受 Tailwind Preflight 影响，此处全局恢复列表标号 -->
<style>
.md-editor-preview ol {
  list-style: decimal;
  padding-left: 1.75em;
}
.md-editor-preview ul {
  list-style: disc;
  padding-left: 1.75em;
}
.md-editor-preview ol ol,
.md-editor-preview ul ol {
  list-style-type: lower-alpha;
}
.md-editor-preview ol ul,
.md-editor-preview ul ul {
  list-style-type: circle;
}
</style>
