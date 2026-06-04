<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { activePageTheme } from '@/stores/pageTheme'

const route = useRoute()
const usePageTheme = computed(() =>
  ['user-home', 'post-detail'].includes(String(route.name)) && activePageTheme.enabled,
)
const footerStyle = computed(() => usePageTheme.value
  ? {
      backgroundColor: activePageTheme.bg,
      borderColor: activePageTheme.border,
      color: activePageTheme.text,
    }
  : {
      backgroundColor: '#ffffff',
      borderColor: '#e2e8f0',
      color: '#64748b',
    },
)
</script>

<template>
  <footer class="border-t mt-auto transition-colors duration-300" :style="footerStyle">
    <div class="max-w-5xl mx-auto px-4 py-6 text-center text-sm opacity-70">
      欢迎来到风雨世界
    </div>
  </footer>
</template>
