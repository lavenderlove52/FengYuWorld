<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

export interface TradePoint {
  /** 成交时间，yyyy-MM-dd HH:mm:ss */
  date: string
  point: number
  side: 'buy' | 'sell'
}

const props = withDefaults(defineProps<{ points: TradePoint[]; height?: number }>(), { height: 140 })

const COLOR_BUY = '#10b981'
const COLOR_SELL = '#f59e0b'
const COLOR_LINE = '#cbd5e1'
const PAD = { top: 14, right: 14, bottom: 20, left: 54 }

const width = ref(0)
const container = ref<HTMLDivElement | null>(null)
// ResizeObserver 首次回调前按 640 兜底，避免 svg 宽度为 0
const svgWidth = computed(() => (width.value > 0 ? width.value : 640))
const chartW = computed(() => Math.max(160, svgWidth.value - PAD.left - PAD.right))
const chartH = computed(() => Math.max(60, props.height - PAD.top - PAD.bottom))

const points = computed(() =>
  [...props.points].sort((a, b) => (a.date < b.date ? -1 : a.date > b.date ? 1 : 0)),
)

const range = computed(() => {
  const values = points.value.map((p) => p.point)
  if (!values.length) return { min: 0, max: 1, ticks: [0, 1] }
  let min = Math.min(...values)
  let max = Math.max(...values)
  const pad = (max - min) * 0.15 || Math.abs(max) * 0.02 || 1
  min -= pad
  max += pad
  const ticks = [min, (min + max) / 2, max]
  return { min, max, ticks }
})

function x(index: number): number {
  const n = points.value.length
  if (n <= 1) return PAD.left + chartW.value / 2
  return PAD.left + (chartW.value * index) / (n - 1)
}

function y(value: number): number {
  const { min, max } = range.value
  const ratio = max === min ? 0.5 : (value - min) / (max - min)
  return PAD.top + chartH.value * (1 - ratio)
}

const path = computed(() => {
  const pts = points.value
  if (pts.length < 2) return ''
  return pts.map((p, i) => `${i === 0 ? 'M' : 'L'} ${x(i).toFixed(1)} ${y(p.point).toFixed(1)}`).join(' ')
})

function fmt(v: number): string {
  return v.toLocaleString('zh-CN', { maximumFractionDigits: 2 })
}

function labelDate(date: string): string {
  return date.slice(0, 10)
}

function onResize(): void {
  if (container.value) width.value = container.value.clientWidth
}

let observer: ResizeObserver | null = null

onMounted(() => {
  if (!container.value) return
  onResize()
  observer = new ResizeObserver(onResize)
  observer.observe(container.value)
})

onBeforeUnmount(() => {
  observer?.disconnect()
  observer = null
})
</script>

<template>
  <div ref="container" class="w-full min-w-0 overflow-hidden">
    <div
      v-if="!points.length"
      class="flex items-center justify-center rounded-lg border border-dashed border-slate-200 text-xs text-slate-400"
      :style="{ height: `${height}px` }"
    >
      还没有带点位的买卖记录，记录买入/卖出时填上点位就会画在这里
    </div>
    <svg
      v-else
      :width="svgWidth"
      :height="height"
      :viewBox="`0 0 ${svgWidth} ${height}`"
      class="w-full select-none"
    >
      <!-- Y 轴刻度 -->
      <g>
        <line
          v-for="(tick, i) in range.ticks"
          :key="`grid-${i}`"
          :x1="PAD.left"
          :x2="PAD.left + chartW"
          :y1="y(tick)"
          :y2="y(tick)"
          stroke="#eef2f7"
          stroke-width="1"
        />
        <text
          v-for="(tick, i) in range.ticks"
          :key="`tick-${i}`"
          :x="PAD.left - 6"
          :y="y(tick) + 4"
          text-anchor="end"
          class="fill-slate-400"
          style="font-size: 10px"
        >{{ fmt(tick) }}</text>
      </g>

      <!-- 点位轨迹 -->
      <path v-if="path" :d="path" fill="none" :stroke="COLOR_LINE" stroke-width="1.4" />

      <!-- 买卖点 -->
      <g v-for="(p, i) in points" :key="`${p.date}-${i}`">
        <circle :cx="x(i)" :cy="y(p.point)" :r="p.side === 'buy' ? 4 : 4.5"
          :fill="p.side === 'buy' ? COLOR_BUY : COLOR_SELL" opacity="0.9">
          <title>{{ labelDate(p.date) }} {{ p.side === 'buy' ? '买入' : '卖出' }} {{ fmt(p.point) }}</title>
        </circle>
      </g>

      <!-- X 轴：首末两条 -->
      <text :x="x(0)" :y="height - 6" text-anchor="start" class="fill-slate-400" style="font-size: 10px">
        {{ labelDate(points[0]!.date) }}
      </text>
      <text
        v-if="points.length > 1"
        :x="x(points.length - 1)"
        :y="height - 6"
        text-anchor="end"
        class="fill-slate-400"
        style="font-size: 10px"
      >{{ labelDate(points[points.length - 1]!.date) }}</text>
    </svg>

    <div v-if="points.length" class="flex items-center gap-4 text-xs text-slate-500 px-1 mt-1">
      <span class="flex items-center gap-1.5">
        <i class="inline-block w-2.5 h-2.5 rounded-full" :style="{ background: COLOR_BUY }" />买入
      </span>
      <span class="flex items-center gap-1.5">
        <i class="inline-block w-2.5 h-2.5 rounded-full" :style="{ background: COLOR_SELL }" />卖出
      </span>
    </div>
  </div>
</template>
