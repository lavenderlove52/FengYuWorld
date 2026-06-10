<script setup lang="ts">
import { MdEditor, allToolbar } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { onMounted, ref } from 'vue'
import { uploadImage } from '@/api/uploads'
import { ensureMarkdownExtensions } from '@/utils/markdownExtensions'
import PostLinkToolbar from './PostLinkToolbar.vue'

const model = defineModel<string>({ required: true })
const markdownReady = ref(false)
const toolbars = [...allToolbar] as any[]
const linkIndex = toolbars.indexOf('link')
toolbars.splice(linkIndex >= 0 ? linkIndex + 1 : toolbars.length, 0, 0)

onMounted(async () => {
  await ensureMarkdownExtensions()
  markdownReady.value = true
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
    v-if="markdownReady"
    v-model="model"
    language="zh-CN"
    :toolbars="toolbars"
    :on-upload-img="onUpload"
    style="height: 420px"
  >
    <template #defToolbars>
      <PostLinkToolbar />
    </template>
  </MdEditor>
  <div v-else class="markdown-editor-loading">编辑器加载中...</div>
</template>

<style scoped>
.markdown-editor-loading {
  display: grid;
  height: 420px;
  place-items: center;
  border: 1px solid #e5e7eb;
  color: #64748b;
}
</style>
