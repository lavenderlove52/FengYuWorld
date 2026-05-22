<script setup lang="ts">
import { ref } from 'vue'
import { NForm, NFormItem, NInput, NButton, NUpload, useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { uploadAvatar } from '@/api/uploads'

const auth = useAuthStore()
const message = useMessage()

const form = ref({
  nickname: auth.user?.nickname || '',
  bio: auth.user?.bio || '',
  avatar_url: auth.user?.avatar_url || '',
})
const loading = ref(false)

const save = async () => {
  loading.value = true
  try {
    await auth.updateProfile(form.value)
    message.success('资料已更新')
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

const onAvatarUpload = async ({ file, onFinish, onError }: any) => {
  try {
    const res = await uploadAvatar(file.file as File)
    if (res.code === 0 && res.data) {
      form.value.avatar_url = res.data.url
      onFinish()
    } else onError()
  } catch {
    onError()
  }
}
</script>

<template>
  <div class="max-w-lg">
    <h1 class="text-2xl font-bold mb-6">个人设置</h1>
    <div class="mb-4 flex items-center gap-4">
      <img
        v-if="form.avatar_url"
        :src="form.avatar_url"
        class="w-16 h-16 rounded-full object-cover border"
      />
      <NUpload :show-file-list="false" accept="image/*" :custom-request="onAvatarUpload">
        <NButton size="small">上传头像</NButton>
      </NUpload>
    </div>
    <NForm label-placement="top">
      <NFormItem label="昵称">
        <NInput v-model:value="form.nickname" />
      </NFormItem>
      <NFormItem label="简介">
        <NInput v-model:value="form.bio" type="textarea" :rows="4" />
      </NFormItem>
      <NButton type="primary" :loading="loading" @click="save">保存</NButton>
    </NForm>
  </div>
</template>
