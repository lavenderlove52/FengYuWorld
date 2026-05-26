<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { listPublicPosts } from '@/api/posts'
import { getSitePins, addSitePin, removeSitePin, grantSitePinPermission, revokeSitePinPermission } from '@/api/pins'
import { listUsers } from '@/api/users'
import type { PostSummary, PinnedPost } from '@/types/post'
import type { UserInfo } from '@/types/auth'

const message = useMessage()
const activeTab = ref<'pins' | 'grant'>('pins')

// ── 置顶管理 ──────────────────────────────────────────
const pins = ref<PinnedPost[]>([])
const allPosts = ref<PostSummary[]>([])
const loadingPins = ref(false)
const loadingPosts = ref(false)
const operatingPost = ref<number | null>(null)
const postKeyword = ref('')

const pinnedPostIds = computed(() => new Set(pins.value.map((p) => p.post_id)))

const filteredPosts = computed(() => {
  const kw = postKeyword.value.trim().toLowerCase()
  if (!kw) return allPosts.value
  return allPosts.value.filter(
    (p) => p.title.toLowerCase().includes(kw) || p.author?.nickname?.toLowerCase().includes(kw),
  )
})

const loadPins = async () => {
  loadingPins.value = true
  loadingPosts.value = true
  try {
    const [pinsRes, postsRes] = await Promise.all([getSitePins(), listPublicPosts(1, 100)])
    if (pinsRes.code === 0 && pinsRes.data) pins.value = pinsRes.data
    if (postsRes.code === 0 && postsRes.data) allPosts.value = postsRes.data.items
  } finally {
    loadingPins.value = false
    loadingPosts.value = false
  }
}

const togglePin = async (post: PostSummary) => {
  operatingPost.value = post.id
  try {
    if (pinnedPostIds.value.has(post.id)) {
      const res = await removeSitePin(post.id)
      if (res.code === 0) {
        pins.value = pins.value.filter((p) => p.post_id !== post.id)
        message.success('已取消首页置顶')
      } else {
        message.error(res.message || '操作失败')
      }
    } else {
      if (pins.value.length >= 20) {
        message.warning('首页置顶最多 20 条，请先移除部分置顶')
        return
      }
      const res = await addSitePin(post.id)
      if (res.code === 0) {
        await loadPins()
        message.success('已置顶到首页')
      } else {
        message.error(res.message || '操作失败')
      }
    }
  } finally {
    operatingPost.value = null
  }
}

// ── 用户授权 ──────────────────────────────────────────
const users = ref<UserInfo[]>([])
const userTotal = ref(0)
const loadingUsers = ref(false)
const operatingUser = ref<number | null>(null)
const userKeyword = ref('')
let userSearchTimer: ReturnType<typeof setTimeout> | null = null

const loadUsers = async (kw = '') => {
  loadingUsers.value = true
  try {
    const res = await listUsers(kw, 1, 30)
    if (res.code === 0 && res.data) {
      users.value = res.data.items
      userTotal.value = res.data.total
    }
  } finally {
    loadingUsers.value = false
  }
}

const onUserSearch = () => {
  if (userSearchTimer) clearTimeout(userSearchTimer)
  userSearchTimer = setTimeout(() => loadUsers(userKeyword.value), 300)
}

const toggleGrant = async (user: UserInfo) => {
  operatingUser.value = user.id
  try {
    if (user.can_site_pin) {
      const res = await revokeSitePinPermission(user.id)
      if (res.code === 0) {
        user.can_site_pin = false
        message.success(`已撤销 ${user.nickname} 的首页置顶权限`)
      } else {
        message.error(res.message || '操作失败')
      }
    } else {
      const res = await grantSitePinPermission(user.id)
      if (res.code === 0) {
        user.can_site_pin = true
        message.success(`已授权 ${user.nickname} 管理首页置顶`)
      } else {
        message.error(res.message || '操作失败')
      }
    }
  } finally {
    operatingUser.value = null
  }
}

onMounted(() => {
  loadPins()
  loadUsers()
})

const formatDate = (d: string) => d?.slice(0, 10) ?? ''
</script>

