<script setup lang="ts">
import { ref, computed } from 'vue'
import { NForm, NFormItem, NInput, NButton, NUpload, useMessage, NModal } from 'naive-ui'
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
const showApplyModal = ref(false)
const applyingFor = ref('')

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

// 权限定义
const permissions = computed(() => [
  {
    key: 'write',
    label: '发布文章',
    desc: '创建、编辑、删除自己的文章',
    granted: true,
    applyable: false,
  },
  {
    key: 'user_pin',
    label: '个人主页置顶',
    desc: '在个人主页顶部展示置顶文章轮播',
    granted: true,
    applyable: false,
  },
  {
    key: 'friends',
    label: '好友系统',
    desc: '添加好友，分享仅好友可见的内容',
    granted: true,
    applyable: false,
  },
  {
    key: 'site_pin',
    label: '首页置顶管理',
    desc: '在全站首页轮播中添加或移除置顶文章',
    granted: auth.user?.role === 'admin' || !!auth.user?.can_site_pin,
    applyable: true,
    applyLabel: '首页置顶管理权限',
  },
  {
    key: 'custom_js',
    label: '自定义 JS',
    desc: '在个人主页和文章页注入自定义 JavaScript 脚本（鼠标特效等）',
    granted: auth.user?.role === 'admin' || !!auth.user?.can_custom_js,
    applyable: true,
    applyLabel: '自定义 JS 权限',
  },
  {
    key: 'admin',
    label: '管理员',
    desc: '管理全站用户、授权、内容等所有后台功能',
    granted: auth.user?.role === 'admin',
    applyable: false,
  },
])

const roleLabel = computed(() => {
  if (auth.user?.role === 'admin') return { text: 'Admin', cls: 'bg-amber-50 text-amber-600 border-amber-200' }
  if (auth.user?.can_site_pin) return { text: '授权用户', cls: 'bg-indigo-50 text-indigo-500 border-indigo-200' }
  return { text: '普通用户', cls: 'bg-slate-100 text-slate-500 border-slate-200' }
})

const openApply = (label: string) => {
  applyingFor.value = label
  showApplyModal.value = true
}

const copyApplyText = () => {
  const text = `你好，我是 ${auth.user?.nickname}（@${auth.user?.username}），申请获得「${applyingFor.value}」权限，请帮忙审核，谢谢！`
  navigator.clipboard.writeText(text)
  message.success('申请内容已复制')
}
</script>

