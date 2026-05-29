<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, useMessage, type FormRules, type FormInst } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { sendEmailCode } from '@/api/auth'
import { useGeetest } from '@/composables/useGeetest'

const router = useRouter()
const auth = useAuthStore()
const message = useMessage()

const GEETEST_CAPTCHA_ID = import.meta.env.VITE_GEETEST_REGISTER_CAPTCHA_ID || ''

// 步骤：1 = 邮箱+发验证码，2 = 完整表单
const step = ref(1)
const loading = ref(false)
const sending = ref(false)
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const formRef = ref<FormInst | null>(null)

const form = ref({
  email: '',
  email_code: '',
  username: '',
  nickname: '',
  password: '',
  invite_code: '',
})

const formRules: FormRules = {
  email_code: [{ required: true, message: '请填写邮箱验证码', trigger: ['blur', 'input'] }],
  username: [
    { required: true, message: '请填写用户名', trigger: ['blur', 'input'] },
    { min: 3, max: 50, message: '用户名为 3-50 个字符', trigger: ['blur', 'input'] },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '仅允许字母、数字和下划线', trigger: ['blur', 'input'] },
  ],
  nickname: [
    { required: true, message: '请填写昵称', trigger: ['blur', 'input'] },
    { min: 1, max: 80, message: '昵称不能为空且不超过 80 个字符', trigger: ['blur', 'input'] },
  ],
  password: [
    { required: true, message: '请填写密码', trigger: ['blur', 'input'] },
    { min: 6, message: '密码至少 6 位', trigger: ['blur', 'input'] },
  ],
  invite_code: [{ required: true, message: '请填写邀请码', trigger: ['blur', 'input'] }],
}

const { init, showCaptcha } = useGeetest('geetest-register', GEETEST_CAPTCHA_ID)

onMounted(() => init())

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const emailError = ref('')

const validateEmail = () => {
  if (!form.value.email) {
    emailError.value = '请填写邮箱'
    return false
  }
  if (!EMAIL_RE.test(form.value.email)) {
    emailError.value = '邮箱格式不正确'
    return false
  }
  emailError.value = ''
  return true
}

// 倒计时
const startCountdown = () => {
  countdown.value = 60
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(countdownTimer!)
      countdown.value = 0
    }
  }, 1000)
}

const sendCode = async () => {
  if (!validateEmail()) return
  if (countdown.value > 0) return
  sending.value = true
  try {
    let gt: import('@/composables/useGeetest').GeetestResult | undefined
    if (GEETEST_CAPTCHA_ID) {
      gt = await showCaptcha()
    }
    const res = await sendEmailCode(form.value.email, gt)
    if (res.code === 0) {
      message.success('验证码已发送，请查收邮件')
      step.value = 2
      startCountdown()
    } else {
      message.error(res.message || '发送失败')
    }
  } catch (e: any) {
    message.error(e.message || e.response?.data?.message || '发送失败，请稍后重试')
  } finally {
    sending.value = false
  }
}

const resendCode = async () => {
  if (countdown.value > 0 || sending.value) return
  sending.value = true
  try {
    let gt: import('@/composables/useGeetest').GeetestResult | undefined
    if (GEETEST_CAPTCHA_ID) {
      gt = await showCaptcha()
    }
    const res = await sendEmailCode(form.value.email, gt)
    if (res.code === 0) {
      message.success('验证码已重新发送')
      startCountdown()
    } else {
      message.error(res.message || '发送失败')
    }
  } catch (e: any) {
    message.error(e.message || '发送失败，请稍后重试')
  } finally {
    sending.value = false
  }
}

