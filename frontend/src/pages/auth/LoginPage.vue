<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const message = useMessage()

const form = ref({ login: '', password: '' })
const loading = ref(false)

const submit = async () => {
  loading.value = true
  try {
    await auth.login(form.value.login, form.value.password)
    message.success('登录成功')
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.push(redirect)
  } catch (e: any) {
    message.error(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-md mx-auto mt-12 bg-white p-8 rounded-xl border shadow-sm">
    <h1 class="text-2xl font-bold mb-6">登录</h1>
    <NForm @submit.prevent="submit">
      <NFormItem label="用户名或邮箱">
        <NInput v-model:value="form.login" placeholder="用户名或邮箱" />
      </NFormItem>
      <NFormItem label="密码">
        <NInput v-model:value="form.password" type="password" placeholder="密码" />
      </NFormItem>
      <NButton type="primary" attr-type="submit" block :loading="loading">登录</NButton>
    </NForm>
    <p class="mt-4 text-sm text-center text-slate-500">
      还没有账号？
      <RouterLink to="/register" class="text-indigo-600">注册</RouterLink>
    </p>
  </div>
</template>