<template>
  <div class="max-w-2xl space-y-10">

    <!-- 账号权限卡片 -->
    <div>
      <h2 class="text-lg font-bold text-slate-800 mb-4">账号与权限</h2>
      <div class="bg-white border border-slate-100 rounded-2xl overflow-hidden">

        <!-- 用户信息头 -->
        <div class="flex items-center gap-4 p-5 border-b border-slate-100">
          <div class="w-12 h-12 rounded-full overflow-hidden bg-slate-100 flex-shrink-0">
            <img v-if="auth.user?.avatar_url" :src="auth.user.avatar_url" class="w-full h-full object-cover" />
            <div v-else class="w-full h-full flex items-center justify-center text-lg font-bold text-slate-400">
              {{ auth.user?.nickname?.[0]?.toUpperCase() }}
            </div>
          </div>
          <div class="flex-1">
            <div class="flex items-center gap-2">
              <span class="font-semibold text-slate-800">{{ auth.user?.nickname }}</span>
              <span class="text-xs px-2 py-0.5 rounded-full border" :class="roleLabel.cls">{{ roleLabel.text }}</span>
            </div>
            <div class="text-sm text-slate-400 mt-0.5">@{{ auth.user?.username }}</div>
          </div>
        </div>

        <!-- 权限列表 -->
        <div>
          <div
            v-for="(perm, i) in permissions"
            :key="perm.key"
            class="flex items-center gap-4 px-5 py-4"
            :class="i < permissions.length - 1 ? 'border-b border-slate-50' : ''"
          >
            <!-- 状态图标 -->
            <div class="flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center"
              :class="perm.granted ? 'bg-emerald-50' : 'bg-slate-50'">
              <svg v-if="perm.granted" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                class="w-4 h-4 text-emerald-500">
                <path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z" />
              </svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
                class="w-4 h-4 text-slate-300">
                <path d="M10 1a4.5 4.5 0 0 0-4.5 4.5V9H5a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-6a2 2 0 0 0-2-2h-.5V5.5A4.5 4.5 0 0 0 10 1Zm3 8V5.5a3 3 0 1 0-6 0V9h6Z" />
              </svg>
            </div>

            <!-- 权限说明 -->
            <div class="flex-1">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium" :class="perm.granted ? 'text-slate-800' : 'text-slate-400'">
                  {{ perm.label }}
                </span>
              </div>
              <div class="text-xs mt-0.5" :class="perm.granted ? 'text-slate-400' : 'text-slate-300'">
                {{ perm.desc }}
              </div>
            </div>

            <!-- 操作 -->
            <div class="flex-shrink-0">
              <span v-if="perm.granted" class="text-xs text-emerald-500 font-medium">已拥有</span>
              <button
                v-else-if="perm.applyable"
                class="text-xs px-3 py-1.5 rounded-lg border border-slate-200 text-slate-500 hover:border-indigo-300 hover:text-indigo-500 hover:bg-indigo-50 transition"
                @click="openApply(perm.applyLabel || perm.label)"
              >申请权限</button>
              <span v-else class="text-xs text-slate-300">—</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 个人资料 -->
    <div>
      <h2 class="text-lg font-bold text-slate-800 mb-4">个人资料</h2>
      <div class="bg-white border border-slate-100 rounded-2xl p-5">
        <div class="flex items-center gap-4 mb-6">
          <div class="relative">
            <div class="w-16 h-16 rounded-full overflow-hidden bg-slate-100">
              <img v-if="form.avatar_url" :src="form.avatar_url" class="w-full h-full object-cover" />
              <div v-else class="w-full h-full flex items-center justify-center text-2xl font-bold text-slate-300">
                {{ auth.user?.nickname?.[0]?.toUpperCase() }}
              </div>
            </div>
          </div>
          <NUpload :show-file-list="false" accept="image/*" :custom-request="onAvatarUpload">
            <NButton size="small">更换头像</NButton>
          </NUpload>
        </div>
        <NForm label-placement="top">
          <NFormItem label="昵称">
            <NInput v-model:value="form.nickname" placeholder="你的昵称" />
          </NFormItem>
          <NFormItem label="个人简介">
            <NInput v-model:value="form.bio" type="textarea" :rows="3" placeholder="介绍一下自己..." />
          </NFormItem>
          <NButton type="primary" :loading="loading" @click="save">保存修改</NButton>
        </NForm>
      </div>
    </div>

    <!-- 申请权限弹窗 -->
    <NModal v-model:show="showApplyModal" preset="card" title="申请权限" style="max-width: 440px">
      <div class="space-y-4">
        <div class="flex items-start gap-3 p-3 bg-indigo-50 rounded-xl">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5 text-indigo-400 flex-shrink-0 mt-0.5">
            <path fill-rule="evenodd" d="M18 10a8 8 0 1 1-16 0 8 8 0 0 1 16 0Zm-7-4a1 1 0 1 1-2 0 1 1 0 0 1 2 0ZM9 9a.75.75 0 0 0 0 1.5h.253a.25.25 0 0 1 .244.304l-.459 2.066A1.75 1.75 0 0 0 10.747 15H11a.75.75 0 0 0 0-1.5h-.253a.25.25 0 0 1-.244-.304l.459-2.066A1.75 1.75 0 0 0 9.253 9H9Z" />
          </svg>
          <p class="text-sm text-indigo-700 leading-relaxed">
            申请「<strong>{{ applyingFor }}</strong>」权限需要管理员审核。
            复制下方申请内容，通过任意渠道发送给管理员即可。
          </p>
        </div>

        <div>
          <div class="text-xs text-slate-400 mb-2">申请内容（点击复制）</div>
          <div
            class="p-3 bg-slate-50 border border-slate-200 rounded-lg text-sm text-slate-700 leading-relaxed cursor-pointer hover:border-slate-400 transition select-all"
            @click="copyApplyText"
          >
            你好，我是 {{ auth.user?.nickname }}（@{{ auth.user?.username }}），申请获得「{{ applyingFor }}」权限，请帮忙审核，谢谢！
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-1">
          <NButton @click="showApplyModal = false">关闭</NButton>
          <NButton type="primary" @click="copyApplyText">复制申请内容</NButton>
        </div>
      </div>
    </NModal>
  </div>
</template>
