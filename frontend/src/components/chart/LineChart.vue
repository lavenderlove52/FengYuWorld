<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

export interface LinePoint {
  /** yyyy-MM-dd */
  date: string
  invested: number
  marketValue: number
  pnl: number
}

const props = withDefaults(
  defineProps<{
    points: LinePoint[]
    height?: number
    /** 是否画盈亏零轴 */
    showPnl?: boolean
  }>(),
  { height: 220, showPnl: true },
)

const COLOR_INVESTED = '#94a3b8'
const COLOR_MARKET = '#4f46e5'
const COLOR_PNL = '#f97316'

const PAD = { top: 16, right: 16, bottom: 26, left: 62 }

const width = ref(720)
const container = ref<HTMLDivElement | null>(null)
const hovered = ref<number | null>(null)
const chartW = computed(() => Math.max(240, width.value - PAD.left - PAD.right))
const chartH = computed(() => Math.max(80, props.height - PAD.top - PAD.bottom))

const sorted = computed(() =>
  [...props.points].sort((a, b) => (a.date < b.date ? -1 : a.date > b.date ? 1 : 0)),
)

const hasPnl = computed(() => props.showPnl && sorted.value.some((p) => p.pnl !== 0))

const range = computed(() => {
  const values: number[] = []
  for (const p of sorted.value) {
    values.push(p.invested, p.marketValue)
    if (hasPnl.value) values.push(p.pnl)
  }
  if (values.length === 0) return { min: 0, max: 1, ticks: [0, 1] }
  if (values.length === 1) {
    const v = values[0]!
    const pad = Math.abs(v) * 0.1 || 1
    return niceRange(v - pad, v + pad)
  }
  let min = Math.min(...values)
  let max = Math.max(...values)
  if (min === max) {
    min -= 1
    max += 1
  }
  return niceRange(min, max)
})

function niceRange(min: number, max: number): { min: number; max: number; ticks: number[] } {
  const span = max - min
  const rough = span / 4
  const mag = Math.pow(10, Math.floor(Math.log10(rough || 1)))
  const step = [1, 2, 2.5, 5, 10].map((m) => m * mag).find((s) => s >= rough) ?? mag * 10
  const lo = Math.floor(min / step) * step
  const hi = Math.ceil(max / step) * step
  const ticks: number[] = []
  for (let v = lo; v <= hi + step / 2; v += step) ticks.push(Math.round(v * 1e6) / 1e6)
  return { min: lo, max: hi, ticks }
}

function x(index: number): number {
  const n = sorted.value.length
  if (n <= 1) return PAD.left + chartW.value / 2
  return PAD.left + (chartW.value * index) / (n - 1)
}

function y(value: number): number {
  const { min, max } = range.value
  const ratio = max === min ? 0.5 : (value - min) / (max - min)
  return PAD.top + chartH.value * (1 - ratio)
}

function linePath(key: 'invested' | 'marketValue' | 'pnl'): string {
  const pts = sorted.value
  if (!pts.length) return ''
  return pts.map((p, i) => `${i === 0 ? 'M' : 'L'} ${x(i).toFixed(2)} ${y(p[key]).toFixed(2)}`).join(' ')
}

const zeroY = computed(() => (range.value.min < 0 && range.value.max > 0 ? y(0) : null))

const xLabels = computed(() => {
  const pts = sorted.value
  if (!pts.length) return []
  if (pts.length === 1) return [{ x: x(0), text: pts[0]!.date }]
  const count = Math.min(6, pts.length)
  const step = (pts.length - 1) / (count - 1)
  const out: Array<{ x: number; text: string }> = []
  for (let i = 0; i < count; i += 1) {
    const index = Math.round(i * step)
    out.push({ x: x(index), text: pts[index]!.date.slice(5) })
  }
  return out
})

function onMove(event: MouseEvent): void {
  const el = event.currentTarget as SVGSVGElement
  const rect = el.getBoundingClientRect()
  const scale = width.value / rect.width
  const px = (event.clientX - rect.left) * scale
  const n = sorted.value.length
  if (n === 0) return
  if (n === 1) {
    hovered.value = 0
    return
  }
  const ratio = (px - PAD.left) / chartW.value
  const index = Math.round(ratio * (n - 1))
  hovered.value = Math.max(0, Math.min(n - 1, index))
}

function fmtTick(v: number): string {
  const abs = Math.abs(v)
  if (abs >= 10000) return `${(v / 10000).toFixed(abs >= 100000 ? 0 : 1)}万`
  return v.toLocaleString('zh-CN', { maximumFractionDigits: 0 })
}

