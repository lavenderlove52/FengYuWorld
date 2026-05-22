<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NInput, NButton, NSpace, useMessage } from 'naive-ui'
import * as friendsApi from '@/api/friends'
import { useFriendStore } from '@/stores/friend'
import FriendRequestList from '@/components/friend/FriendRequestList.vue'
import FriendList from '@/components/friend/FriendList.vue'

const message = useMessage()
const friendStore = useFriendStore()
const username = ref('')

const load = async () => {
  await Promise.all([friendStore.loadFriends(), friendStore.loadReceived()])
}

onMounted(load)

const send = async () => {
  if (!username.value.trim()) return
  const res = await friendsApi.sendFriendRequest(username.value.trim())
  if (res.code === 0) {
    message.success('申请已发送')
    username.value = ''
    load()
  } else {
    message.error(res.message)
  }
}

const accept = async (id: number) => {
  await friendsApi.acceptRequest(id)
  message.success('已接受')
  load()
}

const reject = async (id: number) => {
  await friendsApi.rejectRequest(id)
  message.info('已拒绝')
  load()
}

const remove = async (id: number) => {
  await friendsApi.deleteFriend(id)
  message.success('已删除好友')
  load()
}
</script>

<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-2xl font-bold mb-4">好友管理</h1>
      <NSpace>
        <NInput v-model:value="username" placeholder="输入用户名添加好友" style="width: 240px" />
        <NButton type="primary" @click="send">发送申请</NButton>
      </NSpace>
    </div>
    <section>
      <h2 class="font-semibold mb-3">收到的申请</h2>
      <FriendRequestList :items="friendStore.received" @accept="accept" @reject="reject" />
    </section>
    <section>
      <h2 class="font-semibold mb-3">好友列表</h2>
      <FriendList :items="friendStore.friends" @remove="remove" />
    </section>
  </div>
</template>
