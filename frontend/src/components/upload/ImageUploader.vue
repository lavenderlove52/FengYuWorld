<script setup lang="ts">
import { NUpload, NButton } from 'naive-ui'
import { uploadImage } from '@/api/uploads'

const model = defineModel<string>()

const customRequest = async ({ file, onFinish, onError }: any) => {
  try {
    const res = await uploadImage(file.file as File)
    if (res.code === 0 && res.data) {
      model.value = res.data.url
      onFinish()
    } else {
      onError()
    }
  } catch {
    onError()
  }
}
</script>

<template>
  <div class="space-y-2">
    <img v-if="model" :src="model" alt="cover" class="max-h-40 rounded-lg border" />
    <NUpload :show-file-list="false" accept="image/*" :custom-request="customRequest">
      <NButton size="small">上传封面</NButton>
    </NUpload>
  </div>
</template>
