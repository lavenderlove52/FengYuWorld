<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { NSelect } from 'naive-ui'
import { useFriendStore } from '@/stores/friend'

const model = defineModel<number[]>({ default: () => [] })
const friendStore = useFriendStore()

onMounted(() => {
  friendStore.loadFriends()
})

const options = computed(() =>
  friendStore.friends.map((f) => ({ label: f.nickname, value: f.id })),
)
</script>

<template>
  <NSelect
    v-model:value="model"
    multiple
    filterable
    placeholder="选择可见好友"
    :options="options"
  />
</template>
