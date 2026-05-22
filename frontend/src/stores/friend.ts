import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as friendsApi from '@/api/friends'
import type { FriendListItem, FriendshipItem } from '@/types/friend'

export const useFriendStore = defineStore('friend', () => {
  const friends = ref<FriendListItem[]>([])
  const received = ref<FriendshipItem[]>([])

  const loadFriends = async () => {
    const res = await friendsApi.listFriends()
    if (res.code === 0 && res.data) {
      friends.value = res.data
    }
  }

  const loadReceived = async () => {
    const res = await friendsApi.listReceivedRequests()
    if (res.code === 0 && res.data) {
      received.value = res.data
    }
  }

  return { friends, received, loadFriends, loadReceived }
})
