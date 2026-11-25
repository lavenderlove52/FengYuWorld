<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PublicPlanDetail from '@/components/gridplan/PublicPlanDetail.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { useUserTheme } from '@/composables/useUserTheme'

/**
 * 公开计划的直达页：/u/:username/plans/:id
 * 只读展示，可见性仍由 /users/:username/grid-plans/:id 判定；
 * 个人主页的「对访客展示投资计划」只控制页签，不影响这里的直达链接。
 */

const route = useRoute()
const router = useRouter()
const containerRef = ref<HTMLElement | null>(null)
const { theme, load: loadTheme } = useUserTheme(containerRef)

const username = computed(() => String(route.params.username ?? ''))
const planId = computed(() => Number(route.params.id))
const validPlanId = computed(() => Number.isInteger(planId.value) && planId.value > 0)

/** 与个人主页同一个宽度口径，两个公开入口看起来才是同一个页面。 */
const contentStyle = computed(() => ({
  maxWidth: `${Math.min(1800, Math.max(600, theme.value.hero_config.page_width_px))}px`,
}))

function backToUserHome(): void {
  router.push({ name: 'user-home', params: { username: username.value } })
}

function loadThemeForUser(): void {
  if (username.value) void loadTheme(username.value)
}

onMounted(loadThemeForUser)
watch(username, loadThemeForUser)
</script>

<template>
  <div
    ref="containerRef"
    :data-theme-user="username"
    data-theme-page="user-plan-detail"
    class="relative isolate min-h-screen"
    style="color: var(--theme-text)"
  >
    <div class="mx-auto w-[calc(100%-2rem)] sm:w-[calc(100%-3rem)] pt-8 pb-12 sm:pt-10" :style="contentStyle">
      <EmptyState v-if="!validPlanId" message="计划地址不正确" />
      <PublicPlanDetail
        v-else
        :username="username"
        :plan-id="planId"
        back-label="返回个人主页"
        @back="backToUserHome"
      />
    </div>
  </div>
</template>
