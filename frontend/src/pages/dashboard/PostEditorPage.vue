<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, NSpace, NSelect, NSwitch, NCollapse, NCollapseItem, useMessage } from 'naive-ui'
import { createPost, getPost, updatePost } from '@/api/posts'
import { listCategories, createCategory } from '@/api/categories'
import type { Category } from '@/types/category'
import type { CreatePostPayload, PostVisibility } from '@/types/post'
import MarkdownEditor from '@/components/post/MarkdownEditor.vue'
import VisibilitySelect from '@/components/post/VisibilitySelect.vue'
import FriendMultiSelect from '@/components/post/FriendMultiSelect.vue'
import ImageUploader from '@/components/upload/ImageUploader.vue'
import { validatePostSlug } from '@/utils/postSlug'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const isEdit = computed(() => !!route.params.id)
const postId = computed(() => Number(route.params.id))

const form = ref<CreatePostPayload>({
  title: '',
  slug: '',
  cover_image_url: '',
  summary: '',
  content_md: '# 标题\n\n开始写作...',
  tag_names: [],
  status: 'draft',
  visibility: 'private',
  visible_user_ids: [],
  allow_comment: true,
  show_toc: true,
})

const tagInput = ref('')
const categories = ref<Category[]>([])
const loading = ref(false)

// 新建分类
const showNewCat = ref(false)
const newCatName = ref('')
const addingCat = ref(false)

const openNewCat = () => {
  newCatName.value = ''
  showNewCat.value = true
}

const cancelNewCat = () => {
  showNewCat.value = false
  newCatName.value = ''
}

const submitNewCat = async () => {
  const name = newCatName.value.trim()
  if (!name) {
    message.warning('分类名称不能为空')
    return
  }
  addingCat.value = true
  try {
    const res = await createCategory({ name })
    if (res.code === 0 && res.data) {
      categories.value.push(res.data)
      form.value.category_id = res.data.id
      message.success(`分类「${res.data.name}」已创建`)
      showNewCat.value = false
      newCatName.value = ''
    } else {
      message.error(res.message || '创建失败')
    }
  } finally {
    addingCat.value = false
  }
}

const categoryOptions = computed(() => [
  { label: '无分类', value: undefined as number | undefined },
  ...categories.value.map((c) => ({ label: c.name, value: c.id as number | undefined })),
])
const slugWarning = computed(() => validatePostSlug(form.value.slug || ''))

onMounted(async () => {
  const catRes = await listCategories()
  if (catRes.code === 0 && catRes.data) categories.value = catRes.data

  if (isEdit.value) {
    const res = await getPost(postId.value)
    if (res.code === 0 && res.data) {
      const p = res.data
      form.value = {
        title: p.title,
        slug: p.slug,
        cover_image_url: p.cover_image_url,
        summary: p.summary || '',
        content_md: p.content_md || '',
        category_id: p.category_id,
        tag_names: p.tags || [],
        status: p.status as any,
        visibility: p.visibility as PostVisibility,
        visible_user_ids: p.visible_user_ids || [],
        allow_comment: p.allow_comment,
        show_toc: p.show_toc !== false,
      }
      tagInput.value = (p.tags || []).join(', ')
    }
  }
})

const buildPayload = (status: 'draft' | 'published'): CreatePostPayload => {
  const payload: CreatePostPayload = {
    ...form.value,
    status,
    tag_names: tagInput.value
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean),
  }
  // 未填 slug 则不传，由后端自动生成
  if (!payload.slug?.trim()) {
    delete payload.slug
  }
  return payload
}

const validate = () => {
  if (!form.value.title.trim()) {
    message.warning('标题不能为空')
    return false
  }
  if (!form.value.content_md.trim()) {
    message.warning('正文不能为空')
    return false
  }
  if (form.value.visibility === 'specified' && !form.value.visible_user_ids?.length) {
    message.warning('指定好友可见时至少选择一位好友')
    return false
  }
  const slugErr = validatePostSlug(form.value.slug || '')
  if (slugErr) {
    message.warning(slugErr)
    return false
  }
  return true
}

