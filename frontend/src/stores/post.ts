import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { PostSummary } from '@/types/post'

export const usePostStore = defineStore('post', () => {
  const myPosts = ref<PostSummary[]>([])
  const draftCache = ref<Record<number, string>>({})

  const setMyPosts = (items: PostSummary[]) => {
    myPosts.value = items
  }

  const cacheDraft = (id: number, content: string) => {
    draftCache.value[id] = content
  }

  return { myPosts, draftCache, setMyPosts, cacheDraft }
})
