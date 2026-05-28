<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { useGeetest } from '@/composables/useGeetest'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const message = useMessage()

const GEETEST_CAPTCHA_ID = import.meta.env.VITE_GEETEST_LOGIN_CAPTCHA_ID || ''

const form = ref({ login: '', password: '' })
const loading = ref(false)
const loginEmailError = ref('')

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const onLoginBlur = () => {
  const v = form.value.login
  if (v.includes('@') && !EMAIL_RE.test(v)) {
    loginEmailError.value = '邮箱格式不正确'
  } else {
    loginEmailError.value = ''
  }
}

const { init, showCaptcha } = useGeetest('geetest-login', GEETEST_CAPTCHA_ID)

onMounted(() => {
  init()
  const reason = route.query.reason
  if (reason === 'inactivity') {
    message.warning('长时间未操作，已自动退出登录')
  } else if (reason === 'deleted') {
    message.error('账号已不存在，请联系管理员')
  } else if (reason === 'locked') {
    message.error('账号已被锁定，请联系管理员')
  } else if (reason === 'session_expired') {
    message.warning('在其他窗口已退出登录')
  }
})

const submit = async () => {
  if (!form.value.login || !form.value.password) {
    message.warning('请填写完整信息')
    return
  }
  loading.value = true
  try {
    const gt = await showCaptcha()
    await auth.login(form.value.login, form.value.password, gt)
    message.success('登录成功')
    const raw = (route.query.redirect as string) || ''
    const redirect = raw.startsWith('/') && !raw.startsWith('//') ? raw : '/dashboard'
    router.push(redirect)
  } catch (e: any) {
    message.error(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-[#f8fafc]">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-bold text-slate-800">欢迎回来</h1>
        <p class="text-slate-400 text-sm mt-1">登录你的账号继续创作</p>
      </div>

      <div class="bg-white rounded-2xl border border-slate-100 p-8">
        <NForm @submit.prevent="submit">
          <NFormItem
            label="用户名或邮箱"
            :feedback="loginEmailError"
            :validation-status="loginEmailError ? 'error' : undefined"
          >
            <NInput
              v-model:value="form.login"
              placeholder="输入用户名或邮箱"
              size="large"
              @blur="onLoginBlur"
              @keydown.enter="submit"
            />
          </NFormItem>
          <NFormItem label="密码">
            <NInput
              v-model:value="form.password"
              type="password"
              placeholder="输入密码"
              size="large"
              show-password-on="click"
              @keydown.enter="submit"
            />
          </NFormItem>

          <!-- GeeTest 容器（bind 模式下隐藏，点提交时弹出） -->
          <div id="geetest-login" class="mb-4" />

          <NButton
            type="primary"
            size="large"
            block
            :loading="loading"
            @click="submit"
          >
            登录
          </NButton>
        </NForm>

        <p class="mt-5 text-sm text-center text-slate-400">
          还没有账号？
          <RouterLink to="/register" class="text-indigo-500 hover:text-indigo-600 font-medium">立即注册</RouterLink>
        </p>
      </div>
    </div>
  </div>
</template>
