<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppHeader from '@/components/common/AppHeader.vue'
import AppFooter from '@/components/common/AppFooter.vue'
import { useInactivityLogout } from '@/composables/useInactivityLogout'

useInactivityLogout()
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <AppHeader />
    <main class="flex-1 max-w-7xl w-full mx-auto px-8 py-8">
      <RouterView />
    </main>
    <AppFooter />
  </div>
</template>
