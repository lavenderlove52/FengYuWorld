<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NButton,
  NCollapse,
  NCollapseItem,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSelect,
  NSpace,
  NSwitch,
  useMessage,
} from 'naive-ui'
import { createPost, getPost, updatePost } from '@/api/posts'
import { createPostTreeFolder, listPostTree, updatePostTreeNode } from '@/api/postTree'
import type { CreatePostPayload, PostVisibility } from '@/types/post'
import type { PostTreeNode } from '@/types/postTree'
import MarkdownEditor from '@/components/post/MarkdownEditor.vue'
import VisibilitySelect from '@/components/post/VisibilitySelect.vue'
import FriendMultiSelect from '@/components/post/FriendMultiSelect.vue'
import ImageUploader from '@/components/upload/ImageUploader.vue'
import { validatePostSlug } from '@/utils/postSlug'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const ROOT_PARENT_ID = 0

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
const loading = ref(false)
const tree = ref<PostTreeNode[]>([])

const directoryModalVisible = ref(false)
const newDirectoryName = ref('')
const newDirectoryParentId = ref<number | null>(null)
const renameDirectoryId = ref<number | null>(null)
const renameDirectoryName = ref('')
const savingDirectory = ref(false)

const slugWarning = computed(() => validatePostSlug(form.value.slug || ''))

const flattenNodes = (nodes: PostTreeNode[]): PostTreeNode[] =>
  nodes.flatMap((node) => [node, ...flattenNodes(node.children || [])])

const allNodes = computed(() => flattenNodes(tree.value))
const folderNodes = computed(() => allNodes.value.filter((node) => node.node_type === 'folder'))
const leafFolderNodes = computed(() =>
  folderNodes.value.filter((folder) => !(folder.children || []).some((child) => child.node_type === 'folder')),
)

const folderPath = (folder: PostTreeNode) => {
  const names = [folder.name]
  let parentId = folder.parent_id ?? null
  while (parentId) {
    const parent = folderNodes.value.find((item) => item.id === parentId)
    if (!parent) break
    names.unshift(parent.name)
    parentId = parent.parent_id ?? null
  }
  return names.join(' / ')
}

const directoryOptions = computed<any[]>(() =>
  leafFolderNodes.value.map((folder) => ({
    label: folderPath(folder),
    value: folder.id,
  })),
)

const parentDirectoryOptions = computed<any[]>(() => [
  { label: '根目录', value: ROOT_PARENT_ID },
  ...folderNodes.value.map((folder) => ({
    label: folderPath(folder),
    value: folder.id,
    disabled: (folder.children || []).some((child) => child.node_type === 'post'),
  })),
])

const renameDirectoryOptions = computed<any[]>(() =>
  folderNodes.value.map((folder) => ({
    label: folderPath(folder),
    value: folder.id,
  })),
)

const loadTree = async () => {
  const res = await listPostTree()
  if (res.code === 0 && res.data) {
    tree.value = res.data
  }
}

onMounted(async () => {
  await loadTree()

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
        tree_parent_id: p.tree_parent_id,
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

const openDirectoryModal = () => {
  newDirectoryName.value = ''
  newDirectoryParentId.value = ROOT_PARENT_ID
  renameDirectoryId.value = form.value.tree_parent_id || null
  const current = folderNodes.value.find((folder) => folder.id === renameDirectoryId.value)
  renameDirectoryName.value = current?.name || ''
  directoryModalVisible.value = true
}

const normalizeParentId = (value: number | null) => value === ROOT_PARENT_ID ? null : value

const submitNewDirectory = async () => {
  const name = newDirectoryName.value.trim()
  if (!name) {
    message.warning('目录名称不能为空')
    return
  }
  savingDirectory.value = true
  try {
    const res = await createPostTreeFolder({
      name,
      parent_id: normalizeParentId(newDirectoryParentId.value),
      sort_order: 10,
    })
    if (res.code === 0 && res.data) {
      message.success('目录已创建')
      newDirectoryName.value = ''
      await loadTree()
      form.value.tree_parent_id = res.data.id
    } else {
      message.error(res.message || '创建失败')
    }
  } finally {
    savingDirectory.value = false
  }
}

const onRenameDirectoryChange = (id: number | null) => {
  const node = folderNodes.value.find((folder) => folder.id === id)
  renameDirectoryName.value = node?.name || ''
}

const submitRenameDirectory = async () => {
  if (!renameDirectoryId.value) {
    message.warning('请选择目录')
    return
  }
  const node = folderNodes.value.find((folder) => folder.id === renameDirectoryId.value)
  const name = renameDirectoryName.value.trim()
  if (!node || !name) {
    message.warning('目录名称不能为空')
    return
  }
  savingDirectory.value = true
  try {
    const res = await updatePostTreeNode(node.id, {
      name,
      parent_id: node.parent_id ?? null,
      sort_order: node.sort_order,
    })
    if (res.code === 0) {
      message.success('目录已重命名')
      await loadTree()
    } else {
      message.error(res.message || '重命名失败')
    }
  } finally {
    savingDirectory.value = false
  }
}

const buildPayload = (status: 'draft' | 'published'): CreatePostPayload => {
  const payload: CreatePostPayload = {
    ...form.value,
    status,
    tag_names: tagInput.value
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean),
  }
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
  if (!form.value.tree_parent_id) {
    message.warning('请选择文章所属的最底层目录')
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
      <NFormItem label="所属目录">
        <div class="w-full flex items-center gap-2">
          <NSelect
            v-model:value="form.tree_parent_id"
            :options="directoryOptions"
            placeholder="请选择最底层目录"
            filterable
            class="flex-1"
          />
          <NButton @click="openDirectoryModal">编辑目录</NButton>
        </div>
        <p class="mt-1 text-xs text-slate-400">文章只能归属到没有子目录的最底层目录。</p>
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

    <NModal v-model:show="directoryModalVisible" preset="card" class="directory-modal" title="编辑文章目录">
      <div class="grid gap-5">
        <section class="grid gap-2">
          <h2 class="text-base font-semibold text-slate-800">新增目录</h2>
          <NInput v-model:value="newDirectoryName" placeholder="目录名称" />
          <NSelect v-model:value="newDirectoryParentId" :options="parentDirectoryOptions" placeholder="选择上级目录" />
          <NButton type="primary" :loading="savingDirectory" @click="submitNewDirectory">新增目录</NButton>
        </section>
        <section class="grid gap-2">
          <h2 class="text-base font-semibold text-slate-800">修改目录名称</h2>
          <NSelect
            v-model:value="renameDirectoryId"
            :options="renameDirectoryOptions"
            filterable
            placeholder="选择目录"
            @update:value="onRenameDirectoryChange"
          />
          <NInput v-model:value="renameDirectoryName" placeholder="新的目录名称" />
          <NButton :loading="savingDirectory" @click="submitRenameDirectory">保存名称</NButton>
        </section>
      </div>
      <template #footer>
        <div class="flex justify-end">
          <NButton @click="directoryModalVisible = false">关闭</NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.directory-modal {
  width: min(560px, calc(100vw - 32px));
}
</style>
