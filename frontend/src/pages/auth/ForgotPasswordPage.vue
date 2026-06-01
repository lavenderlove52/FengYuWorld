<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { sendEmailCode, resetPassword } from '@/api/auth'
import { useCaptcha } from '@/composables/useCaptcha'

const router = useRouter()
const message = useMessage()

// 1 = 邮箱 + 图形验证码，2 = 邮箱验证码，3 = 新密码
const step = ref(1)
const loading = ref(false)
const sending = ref(false)
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const email = ref('')
const emailCode = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

// Step 1 填写的图形验证码答案，在 Step 3 提交时复用
const savedCaptchaId = ref('')
const savedCaptchaAnswer = ref('')
const captchaAnswerInput = ref('')

const emailError = ref('')
const captchaError = ref('')
const passwordError = ref('')

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const { captchaId, captchaImage, loading: captchaLoading, refresh: refreshCaptcha } = useCaptcha()

onMounted(() => refreshCaptcha())

// 邮箱脱敏：local 前2位 + *** + @ + domain
const maskedEmail = () => {
  const at = email.value.indexOf('@')
  if (at <= 0) return email.value
  const local = email.value.slice(0, at)
  const domain = email.value.slice(at)
  const visible = local.slice(0, Math.min(2, local.length))
  return `${visible}***${domain}`
}

// 倒计时
const startCountdown = () => {
  countdown.value = 60
  if (countdownTimer) clearInterval(countdownTimer)
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(countdownTimer!)
      countdown.value = 0
    }
  }, 1000)
}

// Step 1：发送邮箱验证码
const sendCode = async () => {
  emailError.value = ''
  captchaError.value = ''

  if (!email.value) {
    emailError.value = '请填写邮箱'
    return
  }
  if (!EMAIL_RE.test(email.value)) {
    emailError.value = '邮箱格式不正确'
    return
  }
  if (!captchaAnswerInput.value) {
    captchaError.value = '请输入图形验证码'
    return
  }

  sending.value = true
  try {
    const res = await sendEmailCode(email.value)
    if (res.code === 0) {
      // 保存图形验证码，Step 3 提交时使用
      savedCaptchaId.value = captchaId.value
      savedCaptchaAnswer.value = captchaAnswerInput.value
      message.success('验证码已发送，请查收邮件')
      step.value = 2
      startCountdown()
    } else {
      message.error(res.message || '发送失败')
      captchaAnswerInput.value = ''
      await refreshCaptcha()
    }
  } catch (e: any) {
    message.error(e.message || '发送失败，请稍后重试')
    captchaAnswerInput.value = ''
    await refreshCaptcha()
  } finally {
    sending.value = false
  }
}

