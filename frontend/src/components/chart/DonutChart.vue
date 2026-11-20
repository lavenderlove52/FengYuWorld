<script setup lang="ts">
import { computed, ref } from 'vue'

export interface DonutSlice {
  label: string
  value: number
  color: string
}

const props = withDefaults(
  defineProps<{
    slices: DonutSlice[]
    /** 环形内径占外径的比例 */
    innerRatio?: number
    size?: number
    /** 居中显示的主数值 */
    centerLabel?: string
    centerValue?: string
  }>(),
  { innerRatio: 0.62, size: 200, centerLabel: '', centerValue: '' },
)

const hovered = ref<number | null>(null)

const total = computed(() => props.slices.reduce((sum, s) => sum + Math.max(0, s.value), 0))

interface Arc {
  d: string
  slice: DonutSlice
  percent: number
  index: number
}

const arcs = computed<Arc[]>(() => {
  const sum = total.value
  if (sum <= 0) return []
  const cx = props.size / 2
  const cy = props.size / 2
  const rOuter = props.size / 2 - 1
  const rInner = rOuter * props.innerRatio
  const result: Arc[] = []
  let angle = -Math.PI / 2
  props.slices.forEach((slice, index) => {
    const value = Math.max(0, slice.value)
    if (value <= 0) return
    const sweep = (value / sum) * Math.PI * 2
    // 整圆用两段弧，否则 SVG 的 A 命令画不出 360°
    const end = angle + sweep
    const largeArc = sweep > Math.PI ? 1 : 0
    const d = [
      `M ${point(cx, cy, rOuter, angle)}`,
      `A ${rOuter} ${rOuter} 0 ${largeArc} 1 ${point(cx, cy, rOuter, end)}`,
      `L ${point(cx, cy, rInner, end)}`,
      `A ${rInner} ${rInner} 0 ${largeArc} 0 ${point(cx, cy, rInner, angle)}`,
      'Z',
    ].join(' ')
    result.push({ d, slice, percent: (value / sum) * 100, index })
    angle = end
  })
  return result
})

function point(cx: number, cy: number, r: number, angle: number): string {
  return `${(cx + r * Math.cos(angle)).toFixed(2)} ${(cy + r * Math.sin(angle)).toFixed(2)}`
}
</script>

<template>
  <div class="flex flex-wrap items-center gap-6">
    <svg :width="size" :height="size" :viewBox="`0 0 ${size} ${size}`" role="img" aria-label="占比分布">
      <circle :cx="size / 2" :cy="size / 2" :r="size / 2 - 1" fill="#f1f5f9" />
      <path
        v-for="arc in arcs"
        :key="arc.slice.label"
        :d="arc.d"
        :fill="arc.slice.color"
        :opacity="hovered === null || hovered === arc.index ? 1 : 0.45"
        class="transition-opacity duration-150 cursor-pointer"
        @mouseenter="hovered = arc.index"
        @mouseleave="hovered = null"
      >
        <title>{{ arc.slice.label }} {{ arc.percent.toFixed(2) }}%</title>
      </path>
      <template v-if="arcs.length === 0">
        <circle :cx="size / 2" :cy="size / 2" :r="(size / 2 - 1) * innerRatio" fill="#fff" />
      </template>
      <template v-else>
        <circle :cx="size / 2" :cy="size / 2" :r="(size / 2 - 1) * innerRatio" fill="#fff" />
      </template>
      <text
        v-if="centerValue"
        :x="size / 2"
        :y="size / 2 - (centerLabel ? 2 : -4)"
        text-anchor="middle"
        class="fill-slate-700"
        style="font-size: 15px; font-weight: 600"
      >{{ centerValue }}</text>
      <text
        v-if="centerLabel"
        :x="size / 2"
        :y="size / 2 + 16"
        text-anchor="middle"
        class="fill-slate-400"
        style="font-size: 11px"
      >{{ centerLabel }}</text>
    </svg>

    <ul class="flex-1 min-w-[160px] flex flex-col gap-1.5 m-0 p-0 list-none">
      <li
        v-for="(slice, index) in slices"
        :key="slice.label"
        class="flex items-center gap-2 text-sm"
        :class="hovered !== null && hovered !== index ? 'opacity-50' : ''"
        @mouseenter="hovered = index"
        @mouseleave="hovered = null"
      >
        <span class="w-2.5 h-2.5 rounded-sm shrink-0" :style="{ background: slice.color }" />
        <span class="truncate text-slate-600">{{ slice.label }}</span>
        <span class="ml-auto tabular-nums text-slate-500">
          {{ total > 0 ? ((Math.max(0, slice.value) / total) * 100).toFixed(1) : '0.0' }}%
        </span>
      </li>
      <li v-if="slices.length === 0" class="text-sm text-slate-400">暂无数据</li>
    </ul>
  </div>
</template>
