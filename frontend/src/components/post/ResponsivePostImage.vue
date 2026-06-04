<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import { postMobileImageURL } from '@/utils/responsiveImage'

defineOptions({ inheritAttrs: false })

const props = defineProps<{
  src: string
  alt?: string
}>()

const attrs = useAttrs()
const mobileSrc = computed(() => postMobileImageURL(props.src))
</script>

<template>
  <picture>
    <source v-if="mobileSrc" media="(max-width: 767px)" :srcset="mobileSrc" />
    <img
      :src="src"
      :alt="alt || ''"
      loading="lazy"
      decoding="async"
      v-bind="attrs"
    />
  </picture>
</template>

<style scoped>
picture {
  display: contents;
}
</style>