const save = async (status: 'draft' | 'published') => {
  if (!validate()) return
  loading.value = true
  try {
    const payload = buildPayload(status)
    const res = isEdit.value
      ? await updatePost(postId.value, payload)
      : await createPost(payload)
    if (res.code === 0) {
      message.success(status === 'draft' ? '草稿已保存' : '文章已发布')
      router.push('/dashboard/posts')
    } else {
      message.error(res.message)
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">{{ isEdit ? '编辑文章' : '写文章' }}</h1>
    <NForm label-placement="top">
      <NFormItem label="标题">
        <NInput v-model:value="form.title" placeholder="文章标题" />
      </NFormItem>
      <NFormItem label="封面">
        <ImageUploader v-model="form.cover_image_url" />
      </NFormItem>
      <NFormItem label="文章简介">
        <NInput
          v-model:value="form.summary"
          type="textarea"
          :rows="3"
          :maxlength="500"
          show-count
          placeholder="可选；留空时首页和列表不会展示简介"
        />
      </NFormItem>
      <NFormItem label="分类">
        <div class="w-full flex items-center gap-2">
          <NSelect
            v-model:value="form.category_id"
            :options="categoryOptions"
            placeholder="无分类"
            clearable
            class="flex-1"
          />
          <button
            type="button"
            title="新建分类"
            class="w-8 h-8 flex-shrink-0 flex items-center justify-center rounded border border-dashed border-slate-400 text-slate-500 hover:border-indigo-500 hover:text-indigo-500 transition-colors text-lg leading-none"
            @click="openNewCat"
          >+</button>
        </div>
        <!-- 内联新建分类 -->
        <div v-if="showNewCat" class="flex items-center gap-2 mt-2 w-full">
          <NInput
            v-model:value="newCatName"
            placeholder="输入分类名称"
            size="small"
            class="flex-1"
            @keyup.enter="submitNewCat"
            @keyup.esc="cancelNewCat"
          />
          <NButton size="small" type="primary" :loading="addingCat" @click="submitNewCat">确认</NButton>
          <NButton size="small" @click="cancelNewCat">取消</NButton>
        </div>
      </NFormItem>
      <NFormItem label="标签（逗号分隔）">
        <NInput v-model:value="tagInput" placeholder="生活, 随笔" />
      </NFormItem>
      <NFormItem label="可见性">
        <VisibilitySelect v-model="form.visibility" />
      </NFormItem>
      <NFormItem v-if="form.visibility === 'specified'" label="指定好友">
        <FriendMultiSelect v-model="form.visible_user_ids!" />
      </NFormItem>
      <NFormItem label="文章目录">
        <div class="flex items-center gap-3">
          <NSwitch v-model:value="form.show_toc" />
          <span class="text-sm text-slate-500">开启后，详情页会自动识别 Markdown 标题并生成目录</span>
        </div>
      </NFormItem>
      <NFormItem label="正文">
        <MarkdownEditor v-model="form.content_md" />
      </NFormItem>
      <!-- 高级选项（折叠） -->
      <NCollapse class="mb-4">
        <NCollapseItem title="高级选项" name="advanced">
          <NFormItem label="链接标识（Slug）" style="margin-bottom:0">
            <div class="w-full">
              <NInput v-model:value="form.slug" placeholder="留空则自动根据标题生成" />
              <p v-if="slugWarning" class="mt-1 text-xs text-red-400">{{ slugWarning }}</p>
              <p v-else class="mt-1 text-xs text-slate-400">可选，留空则自动生成；只允许小写字母、数字和连字符，不能为纯数字</p>
            </div>
          </NFormItem>
        </NCollapseItem>
      </NCollapse>
      <NSpace>
        <NButton :loading="loading" @click="save('draft')">保存草稿</NButton>
        <NButton type="primary" :loading="loading" @click="save('published')">发布</NButton>
      </NSpace>
    </NForm>
  </div>
</template>