const hoverPoint = computed(() => {
  if (hovered.value === null) return null
  const p = sorted.value[hovered.value]
  if (!p) return null
  return { p, cx: x(hovered.value), cy: y(p.marketValue) }
})

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
  <div ref="container" class="w-full">
    <svg
      :width="width"
      :height="height"
      :viewBox="`0 0 ${width} ${height}`"
      class="w-full select-none"
      @mousemove="onMove"
      @mouseleave="hovered = null"
    >
      <!-- Y 轴刻度 -->
      <g>
        <line
          v-for="tick in range.ticks"
          :key="`grid-${tick}`"
          :x1="PAD.left"
          :x2="PAD.left + chartW"
          :y1="y(tick)"
          :y2="y(tick)"
          stroke="#eef2f7"
          stroke-width="1"
        />
        <text
          v-for="tick in range.ticks"
          :key="`tick-${tick}`"
          :x="PAD.left - 8"
          :y="y(tick) + 4"
          text-anchor="end"
          class="fill-slate-400"
          style="font-size: 11px"
        >{{ fmtTick(tick) }}</text>
      </g>

      <!-- 零轴 -->
      <line
        v-if="zeroY !== null"
        :x1="PAD.left"
        :x2="PAD.left + chartW"
        :y1="zeroY"
        :y2="zeroY"
        stroke="#cbd5e1"
        stroke-dasharray="4 3"
      />

      <!-- 折线 -->
      <path :d="linePath('invested')" fill="none" :stroke="COLOR_INVESTED" stroke-width="1.6" />
      <path :d="linePath('marketValue')" fill="none" :stroke="COLOR_MARKET" stroke-width="2" />
      <path v-if="hasPnl" :d="linePath('pnl')" fill="none" :stroke="COLOR_PNL" stroke-width="1.6" stroke-dasharray="5 3" />

      <!-- 单点也要能看到 -->
      <template v-if="sorted.length === 1">
        <circle :cx="x(0)" :cy="y(sorted[0]!.invested)" r="3.5" :fill="COLOR_INVESTED" />
        <circle :cx="x(0)" :cy="y(sorted[0]!.marketValue)" r="4" :fill="COLOR_MARKET" />
        <circle v-if="hasPnl" :cx="x(0)" :cy="y(sorted[0]!.pnl)" r="3.5" :fill="COLOR_PNL" />
      </template>

      <!-- X 轴 -->
      <g>
        <text
          v-for="label in xLabels"
          :key="`x-${label.text}-${label.x}`"
          :x="label.x"
          :y="height - 8"
          text-anchor="middle"
          class="fill-slate-400"
          style="font-size: 11px"
        >{{ label.text }}</text>
      </g>

      <!-- 十字线与提示 -->
      <template v-if="hoverPoint">
        <line
          :x1="hoverPoint.cx"
          :x2="hoverPoint.cx"
          :y1="PAD.top"
          :y2="PAD.top + chartH"
          stroke="#94a3b8"
          stroke-dasharray="3 3"
        />
        <circle :cx="hoverPoint.cx" :cy="y(hoverPoint.p.marketValue)" r="4" :fill="COLOR_MARKET" />
        <circle :cx="hoverPoint.cx" :cy="y(hoverPoint.p.invested)" r="3" :fill="COLOR_INVESTED" />
      </template>

      <text v-if="sorted.length === 0" :x="width / 2" :y="height / 2" text-anchor="middle"
        class="fill-slate-400" style="font-size: 13px">还没有净值记录</text>
    </svg>

    <div class="flex flex-wrap items-center gap-4 text-xs text-slate-500 px-2">
      <span class="flex items-center gap-1.5">
        <i class="inline-block w-3 h-[2px]" :style="{ background: COLOR_INVESTED }" />投入
      </span>
      <span class="flex items-center gap-1.5">
        <i class="inline-block w-3 h-[2px]" :style="{ background: COLOR_MARKET }" />市值
      </span>
      <span v-if="hasPnl" class="flex items-center gap-1.5">
        <i class="inline-block w-3 h-[2px]" :style="{ background: COLOR_PNL, opacity: 0.9 }" />盈亏
      </span>
      <span v-if="hoverPoint" class="ml-auto tabular-nums text-slate-700">
        {{ hoverPoint.p.date }} · 投入 {{ fmtTick(hoverPoint.p.invested) }} ·
        市值 {{ fmtTick(hoverPoint.p.marketValue) }} ·
        盈亏 {{ hoverPoint.p.pnl >= 0 ? '+' : '' }}{{ fmtTick(hoverPoint.p.pnl) }}
      </span>
    </div>
  </div>
</template>
