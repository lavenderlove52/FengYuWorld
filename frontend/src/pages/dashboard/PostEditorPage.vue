<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, NSpace, useMessage } from 'naive-ui'
import { createPost, getPost, updatePost } from '@/api/posts'
import { listCategories } from '@/api/categories'
import type { Category } from '@/types/category'
import type { CreatePostPayload, PostVisibility } from '@/types/post'
import MarkdownEditor from '@/components/post/MarkdownEditor.vue'
import VisibilitySelect from '@/components/post/VisibilitySelect.vue'
import FriendMultiSelect from '@/components/post/FriendMultiSelect.vue'
import ImageUploader from '@/components/upload/ImageUploader.vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const isEdit = computed(() => !!route.params.id)
const postId = computed(() => Number(route.params.id))

const form = ref<CreatePostPayload>({
  title: '',
  slug: '',
  cover_image_url: '',
  content_md: '# 标题\n\n开始写作...',
  tag_names: [],
  status: 'draft',
  visibility: 'private',
  visible_user_ids: [],
  allow_comment: true,
})

const tagInput = ref('')
const categories = ref<Category[]>([])
const loading = ref(false)

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
        content_md: p.content_md || '',
        category_id: p.category_id,
        tag_names: p.tags || [],
        status: p.status as any,
        visibility: p.visibility as PostVisibility,
        visible_user_ids: p.visible_user_ids || [],
        allow_comment: p.allow_comment,
      }
      tagInput.value = (p.tags || []).join(', ')
    }
  }
})

const buildPayload = (status: 'draft' | 'published'): CreatePostPayload => ({
  ...form.value,
  status,
  tag_names: tagInput.value
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean),
})

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
      <NFormItem label="Slug（可选）">
        <NInput v-model:value="form.slug" placeholder="url-slug" />
      </NFormItem>
      <NFormItem label="封面">
        <ImageUploader v-model="form.cover_image_url" />
      </NFormItem>
      <NFormItem label="分类">
        <select v-model="form.category_id" class="border rounded px-3 py-2 w-full">
          <option :value="undefined">无分类</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
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
      <NFormItem label="正文">
        <MarkdownEditor v-model="form.content_md" />
      </NFormItem>
      <NSpace>
        <NButton :loading="loading" @click="save('draft')">保存草稿</NButton>
        <NButton type="primary" :loading="loading" @click="save('published')">发布</NButton>
      </NSpace>
    </NForm>
  </div>
</template>