const submit = async () => {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  const { email, email_code, username, nickname, password, invite_code } = form.value
  loading.value = true
  try {
    const gt = await showCaptcha()
    await auth.register({ email, email_code, username, nickname, password, invite_code, ...gt })
    message.success('注册成功，欢迎加入！')
    router.push('/dashboard')
  } catch (e: any) {
    message.error(e.message || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-[#f8fafc]">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-bold text-slate-800">创建账号</h1>
        <p class="text-slate-400 text-sm mt-1">加入风雨世界，开始记录你的故事</p>
      </div>

      <!-- 步骤指示器 -->
      <div class="flex items-center justify-center gap-3 mb-6">
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold"
            :class="step >= 1 ? 'bg-indigo-500 text-white' : 'bg-slate-200 text-slate-400'">1</div>
          <span class="text-sm" :class="step >= 1 ? 'text-slate-700' : 'text-slate-400'">邮箱验证</span>
        </div>
        <div class="w-8 h-px" :class="step >= 2 ? 'bg-indigo-300' : 'bg-slate-200'" />
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold"
            :class="step >= 2 ? 'bg-indigo-500 text-white' : 'bg-slate-200 text-slate-400'">2</div>
          <span class="text-sm" :class="step >= 2 ? 'text-slate-700' : 'text-slate-400'">填写信息</span>
        </div>
      </div>

      <div class="bg-white rounded-2xl border border-slate-100 p-8">

        <!-- Step 1：邮箱 -->
        <div v-if="step === 1">
          <NForm @submit.prevent="sendCode">
            <NFormItem
              label="邮箱"
              :feedback="emailError"
              :validation-status="emailError ? 'error' : undefined"
            >
              <NInput
                v-model:value="form.email"
                placeholder="输入你的邮箱地址"
                size="large"
                type="text"
                inputmode="email"
                @blur="validateEmail"
                @keydown.enter="sendCode"
              />
            </NFormItem>
            <NButton
              type="primary"
              size="large"
              block
              :loading="sending"
              @click="sendCode"
            >
              获取验证码
            </NButton>
          </NForm>
          <p class="mt-5 text-sm text-center text-slate-400">
            已有账号？
            <RouterLink to="/login" class="text-indigo-500 hover:text-indigo-600 font-medium">直接登录</RouterLink>
          </p>
        </div>

        <!-- Step 2：完整表单 -->
        <div v-else>
          <NForm ref="formRef" :model="form" :rules="formRules" @submit.prevent="submit">

            <!-- 邮箱（只读展示） -->
            <NFormItem label="邮箱">
              <div class="flex items-center gap-2 w-full">
                <NInput
                  :value="form.email"
                  disabled
                  size="large"
                  style="flex: 1"
                />
                <button
                  type="button"
                  class="text-xs px-3 py-1.5 rounded-lg border border-slate-200 text-slate-500 hover:border-indigo-300 hover:text-indigo-500 transition flex-shrink-0 whitespace-nowrap"
                  :class="{ 'opacity-40 cursor-not-allowed': countdown > 0 || sending }"
                  @click="step = 1"
                >
                  修改
                </button>
              </div>
            </NFormItem>

            <!-- 邮箱验证码 -->
            <NFormItem label="邮箱验证码" path="email_code">
              <div class="flex items-center gap-2 w-full">
                <NInput
                  v-model:value="form.email_code"
                  placeholder="6位验证码"
                  size="large"
                  maxlength="6"
                  style="flex: 1"
                />
                <button
                  type="button"
                  class="text-xs px-3 py-1.5 rounded-lg border transition flex-shrink-0 whitespace-nowrap"
                  :class="countdown > 0 || sending
                    ? 'border-slate-200 text-slate-400 cursor-not-allowed'
                    : 'border-indigo-200 text-indigo-500 hover:bg-indigo-50'"
                  :disabled="countdown > 0 || sending"
                  @click="resendCode"
                >
                  {{ countdown > 0 ? `${countdown}s 后重发` : '重新发送' }}
                </button>
              </div>
            </NFormItem>

            <NFormItem label="用户名" path="username">
              <NInput v-model:value="form.username" placeholder="3-50 个字符，仅字母数字下划线" size="large" />
            </NFormItem>

            <NFormItem label="昵称" path="nickname">
              <NInput v-model:value="form.nickname" placeholder="显示给其他用户的名字（必填）" size="large" />
            </NFormItem>

            <NFormItem label="密码" path="password">
              <NInput
                v-model:value="form.password"
                type="password"
                placeholder="至少 6 位"
                size="large"
                show-password-on="click"
              />
            </NFormItem>

            <NFormItem label="邀请码" path="invite_code">
              <NInput v-model:value="form.invite_code" placeholder="8位邀请码，联系管理员获取" size="large" />
            </NFormItem>

            <!-- GeeTest 容器 -->
            <div id="geetest-register" class="mb-4" />

            <NButton
              type="primary"
              size="large"
              block
              :loading="loading"
              @click="submit"
            >
              完成注册
            </NButton>
          </NForm>

          <p class="mt-5 text-sm text-center text-slate-400">
            已有账号？
            <RouterLink to="/login" class="text-indigo-500 hover:text-indigo-600 font-medium">直接登录</RouterLink>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
