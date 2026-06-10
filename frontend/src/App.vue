<script setup lang="ts">
import { RouterView } from 'vue-router'
import { NConfigProvider, NMessageProvider, NDialogProvider, zhCN, dateZhCN } from 'naive-ui'
import { onBeforeUnmount, onMounted } from 'vue'
import { FORCE_LOGOUT_EVENT } from '@/api/http'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const handleForceLogout = () => {
  auth.user = null
  auth.tokenExists = false
}

onMounted(() => {
  window.addEventListener(FORCE_LOGOUT_EVENT, handleForceLogout)
})

onBeforeUnmount(() => {
  window.removeEventListener(FORCE_LOGOUT_EVENT, handleForceLogout)
})
</script>

<template>
  <NConfigProvider :locale="zhCN" :date-locale="dateZhCN">
    <NMessageProvider>
      <NDialogProvider>
        <RouterView />
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>
