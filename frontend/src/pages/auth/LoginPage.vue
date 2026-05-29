<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { useCaptcha } from '@/composables/useCaptcha'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const message = useMessage()

const form = ref({ login: '', password: '' })
const captchaAnswer = ref('')
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

const { captchaId, captchaImage, loading: captchaLoading, refresh: refreshCaptcha } = useCaptcha()

onMounted(async () => {
  await refreshCaptcha()
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
  if (!captchaAnswer.value) {
    message.warning('请输入验证码')
    return
  }
  loading.value = true
  try {
    await auth.login(form.value.login, form.value.password, captchaId.value, captchaAnswer.value)
    message.success('登录成功')
    const raw = (route.query.redirect as string) || ''
    const redirect = raw.startsWith('/') && !raw.startsWith('//') ? raw : '/dashboard'
    router.push(redirect)
  } catch (e: any) {
    message.error(e.message || '登录失败')
    captchaAnswer.value = ''
    await refreshCaptcha()
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

          <!-- 图形验证码 -->
          <NFormItem label="验证码">
            <div class="flex items-center gap-3 w-full">
              <NInput
                v-model:value="captchaAnswer"
                placeholder="输入图中字符"
                size="large"
                style="flex: 1"
                @keydown.enter="submit"
              />
              <div
                class="flex-shrink-0 cursor-pointer rounded-lg overflow-hidden border border-slate-200 hover:border-indigo-300 transition"
                style="height: 40px; width: 120px"
                :title="captchaLoading ? '加载中...' : '点击刷新验证码'"
                @click="refreshCaptcha"
              >
                <img
                  v-if="captchaImage"
                  :src="captchaImage"
                  alt="验证码"
                  class="w-full h-full object-cover"
                />
                <div v-else class="w-full h-full flex items-center justify-center text-xs text-slate-400">
                  加载中...
                </div>
              </div>
            </div>
          </NFormItem>

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
