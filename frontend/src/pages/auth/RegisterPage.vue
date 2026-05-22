<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()
const message = useMessage()

const form = ref({ username: '', email: '', password: '', nickname: '' })
const loading = ref(false)

const submit = async () => {
  loading.value = true
  try {
    await auth.register(form.value)
    message.success('注册成功')
    router.push('/dashboard')
  } catch (e: any) {
    message.error(e.message || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-md mx-auto mt-12 bg-white p-8 rounded-xl border shadow-sm">
    <h1 class="text-2xl font-bold mb-6">注册</h1>
    <NForm @submit.prevent="submit">
      <NFormItem label="用户名">
        <NInput v-model:value="form.username" placeholder="3-50 字符" />
      </NFormItem>
      <NFormItem label="邮箱">
        <NInput v-model:value="form.email" placeholder="email@example.com" />
      </NFormItem>
      <NFormItem label="昵称">
        <NInput v-model:value="form.nickname" placeholder="显示名称" />
      </NFormItem>
      <NFormItem label="密码">
        <NInput v-model:value="form.password" type="password" placeholder="至少 6 位" />
      </NFormItem>
      <NButton type="primary" attr-type="submit" block :loading="loading">注册</NButton>
    </NForm>
    <p class="mt-4 text-sm text-center text-slate-500">
      已有账号？
      <RouterLink to="/login" class="text-indigo-600">登录</RouterLink>
    </p>
  </div>
</template>
