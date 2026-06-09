<script setup lang="ts">
import { MdEditor, allToolbar } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { uploadImage } from '@/api/uploads'
import PostLinkToolbar from './PostLinkToolbar.vue'

const model = defineModel<string>({ required: true })
const toolbars = [...allToolbar] as any[]
const linkIndex = toolbars.indexOf('link')
toolbars.splice(linkIndex >= 0 ? linkIndex + 1 : toolbars.length, 0, 0)

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
    style="height: 420px"
  >
    <template #defToolbars>
      <PostLinkToolbar />
    </template>
  </MdEditor>
</template>
