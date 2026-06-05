import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import * as settingsApi from '@/api/settings'

export const useSettingsStore = defineStore('settings', () => {
  const friendChatEnabled = ref(true)
  const loaded = ref(false)
  const loading = ref(false)

  const canChat = computed(() => friendChatEnabled.value)

  async function loadPublicSettings() {
    if (loading.value) return
    loading.value = true
    try {
      const res = await settingsApi.getPublicSettings()
      if (res.code === 0 && res.data) {
        friendChatEnabled.value = res.data.friend_chat_enabled
        loaded.value = true
      }
    } finally {
      loading.value = false
    }
  }

  async function setFriendChatEnabled(enabled: boolean) {
    const res = await settingsApi.updateFriendChatSetting(enabled)
    if (res.code === 0 && res.data) {
      friendChatEnabled.value = res.data.friend_chat_enabled
      loaded.value = true
    }
    return res
  }

  return {
    friendChatEnabled,
    loaded,
    loading,
    canChat,
    loadPublicSettings,
    setFriendChatEnabled,
  }
})
