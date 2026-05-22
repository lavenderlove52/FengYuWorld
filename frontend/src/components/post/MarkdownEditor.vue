<script setup lang="ts">
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { uploadImage } from '@/api/uploads'

const model = defineModel<string>({ required: true })

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
  <MdEditor v-model="model" language="zh-CN" :on-upload-img="onUpload" style="height: 420px" />
</template>