<template>
  <div>
    <!-- 页头 -->
    <div class="flex items-start justify-between mb-6">
      <div>
        <div class="flex items-center gap-2 mb-1">
          <span class="text-xs font-semibold tracking-widest text-slate-400 uppercase">Admin</span>
        </div>
        <h1 class="text-2xl font-bold text-slate-800">首页置顶管理</h1>
        <p class="text-sm text-slate-400 mt-1">管理首页轮播内容及授权其他用户操作权限</p>
      </div>
    </div>

    <!-- Tab 切换 -->
    <div class="flex gap-1 mb-8 border-b border-slate-100">
      <button
        v-for="tab in [{ key: 'pins', label: '置顶内容' }, { key: 'grant', label: '权限授权' }]"
        :key="tab.key"
        class="px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors"
        :class="activeTab === tab.key
          ? 'border-slate-800 text-slate-800'
          : 'border-transparent text-slate-400 hover:text-slate-600'"
        @click="activeTab = (tab.key as 'pins' | 'grant')"
      >{{ tab.label }}</button>
    </div>

    <!-- Tab: 置顶内容 -->
    <div v-if="activeTab === 'pins'" class="grid grid-cols-5 gap-8">
      <!-- 左：当前置顶 -->
      <div class="col-span-2">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-base font-semibold text-slate-700">当前置顶</h2>
          <span
            class="text-xs px-2 py-0.5 rounded-full"
            :class="pins.length >= 20 ? 'bg-red-50 text-red-500' : 'bg-slate-100 text-slate-500'"
          >{{ pins.length }} / 20</span>
        </div>
        <div v-if="loadingPins" class="text-center py-8 text-slate-400 text-sm">加载中...</div>
        <div v-else-if="!pins.length" class="text-center py-12 text-slate-300">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-10 h-10 mx-auto mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11.48 3.499a.562.562 0 0 1 1.04 0l2.125 5.111a.563.563 0 0 0 .475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 0 0-.182.557l1.285 5.385a.562.562 0 0 1-.84.61l-4.725-2.885a.562.562 0 0 0-.586 0L6.982 20.54a.562.562 0 0 1-.84-.61l1.285-5.386a.562.562 0 0 0-.182-.557l-4.204-3.602a.562.562 0 0 1 .321-.988l5.518-.442a.563.563 0 0 0 .475-.345L11.48 3.5Z" />
          </svg>
          <p class="text-sm">暂无置顶文章</p>
          <p class="text-xs mt-1 text-slate-200">从右侧添加</p>
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="(pin, index) in pins"
            :key="pin.id"
            class="group flex items-start gap-3 p-3 bg-white border border-slate-100 rounded-xl hover:border-slate-200 transition"
          >
            <span class="text-lg font-bold text-slate-200 w-5 flex-shrink-0 mt-0.5 text-right leading-none">{{ index + 1 }}</span>
            <div class="flex-1 min-w-0">
              <div class="text-sm font-semibold text-slate-800 leading-snug line-clamp-1">{{ pin.post.title }}</div>
              <div class="text-xs text-slate-400 mt-1">{{ pin.post.author?.nickname }} · {{ formatDate(pin.post.published_at || pin.post.created_at) }}</div>
            </div>
            <button
              class="flex-shrink-0 opacity-0 group-hover:opacity-100 transition p-1 text-slate-300 hover:text-red-500"
              :disabled="operatingPost === pin.post_id"
              @click="togglePin(pin.post)"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
                <path d="M6.28 5.22a.75.75 0 0 0-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 1 0 1.06 1.06L10 11.06l3.72 3.72a.75.75 0 1 0 1.06-1.06L11.06 10l3.72-3.72a.75.75 0 0 0-1.06-1.06L10 8.94 6.28 5.22Z" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- 右：全站文章 -->
      <div class="col-span-3 border-l border-slate-100 pl-8">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-base font-semibold text-slate-700">全站公开文章</h2>
          <span class="text-xs text-slate-400">点击「+ 置顶」添加到首页</span>
        </div>
        <div class="relative mb-4">
          <svg xmlns="http://www.w3.org/2000/svg" class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
          </svg>
          <input v-model="postKeyword" type="text" placeholder="搜索文章标题或作者..."
            class="w-full pl-9 pr-4 py-2 text-sm border border-slate-200 rounded-lg bg-white focus:outline-none focus:border-slate-400 transition" />
        </div>
        <div v-if="loadingPosts" class="text-center py-8 text-slate-400 text-sm">加载中...</div>
        <div v-else-if="!filteredPosts.length" class="text-center py-8 text-slate-300 text-sm">没有找到匹配的文章</div>
        <div v-else class="space-y-1 max-h-[560px] overflow-y-auto pr-1">
          <div v-for="post in filteredPosts" :key="post.id"
            class="flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
            <div class="w-10 h-10 rounded-lg overflow-hidden bg-slate-100 flex-shrink-0">
              <img v-if="post.cover_image_url" :src="post.cover_image_url" class="w-full h-full object-cover" />
              <div v-else class="w-full h-full flex items-center justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="m2.25 15.75 5.159-5.159a2.25 2.25 0 0 1 3.182 0l5.159 5.159m-1.5-1.5 1.409-1.409a2.25 2.25 0 0 1 3.182 0l2.909 2.909m-18 3.75h16.5a1.5 1.5 0 0 0 1.5-1.5V6a1.5 1.5 0 0 0-1.5-1.5H3.75A1.5 1.5 0 0 0 2.25 6v12a1.5 1.5 0 0 0 1.5 1.5Zm10.5-11.25h.008v.008h-.008V8.25Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />
                </svg>
              </div>
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-sm font-medium text-slate-700 truncate">{{ post.title }}</div>
              <div class="text-xs text-slate-400 mt-0.5">{{ post.author?.nickname }} · {{ formatDate(post.published_at || post.created_at) }} · {{ post.view_count }} 阅读</div>
            </div>
            <button
              class="flex-shrink-0 text-xs px-3 py-1.5 rounded-lg border transition min-w-[60px] text-center"
              :class="pinnedPostIds.has(post.id)
                ? 'bg-indigo-50 text-indigo-500 border-indigo-100 hover:bg-red-50 hover:text-red-500 hover:border-red-100'
                : 'bg-white text-slate-500 border-slate-200 hover:bg-indigo-50 hover:text-indigo-600 hover:border-indigo-200'"
              :disabled="operatingPost === post.id"
              @click="togglePin(post)"
            >
              <span v-if="operatingPost === post.id">...</span>
              <span v-else-if="pinnedPostIds.has(post.id)">已置顶</span>
              <span v-else>+ 置顶</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab: 权限授权 -->
    <div v-else-if="activeTab === 'grant'">
      <div class="max-w-2xl">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h2 class="text-base font-semibold text-slate-700">首页置顶权限授权</h2>
            <p class="text-xs text-slate-400 mt-1">授权后，该用户可在仪表盘操作首页置顶，admin 始终拥有此权限</p>
          </div>
        </div>

        <!-- 搜索用户 -->
        <div class="relative mb-5">
          <svg xmlns="http://www.w3.org/2000/svg" class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
          </svg>
          <input v-model="userKeyword" type="text" placeholder="搜索用户名或昵称..."
            class="w-full pl-9 pr-4 py-2 text-sm border border-slate-200 rounded-lg bg-white focus:outline-none focus:border-slate-400 transition"
            @input="onUserSearch" />
        </div>

        <div v-if="loadingUsers" class="text-center py-8 text-slate-400 text-sm">加载中...</div>
        <div v-else-if="!users.length" class="text-center py-10 text-slate-300 text-sm">没有找到用户</div>
        <div v-else class="space-y-2">
          <div
            v-for="user in users"
            :key="user.id"
            class="flex items-center gap-4 px-4 py-3 bg-white border border-slate-100 rounded-xl hover:border-slate-200 transition"
          >
            <!-- 头像 -->
            <div class="w-9 h-9 rounded-full overflow-hidden bg-slate-100 flex-shrink-0">
              <img v-if="user.avatar_url" :src="user.avatar_url" class="w-full h-full object-cover" />
              <div v-else class="w-full h-full flex items-center justify-center text-sm font-bold text-slate-400">
                {{ user.nickname?.[0]?.toUpperCase() }}
              </div>
            </div>

            <!-- 用户信息 -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <span class="text-sm font-semibold text-slate-800">{{ user.nickname }}</span>
                <span v-if="user.role === 'admin'"
                  class="text-[10px] px-1.5 py-0.5 bg-amber-50 text-amber-600 border border-amber-100 rounded-full">admin</span>
                <span v-else-if="user.can_site_pin"
                  class="text-[10px] px-1.5 py-0.5 bg-indigo-50 text-indigo-500 border border-indigo-100 rounded-full">已授权</span>
              </div>
              <div class="text-xs text-slate-400 mt-0.5">@{{ user.username }}</div>
            </div>

            <!-- 授权按钮（admin 不需要授权） -->
            <div v-if="user.role === 'admin'" class="text-xs text-slate-300">内置权限</div>
            <button
              v-else
              class="flex-shrink-0 text-xs px-3 py-1.5 rounded-lg border transition min-w-[72px] text-center"
              :class="user.can_site_pin
                ? 'bg-indigo-50 text-indigo-500 border-indigo-100 hover:bg-red-50 hover:text-red-500 hover:border-red-100'
                : 'bg-white text-slate-500 border-slate-200 hover:bg-indigo-50 hover:text-indigo-600 hover:border-indigo-200'"
              :disabled="operatingUser === user.id"
              @click="toggleGrant(user)"
            >
              <span v-if="operatingUser === user.id">...</span>
              <span v-else-if="user.can_site_pin">撤销授权</span>
              <span v-else>授予权限</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