// Step 2：重新发送
const resendCode = async () => {
  if (countdown.value > 0 || sending.value) return
  sending.value = true
  try {
    const res = await sendEmailCode(email.value)
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

// Step 2：下一步
const goToStep3 = () => {
  if (!emailCode.value) {
    message.warning('请输入邮箱验证码')
    return
  }
  step.value = 3
}

// Step 3：提交重置
const submit = async () => {
  passwordError.value = ''
  if (!newPassword.value || newPassword.value.length < 6) {
    passwordError.value = '密码至少 6 位'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = '两次输入的密码不一致'
    return
  }

  loading.value = true
  try {
    const res = await resetPassword({
      email: email.value,
      email_code: emailCode.value,
      new_password: newPassword.value,
      captcha_id: savedCaptchaId.value,
      captcha_answer: savedCaptchaAnswer.value,
    })
    if (res.code === 0) {
      message.success('密码已重置，请重新登录')
      router.push('/login')
    } else {
      message.error(res.message || '重置失败')
      // 若验证码相关错误，回到 Step 1
      if (res.message?.includes('验证码')) {
        step.value = 1
        captchaAnswerInput.value = ''
        emailCode.value = ''
        await refreshCaptcha()
      }
    }
  } catch (e: any) {
    const msg: string = e.message || '重置失败'
    message.error(msg)
    if (msg.includes('验证码') || msg.includes('未注册')) {
      step.value = 1
      captchaAnswerInput.value = ''
      emailCode.value = ''
      await refreshCaptcha()
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-[#f8fafc]">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-bold text-slate-800">重置密码</h1>
        <p class="text-slate-400 text-sm mt-1">通过邮箱验证码找回你的账号</p>
      </div>

      <!-- 步骤指示器 -->
      <div class="flex items-center justify-center gap-2 mb-6">
        <div class="flex items-center gap-2">
          <div
            class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold"
            :class="step >= 1 ? 'bg-indigo-500 text-white' : 'bg-slate-200 text-slate-400'"
          >1</div>
          <span class="text-sm" :class="step >= 1 ? 'text-slate-700' : 'text-slate-400'">验证邮箱</span>
        </div>
        <div class="w-8 h-px" :class="step >= 2 ? 'bg-indigo-300' : 'bg-slate-200'" />
        <div class="flex items-center gap-2">
          <div
            class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold"
            :class="step >= 2 ? 'bg-indigo-500 text-white' : 'bg-slate-200 text-slate-400'"
          >2</div>
          <span class="text-sm" :class="step >= 2 ? 'text-slate-700' : 'text-slate-400'">输入验证码</span>
        </div>
        <div class="w-8 h-px" :class="step >= 3 ? 'bg-indigo-300' : 'bg-slate-200'" />
        <div class="flex items-center gap-2">
          <div
            class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold"
            :class="step >= 3 ? 'bg-indigo-500 text-white' : 'bg-slate-200 text-slate-400'"
          >3</div>
          <span class="text-sm" :class="step >= 3 ? 'text-slate-700' : 'text-slate-400'">设置新密码</span>
        </div>
      </div>

      <div class="bg-white rounded-2xl border border-slate-100 p-8">

        <!-- Step 1：邮箱 + 图形验证码 -->
        <div v-if="step === 1">
          <NForm @submit.prevent="sendCode">
            <NFormItem
              label="注册邮箱"
              :feedback="emailError"
              :validation-status="emailError ? 'error' : undefined"
            >
              <NInput
                v-model:value="email"
                placeholder="请输入你的注册邮箱"
                size="large"
                type="text"
                inputmode="email"
                @blur="() => { if (email && !EMAIL_RE.test(email)) emailError = '邮箱格式不正确'; else emailError = '' }"
                @keydown.enter="sendCode"
              />
            </NFormItem>

            <!-- 图形验证码 -->
            <NFormItem
              label="图形验证码"
              :feedback="captchaError"
              :validation-status="captchaError ? 'error' : undefined"
            >
              <div class="flex items-center gap-3 w-full">
                <NInput
                  v-model:value="captchaAnswerInput"
                  placeholder="输入图中字符"
                  size="large"
                  style="flex: 1"
                  @keydown.enter="sendCode"
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
              :loading="sending"
              @click="sendCode"
            >
              获取邮箱验证码
            </NButton>
          </NForm>

          <p class="mt-5 text-sm text-center text-slate-400">
            想起密码了？
            <RouterLink to="/login" class="text-indigo-500 hover:text-indigo-600 font-medium">直接登录</RouterLink>
          </p>
        </div>

        <!-- Step 2：输入邮箱验证码 -->
        <div v-else-if="step === 2">
          <p class="text-sm text-slate-500 mb-5 text-center">
            验证码已发送至
            <span class="font-medium text-slate-700">{{ maskedEmail() }}</span>
          </p>

          <NForm @submit.prevent="goToStep3">
            <NFormItem label="邮箱验证码">
              <div class="flex items-center gap-2 w-full">
                <NInput
                  v-model:value="emailCode"
                  placeholder="请输入 6 位验证码"
                  size="large"
                  maxlength="6"
                  style="flex: 1"
                  @keydown.enter="goToStep3"
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

            <NButton
              type="primary"
              size="large"
              block
              :disabled="!emailCode"
              @click="goToStep3"
            >
              下一步
            </NButton>
          </NForm>

          <p class="mt-4 text-center">
            <button
              type="button"
              class="text-sm text-slate-400 hover:text-indigo-500 transition-colors"
              @click="step = 1"
            >
              ← 返回修改邮箱
            </button>
          </p>
        </div>

        <!-- Step 3：设置新密码 -->
        <div v-else>
          <NForm @submit.prevent="submit">
            <NFormItem
              label="新密码"
              :feedback="passwordError"
              :validation-status="passwordError ? 'error' : undefined"
            >
              <NInput
                v-model:value="newPassword"
                type="password"
                placeholder="至少 6 位"
                size="large"
                show-password-on="click"
                @keydown.enter="submit"
              />
            </NFormItem>

            <NFormItem label="确认密码">
              <NInput
                v-model:value="confirmPassword"
                type="password"
                placeholder="再次输入新密码"
                size="large"
                show-password-on="click"
                @keydown.enter="submit"
              />
            </NFormItem>

            <NButton
              type="primary"
              size="large"
              block
              :loading="loading"
              @click="submit"
            >
              重置密码
            </NButton>
          </NForm>

          <p class="mt-4 text-center">
            <button
              type="button"
              class="text-sm text-slate-400 hover:text-indigo-500 transition-colors"
              @click="step = 2"
            >
              ← 返回输入验证码
            </button>
          </p>
        </div>

      </div>
    </div>
  </div>
</template>
