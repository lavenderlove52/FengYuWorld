<script setup lang="ts">
import { computed, h, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import type { VNodeChild } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import {
  NButton, NCard, NDataTable, NDatePicker, NFormItem, NInput, NInputNumber, NResult, NSelect, NSpin, NSwitch, NTag, NTooltip,
  useDialog, useMessage,
} from 'naive-ui'
import type { DataTableColumns, SelectGroupOption, SelectOption } from 'naive-ui'
import {
  addAsset, addGridLayer, applyOptimizedGrid, applyTemplate, clearExecutedLayers, discardChanges, loadPlan,
  planMetrics, planState, PlanConflictError, recordExecution, recordSnapshot, removeAsset, removeGridLayer,
  resetPlanState, restoreLayer, savePlan, setLayerAmount, setReferenceHighToCurrent, undoLastExecution,
  updateAggregation, updateAssetEnabled,
  updateAssetLeveraged, updateAssetNumber, updateAssetText, updateBudget, updateCurrency, updateFrequency,
  updateGateMode, updateGateThreshold, updateLayer, updateNextDate, updatePlanName, updateTargetBudget,
  updateUnallocated, updateVisibility, updateVisibleUsers, updateWeightMethod, voidLayer,
} from '@/core/gridplan/planState'
import { formatPlanMoney, formatSignedMoney } from '@/core/gridplan/planFormat'
import { formatNumber, formatPercent, formatPlainPercent } from '@/core/gridplan/format'
import type { PlanAssetMetrics, PlanAssetStatus, PlanGridRow } from '@/core/gridplan/planMetrics'
import { FREQUENCY_LABEL, todayISO } from '@/core/gridplan/format'
import { listGridSnapshots } from '@/api/gridPlans'
import DonutChart, { type DonutSlice } from '@/components/chart/DonutChart.vue'
import LineChart, { type LinePoint } from '@/components/chart/LineChart.vue'
import FriendMultiSelect from '@/components/post/FriendMultiSelect.vue'
import { CURRENCIES, findTemplate, gridForIndex, indicesFor, PLAN_TEMPLATES } from '@/core/gridplan/templates'
import {
  MIN_LAYERS, optimizeGrid, suggestLayerCount, VOLATILITY_LABEL, volatilityOf, type GridOptimizeResult,
} from '@/core/gridplan/gridOptimizer'
import { ensurePresets, presetsFor, refreshPresets, removePreset } from '@/core/gridplan/presetStore'
import type { Frequency, GateMode, GridAssetDTO, GridLayerDTO, GridSnapshot, PlanVisibility, WeightMethod } from '@/types/gridPlan'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const planId = computed(() => Number(route.params.id))

const statusMeta: Record<PlanAssetStatus, { label: string; type: 'success' | 'warning' | 'error' | 'info' | 'default' }> = {
  invalid: { label: '参数异常', type: 'error' },
  'gate-blocked': { label: '闸门限制', type: 'warning' },
  'budget-blocked': { label: '预算受限', type: 'warning' },
  'grid-triggered': { label: '网格触发', type: 'success' },
  'regular-due': { label: '定投到期', type: 'info' },
  waiting: { label: '等待', type: 'default' },
}

const plan = computed(() => planState.plan)
const metrics = computed(() => planMetrics.value)

const VISIBILITY_OPTIONS = [
  { value: 'private', label: '仅自己' },
  { value: 'login', label: '登录用户可见' },
  { value: 'friends', label: '全部好友' },
  { value: 'specified', label: '指定好友' },
  { value: 'public', label: '公开（含未登录访客）' },
]

/** 指定可见好友的多选值，直接双向绑到 planState，切档位或保存都以它为准。 */
const visibleUserIds = computed<number[]>({
  get: () => planState.plan?.visible_user_ids ?? [],
  set: (ids) => updateVisibleUsers(ids),
})

// ---------- 快照曲线 ----------

const snapshots = ref<GridSnapshot[]>([])
const snapshotLoading = ref(false)

// days=0 表示全部历史
const SNAPSHOT_RANGES = [
  { days: 30, label: '1个月' },
  { days: 60, label: '2个月' },
  { days: 90, label: '3个月' },
  { days: 180, label: '6个月' },
  { days: 365, label: '1年' },
  { days: 730, label: '2年' },
  { days: 1095, label: '3年' },
  { days: 1825, label: '5年' },
  { days: 0, label: '全部' },
] as const
const rangeDays = ref<number>(0)
const rangeLabel = computed(
  () => SNAPSHOT_RANGES.find((item) => item.days === rangeDays.value)?.label ?? '全部',
)

async function loadSnapshots(): Promise<void> {
  if (!planId.value) return
  snapshotLoading.value = true
  try {
    const res = await listGridSnapshots(planId.value, rangeDays.value)
    if (res.code === 0) snapshots.value = res.data ?? []
  } catch {
    // 曲线属于辅助信息，失败不打断主流程
  } finally {
    snapshotLoading.value = false
  }
}

watch(rangeDays, () => {
  void loadSnapshots()
})

const chartPoints = computed<LinePoint[]>(() =>
  snapshots.value.map((s) => ({
    date: s.snapshot_date,
    invested: s.invested,
    marketValue: s.market_value,
    pnl: s.pnl,
  })),
)

const lastSnapshotLabel = computed(() => {
  const last = snapshots.value[snapshots.value.length - 1]
  if (!last) return '还没有记录过净值'
  const days = Math.floor((Date.now() - new Date(`${last.snapshot_date}T00:00:00`).getTime()) / 86_400_000)
  if (days <= 0) return '最近记录：今天'
  return `最近记录：${days} 天前（${last.snapshot_date}）`
})

const snapshotStale = computed(() => {
  const last = snapshots.value[snapshots.value.length - 1]
  if (!last) return true
  const days = Math.floor((Date.now() - new Date(`${last.snapshot_date}T00:00:00`).getTime()) / 86_400_000)
  return days >= 7
})

// 记录净值走后端汇总已保存的数据，本地改动没保存就记会把旧值写进曲线
const recordBlocked = computed(() => planState.dirty || planState.saving)

const DONUT_COLORS = ['#4f46e5', '#0ea5e9', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#14b8a6', '#f97316']

const donutSlices = computed<DonutSlice[]>(() => {
  const rows = metrics.value?.assets ?? []
  const slices: DonutSlice[] = []
  let colorIndex = 0
  let filled = 0
  let missing = 0
  let unfilledInvested = 0
  for (const row of rows) {
    if (!row.asset.enabled) continue
    if (!row.hasMarketValue) {
      missing += 1
      unfilledInvested += Number(row.asset.invested) || 0
      continue
    }
    const value = Number(row.asset.market_value) || 0
    filled += value
    slices.push({
      label: row.asset.name || '未命名标的',
      value,
      color: DONUT_COLORS[colorIndex % DONUT_COLORS.length]!,
    })
    colorIndex += 1
  }
  if (filled === 0) return []
  // 市值未填但确实有投入的标的单独一块，避免被当成 0% 悄悄吞掉
  if (missing > 0 && unfilledInvested > 0) {
    slices.push({ label: `未填市值（${missing} 个）`, value: unfilledInvested, color: '#cbd5e1' })
  }
  return slices
})

const donutCenter = computed(() => {
  const m = metrics.value
  if (!m || !m.marketValueFilled || !plan.value) return { value: '未填市值', label: '' }
  return { value: formatPlanMoney(m.marketValue, plan.value), label: '当前市值' }
})

function toast(text: string, error = false): void {
  if (error) message.error(text)
  else message.success(text)
}

function describeError(err: unknown): string {
  return err instanceof Error ? err.message : '操作失败'
}

async function reload(): Promise<void> {
  try {
    await loadPlan(planId.value)
    await Promise.all([loadSnapshots(), ensurePresets(planState.plan?.market ?? 'cn')])
  } catch (err) {
    toast(describeError(err), true)
  }
}

onMounted(reload)
watch(planId, reload)
onBeforeUnmount(resetPlanState)

function confirmDiscard(): Promise<boolean> {
  return new Promise((resolve) => {
    dialog.warning({
      title: '有未保存的改动',
      content: '离开后本次修改会丢失，确定要离开吗？',
      positiveText: '放弃并离开',
      negativeText: '继续编辑',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
      onMaskClick: () => resolve(false),
    })
  })
}

onBeforeRouteLeave(async () => {
  if (!planState.dirty) return true
  return confirmDiscard()
})

function onBeforeUnload(event: BeforeUnloadEvent): void {
  if (!planState.dirty) return
  event.preventDefault()
  event.returnValue = ''
}

window.addEventListener('beforeunload', onBeforeUnload)
onBeforeUnmount(() => window.removeEventListener('beforeunload', onBeforeUnload))

async function onSave(): Promise<void> {
  try {
    await savePlan()
    await loadSnapshots()
    // 后端在保存时已把自定义标的收进联想列表，这里只刷新缓存
    await refreshPresets(planState.plan?.market ?? 'cn')
    toast('计划已保存')
  } catch (err) {
    if (err instanceof PlanConflictError) {
      // 别处已经改过这份计划：本地改动先留着，让用户决定是拉取最新还是继续编辑再存一次
      const reloadNow = await new Promise<boolean>((resolve) => {
        dialog.warning({
          title: '这份计划已在别处被修改',
          content:
            '另一个标签页或设备保存过它，你手里的数据已经过期。'
            + '刷新会丢弃本次未保存的改动；继续编辑可以用「放弃修改」手动拉取最新版本后再改。',
          positiveText: '刷新为最新',
          negativeText: '继续编辑',
          onPositiveClick: () => resolve(true),
          onNegativeClick: () => resolve(false),
          onClose: () => resolve(false),
          onMaskClick: () => resolve(false),
        })
      })
      if (!reloadNow) return
      try {
        await reload()
        toast('已刷新为最新版本')
      } catch (reloadErr) {
        toast(describeError(reloadErr), true)
      }
      return
    }
    toast(describeError(err), true)
  }
}

async function onDiscard(): Promise<void> {
  if (!(await confirmDiscard())) return
  try {
    await discardChanges()
    toast('已放弃未保存的改动')
  } catch (err) {
    toast(describeError(err), true)
  }
}

const snapshotDate = ref<string>(todayISO())
const snapshotting = ref(false)

async function onRecordSnapshot(): Promise<void> {
  if (!metrics.value) return
  const enabledCount = metrics.value.enabledAssets.length
  if (metrics.value.marketValueFilled < enabledCount) {
    const missing = enabledCount - metrics.value.marketValueFilled
    const ok = await new Promise<boolean>((resolve) => {
      dialog.warning({
        title: '有标的还没填市值',
        content: `还有 ${missing} 个标的未填当前市值，记录后曲线会偏低。仍要继续吗？`,
        positiveText: '仍然记录',
        negativeText: '先去填写',
        onPositiveClick: () => resolve(true),
        onNegativeClick: () => resolve(false),
        onClose: () => resolve(false),
        onMaskClick: () => resolve(false),
      })
    })
    if (!ok) return
  }
  snapshotting.value = true
  try {
    await recordSnapshot(snapshotDate.value)
    await loadSnapshots()
    toast('已记录当日净值')
  } catch (err) {
    toast(describeError(err), true)
  } finally {
    snapshotting.value = false
  }
}

async function executeAsset(row: PlanAssetMetrics): Promise<void> {
  const amount = Math.round(row.suggestion)
  if (amount <= 0) {
    toast('当前建议金额为 0，无法记录执行', true)
    return
  }
  // 只标记真正买到的那几层：金额被预算封顶时，前面的层可能一分钱都没买，
  // 一起标成已投会让它们以后既不能作废也不能删除。
  const funded = fundedLayers(row, amount)
  const layers = funded.map((r) => r.layer.key).join('、') || '无（仅常规定投）'
  const ok = await new Promise<boolean>((resolve) => {
    dialog.warning({
      title: '记录本次执行',
      content: `标的：${row.asset.name}\n金额：${formatPlanMoney(amount, plan.value!)}\n网格层：${layers}${row.due ? '\n同时推进下一期定投日期' : ''}`,
      positiveText: '确认记录',
      negativeText: '取消',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
      onMaskClick: () => resolve(false),
    })
  })
  if (!ok) return
  try {
    await recordExecution({
      assetId: row.asset.id,
      amount,
      layerKeys: funded.map((r) => r.layer.key),
      wasRegular: row.due,
    })
    await loadSnapshots()
    toast(`已记录 ${row.asset.name} ${formatPlanMoney(amount, plan.value!)}`)
  } catch (err) {
    toast(describeError(err), true)
  }
}

/**
 * 从「最靠内」的待触发层开始往外数，累计金额装得下的层才算真的买到。
 * 与 planMetrics 的 gridAmount 口径一致：highest 只取最新一层，cumulative 取全部。
 */
function fundedLayers(row: PlanAssetMetrics, amount: number): PlanAssetMetrics['rows'] {
  const pending = row.pendingRows
  if (!pending.length) return []
  if (plan.value?.grid_aggregation === 'cumulative') {
    const out: PlanAssetMetrics['rows'] = []
    let sum = 0
    for (const layerRow of [...pending].reverse()) {
      if (sum + layerRow.amount > amount + 0.01) break
      sum += layerRow.amount
      out.unshift(layerRow)
    }
    return out
  }
  const latest = pending[pending.length - 1]
  return latest && latest.amount <= amount + 0.01 ? [latest] : []
}

async function onUndo(): Promise<void> {
  try {
    const exec = await undoLastExecution()
    await loadSnapshots()
    toast(`已撤销 ${exec.asset_name} ${formatPlanMoney(exec.amount, plan.value!)}`)
  } catch (err) {
    toast(describeError(err), true)
  }
}

async function onDeleteAsset(row: PlanAssetMetrics): Promise<void> {  const ok = await new Promise<boolean>((resolve) => {
    dialog.warning({
      title: `删除标的「${row.asset.name}」`,
      content: '删除后需要点「保存」才会生效。',
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
      onMaskClick: () => resolve(false),
    })
  })
  if (ok) removeAsset(row.asset.id)
}

async function onResetHigh(row: PlanAssetMetrics): Promise<void> {
  const ok = await new Promise<boolean>((resolve) => {
    dialog.warning({
      title: `更新「${row.asset.name}」参考高点`,
      content: '参考高点将设为当前点位，所有网格触发点位会随之下移。',
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
      onMaskClick: () => resolve(false),
    })
  })
  if (ok) setReferenceHighToCurrent(row.asset.id)
}

function onCurrencyChange(value: string): void {
  const found = CURRENCIES.find((c) => c.value === value)
  if (found) updateCurrency(found.value, found.name)
}

const CUSTOM_VALUE = '__custom__'
const CUSTOM_PREFIX = '__mine:'

const assetOptions = computed(() => {
  const market = (plan.value?.market ?? 'cn') as 'cn' | 'us'
  const presetGroup = {
    type: 'group',
    label: '预置指数',
    key: 'preset',
    children: indicesFor(market).map((p) => ({
      label: `${p.name}（${p.code}）`,
      value: p.name,
    })),
  }
  const mine = presetsFor(market)
  const groups: Array<SelectGroupOption | SelectOption> = [presetGroup]
  if (mine.length) {
    groups.push({
      type: 'group',
      label: '我添加的',
      key: 'mine',
      children: mine.map((item) => ({
        label: `${item.name}（${item.code || '无代码'}）`,
        value: `${CUSTOM_PREFIX}${item.name}`,
      })),
    })
  }
  groups.push({ label: '自定义…', value: CUSTOM_VALUE })
  return groups
})

/** 下拉里当前该高亮哪一项：预置项直接用名字，自建项用前缀，其余视为自定义。 */
function assetSelectValue(assetId: number): string {
  const asset = plan.value?.assets.find((a) => a.id === assetId)
  if (!asset) return CUSTOM_VALUE
  const market = (plan.value?.market ?? 'cn') as 'cn' | 'us'
  if (indicesFor(market).some((p) => p.name === asset.name)) return asset.name
  if (presetsFor(market).some((p) => p.name === asset.name)) return `${CUSTOM_PREFIX}${asset.name}`
  return CUSTOM_VALUE
}

function onAssetPresetChange(assetId: number, value: string): void {
  const asset = plan.value?.assets.find((a) => a.id === assetId)
  if (!asset || !plan.value) return
  if (value === CUSTOM_VALUE) {
    updateAssetText(assetId, 'name', '')
    return
  }

  const market = plan.value.market as 'cn' | 'us'
  let name = value
  let code = ''
  let leveraged: boolean | undefined

  if (value.startsWith(CUSTOM_PREFIX)) {
    name = value.slice(CUSTOM_PREFIX.length)
    const mine = presetsFor(market).find((p) => p.name === name)
    code = mine?.code ?? ''
    leveraged = mine?.leveraged
  } else {
    const preset = indicesFor(market).find((p) => p.name === value)
    if (!preset) return
    code = preset.code
    leveraged = preset.leveraged
  }

  updateAssetText(assetId, 'name', name)
  updateAssetText(assetId, 'code', code)
  // 一定要显式写回 false：从 2 倍杠杆换到普通标的时，残留的杠杆标记会一直误报。
  updateAssetLeveraged(assetId, leveraged ?? false)

  // 换标的时顺带换成该指数自己的网格；已经有投入的标的保持原样，避免和历史记录对不上。
  if (asset.invested <= 0 && asset.executed_layers.length === 0) {
    const rows = gridForIndex(name, market)
    asset.grid = rows.map(([threshold, allocation], index) => ({
      key: `g${index + 1}`,
      threshold,
      allocation,
      voided: false,
    }))
    toast(`已按「${name}」换成 ${rows.length} 档网格`)
  }
}

async function onRemovePreset(name: string): Promise<void> {
  const market = (plan.value?.market ?? 'cn') as 'cn' | 'us'
  const item = presetsFor(market).find((p) => p.name === name)
  if (!item) return
  if (!(await removePreset(market, item.id))) {
    toast('移除失败', true)
    return
  }
  toast(`已移除自建标的「${name}」`)
}

/** 自建项的选项文字：加「我加的」标签和一个小移除按钮。 */
function renderAssetLabel(option: { value?: string | number; label?: string }): VNodeChild {
  const value = String(option.value ?? '')
  if (!value.startsWith(CUSTOM_PREFIX)) return option.label
  return h('div', { style: 'display:flex;align-items:center;gap:6px;width:100%' }, [
    h('span', { style: 'flex:1;min-width:0;overflow:hidden;text-overflow:ellipsis' }, String(option.label)),
    h('span', { style: 'font-size:11px;color:#0891b2;background:#ecfeff;border-radius:3px;padding:0 4px' }, '我加的'),
    h(
      'span',
      {
        style: 'font-size:11px;color:#ef4444;cursor:pointer',
        onClick: (e: MouseEvent) => {
          e.stopPropagation()
          e.preventDefault()
          void onRemovePreset(value.slice(CUSTOM_PREFIX.length))
        },
      },
      '移除',
    ),
  ])
}

async function onApplyTemplate(key: 'cn' | 'us'): Promise<void> {
  const tpl = findTemplate(key)
  if (!tpl || !plan.value) return
  const ok = await new Promise<boolean>((resolve) => {
    dialog.warning({
      title: `套用${tpl.label}`,
      content: '会替换当前全部标的，并把已投入与市值清零（参考高点也留空）。执行记录和净值快照不受影响。确定继续吗？',
      positiveText: '套用',
      negativeText: '取消',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
      onMaskClick: () => resolve(false),
    })
  })
  if (!ok) return
  applyTemplate(tpl)
  toast(`已套用${tpl.label}，记得点「保存」`)
}

// 每个标的的列表列只建一次：之前每次渲染都重建整套闭包，标的 × 层数越多越卡。
// 渲染时需要的实时状态（是否倒金字塔、当前计划）在 render 里现取，不进闭包。
const assetMetricsById = computed(() => {
  const map = new Map<number, PlanAssetMetrics>()
  for (const row of metrics.value?.assets ?? []) map.set(row.asset.id, row)
  return map
})

const gridColumnsCache = new Map<number, DataTableColumns<PlanAssetMetrics['rows'][number]>>()

function gridColumns(assetId: number): DataTableColumns<PlanAssetMetrics['rows'][number]> {
  const cached = gridColumnsCache.get(assetId)
  if (cached) return cached
  const columns = buildGridColumns(assetId)
  gridColumnsCache.set(assetId, columns)
  return columns
}

function buildGridColumns(assetId: number): DataTableColumns<PlanAssetMetrics['rows'][number]> {
  return [
  { title: '层级', key: 'key', width: 64, render: (r) => r.layer.key },
  {
    title: '回撤阈值 %', key: 'threshold', width: 104,
    render: (r) =>
      h(NInputNumber, {
        value: r.layer.threshold, min: 0, max: 100, size: 'small', disabled: r.state === 'invested',
        onUpdateValue: (v) => updateLayer(assetId, r.layer.key, 'threshold', Number(v ?? 0)),
      }),
  },
  { title: '触发点位', key: 'point', align: 'right', render: (r) => formatNumber(r.point) },
  {
    title: '资金比例 %', key: 'allocation', width: 104,
    render: (r) =>
      h(NInputNumber, {
        value: r.layer.allocation, min: 0, max: 100, size: 'small', disabled: r.state === 'invested',
        onUpdateValue: (v) => updateLayer(assetId, r.layer.key, 'allocation', Number(v ?? 0)),
      }),
  },
  {
    title: '计划金额', key: 'amount', width: 128,
    render: (r) =>
      h(NInputNumber, {
        value: Math.round(r.amount * 100) / 100,
        min: 0,
        size: 'small',
        disabled: r.state === 'invested' || r.voided || !plan.value,
        onUpdateValue: (v) => setLayerAmount(assetId, r.layer.key, Number(v ?? 0)),
      }),
  },
  {
    title: '倒金字塔', key: 'bar', width: 96,
    render: (r) => {
      const strict = assetMetricsById.value.get(assetId)?.strictPyramid ?? false
      return h('div', { style: 'height:6px;border-radius:99px;background:#e5e7eb;overflow:hidden' }, [
        h('i', {
          style: `display:block;height:100%;width:${Math.min(100, r.layer.allocation * 2.5)}%;background:${strict ? '#2563eb' : '#d97706'}`,
        }),
      ])
    },
  },
  {
    title: '状态', key: 'state', width: 76,
    render: (r) => {
      const meta = LAYER_STATE_META[r.state]
      return h(NTag, { size: 'small', type: meta.type }, { default: () => meta.label })
    },
  },
  {
    title: '操作', key: 'actions', width: 200,
    render: (r) =>
      h('div', { style: 'display:flex;gap:4px;flex-wrap:nowrap' }, [
        r.voided
          ? h(NButton, {
              size: 'tiny', quaternary: true, type: 'primary',
              onClick: () => restoreLayer(assetId, r.layer.key),
            }, { default: () => '生效' })
          : h(NButton, {
              size: 'tiny', quaternary: true,
              disabled: r.state === 'invested',
              onClick: () => onVoidLayer(assetId, r),
            }, { default: () => '作废' }),
        h(NButton, {
          size: 'tiny', quaternary: true, type: 'success',
          disabled: r.voided || r.executed,
          onClick: () => onInvestLayer(assetId, r),
        }, { default: () => '已投' }),
        h(NButton, {
          size: 'tiny', quaternary: true, type: 'error',
          disabled: r.executed,
          onClick: () => onRemoveLayer(assetId, r),
        }, { default: () => '删除' }),
      ]),
  },
  ]
}

const LAYER_STATE_META: Record<PlanGridRow['state'], { label: string; type: 'success' | 'info' | 'default' }> = {
  invested: { label: '已投', type: 'info' },
  voided: { label: '历史', type: 'default' },
  triggered: { label: '已触发', type: 'success' },
  waiting: { label: '等待', type: 'default' },
}

async function onVoidLayer(assetId: number, layerRow: PlanGridRow): Promise<void> {
  const asset = assetMetricsById.value.get(assetId)?.asset
  const ok = await new Promise<boolean>((resolve) => {
    dialog.warning({
      title: `作废 ${layerRow.layer.key}`,
      content: `作废后这一层不再触发、不计入资金比例合计，可随时点「生效」恢复。确定作废 ${asset?.name ?? '该标的'} 的第 ${layerRow.index + 1} 层吗？`,
      positiveText: '作废',
      negativeText: '取消',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
      onMaskClick: () => resolve(false),
    })
  })
  if (!ok) return
  if (!voidLayer(assetId, layerRow.layer.key)) toast('该层已投，不能作废', true)
}

async function onRemoveLayer(assetId: number, layerRow: PlanGridRow): Promise<void> {
  const asset = assetMetricsById.value.get(assetId)?.asset
  if (asset?.executed_layers.includes(layerRow.layer.key)) {
    toast('已投的层不能删除，否则执行记录与净值曲线会对不上', true)
    return
  }
  const ok = await new Promise<boolean>((resolve) => {
    dialog.warning({
      title: `删除 ${layerRow.layer.key}`,
      content: `删除后需要点「保存」才会生效。确定删除 ${asset?.name ?? '该标的'} 的第 ${layerRow.index + 1} 层吗？`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
      onMaskClick: () => resolve(false),
    })
  })
  if (ok && !removeGridLayer(assetId, layerRow.layer.key)) toast('该层已投，不能删除', true)
}

async function onInvestLayer(assetId: number, layerRow: PlanGridRow): Promise<void> {
  const asset = assetMetricsById.value.get(assetId)?.asset
  if (!asset) return
  if (planState.dirty) {
    toast('请先保存计划，再记录已投', true)
    return
  }
  if (asset.id <= 0) {
    toast('标的还没保存，请先保存计划', true)
    return
  }
  const suggested = Math.round(layerRow.amount * 100) / 100
  const amountInput = ref<number | null>(suggested)
  const ok = await new Promise<boolean>((resolve) => {
    dialog.warning({
      title: `确认已买入 ${asset.name} · ${layerRow.layer.key}`,
      content: () =>
        h('div', { style: 'display:flex;flex-direction:column;gap:8px' }, [
          h('div', { style: 'font-size:13px;color:#64748b' }, [
            `这一层的计划金额是 ${plan.value ? formatPlanMoney(suggested, plan.value) : suggested}。`,
            '请确认这次买入真的成交了再点确定；确认后会写一条执行记录并累加到已投入。',
          ]),
          h(NInputNumber, {
            value: amountInput.value,
            min: 0,
            style: 'width:100%',
            onUpdateValue: (v) => { amountInput.value = Number(v ?? 0) },
          }),
        ]),
      positiveText: '确认已买入',
      negativeText: '取消',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
      onMaskClick: () => resolve(false),
    })
  })
  if (!ok) return
  const amount = Math.round(Number(amountInput.value) || 0)
  if (amount <= 0) {
    toast('买入金额必须大于 0', true)
    return
  }
  try {
    await recordExecution({
      assetId,
      amount,
      layerKeys: [layerRow.layer.key],
      wasRegular: false,
      note: `网格层 ${layerRow.layer.key} 已投`,
    })
    await loadSnapshots()
    toast(`已记录 ${asset.name} ${plan.value ? formatPlanMoney(amount, plan.value) : amount}`)
  } catch (err) {
    toast(describeError(err), true)
  }
}

function pnlClass(value: number): string {
  if (value > 0) return 'text-red-600'
  if (value < 0) return 'text-emerald-600'
  return 'text-slate-500'
}

/** 权重偏离：超配标红、低配标绿，和盈亏同一套颜色语言。 */
function deviationClass(deviation: number): string {
  if (deviation > 0) return 'text-red-600'
  if (deviation < 0) return 'text-emerald-600'
  return 'text-slate-400'
}

// ---------- 一键优化 ----------

/**
 * 每个标的的目标档数。按「名称 + 代码」记而不是按 id：
 * 保存后后端会重建标的、id 变化，用 id 做键会让用户填的数悄悄退回默认值。
 */
const optimizeCounts = reactive<Record<string, number>>({})

function assetKey(asset: { name: string; code: string }): string {
  return `${asset.name}||${asset.code}`
}

function optimizeLayerCount(asset: GridAssetDTO): number {
  return optimizeCounts[assetKey(asset)] ?? suggestLayerCount(asset)
}

function setOptimizeLayerCount(asset: GridAssetDTO, value: number): void {
  optimizeCounts[assetKey(asset)] = Math.max(MIN_LAYERS, Math.floor(value) || MIN_LAYERS)
}

function formatLayerBrief(layer: GridLayerDTO): string {
  return `${layer.key} 阈值 ${Number(layer.threshold)}% · 比例 ${Number(layer.allocation)}%`
}

async function onOptimizeGrid(row: PlanAssetMetrics): Promise<void> {
  const market = plan.value?.market ?? 'cn'
  const requested = optimizeLayerCount(row.asset)
  let result: GridOptimizeResult
  try {
    result = optimizeGrid(row.asset, market, requested)
  } catch (err) {
    toast(describeError(err), true)
    return
  }

  const anchorText = result.anchorSource === 'drawdown'
    ? `起点取当前回撤 ${result.anchor}%（第一层今天即可买入）`
    : `起点取最深已执行层 ${result.anchor}%，从它往下排，再跌才触发`
  const lines = [
    `${VOLATILITY_LABEL[result.volatility]}标的，排成 ${result.targetLayers} 层，最深 ${result.maxDepth}%。${anchorText}。`,
  ]
  if (result.targetLayers !== requested) {
    lines.push(`目标档数按 ${requested} 层算，已执行层已有 ${result.executedLayers.length} 层，实际排成 ${result.targetLayers} 层。`)
  }
  if (result.executedLayers.length > 0) {
    lines.push(`已执行层不动：${result.executedLayers.map(formatLayerBrief).join('；')}`)
  }
  if (result.voidedLayers.length > 0) {
    lines.push(`作废保留（阈值与金额都不变，可点「生效」恢复）：${result.voidedLayers.map(formatLayerBrief).join('；')}`)
  }
  if (result.alreadyVoidedCount > 0) {
    lines.push(`另有 ${result.alreadyVoidedCount} 层此前已作废，保持不动。`)
  }
  if (result.replacedLayers.length > 0) {
    lines.push(`以下 ${result.replacedLayers.length} 层由新档位取代，同样作废保留以便复盘：${result.replacedLayers.map(formatLayerBrief).join('；')}`)
  }
  if (result.droppedLayers.length > 0) {
    lines.push(`跳过 ${result.droppedLayers.length} 个还没填阈值的空层（不保存）：${result.droppedLayers.map((layer) => layer.key).join('、')}`)
  }
  lines.push(`新排：${result.addedLayers.map(formatLayerBrief).join('；')}`)
  if (result.rowCount > result.targetLayers) {
    lines.push(`优化后表格共 ${result.rowCount} 行（含保留的作废层），后端按行数计上限。`)
  }
  lines.push('改动只在本页生效，确认无误后记得点「保存」。')

  const ok = await new Promise<boolean>((resolve) => {
    dialog.warning({
      title: `一键优化「${row.asset.name}」网格`,
      content: () => h('div', { style: 'display:flex;flex-direction:column;gap:6px;font-size:13px;line-height:1.6' },
        lines.map((line) => h('div', { style: 'color:#475569' }, line))),
      positiveText: '应用',
      negativeText: '取消',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
      onMaskClick: () => resolve(false),
    })
  })
  if (!ok) return
  if (!applyOptimizedGrid(row.asset.id, result.grid)) {
    toast('优化失败：标的已不存在', true)
    return
  }
  toast(`已排成 ${result.targetLayers} 层，记得点「保存」`)
}
</script>

<template>
  <div class="max-w-[1500px] mx-auto">
    <div class="flex flex-wrap items-center justify-between gap-3 mb-4">
      <div class="min-w-0">
        <div class="flex items-center gap-2 flex-wrap">
          <NButton size="small" quaternary @click="router.push({ name: 'grid-plans' })">← 返回列表</NButton>
          <NTag v-if="planState.dirty" size="small" type="warning">未保存</NTag>
          <NTag v-if="planState.conflict" size="small" type="error">版本已过期</NTag>
        </div>
        <h1 class="text-2xl font-bold m-0 mt-2">{{ plan?.name || '计划' }}</h1>
        <p class="text-sm text-slate-500 mt-1 mb-0">
          {{ plan ? `${plan.currency_name}计算 · ${plan.assets.length} 个标的` : '加载中…' }}
        </p>
      </div>
      <div class="flex gap-2">
        <NButton size="small" :disabled="!planState.dirty || planState.saving" @click="onDiscard">放弃改动</NButton>
        <NButton size="small" type="primary" :loading="planState.saving" :disabled="!planState.dirty" @click="onSave">
          保存
        </NButton>
      </div>
    </div>

    <NSpin :show="planState.loading">
      <!-- 保存请求在途时锁住编辑区：保存成功后会用服务器返回值覆盖本地，期间的改动会被静默丢掉 -->
      <div :class="planState.saving ? 'pointer-events-none opacity-70' : ''" :aria-busy="planState.saving">
      <NResult v-if="planState.error && !plan" status="error" title="计划加载失败" :description="planState.error">
        <template #footer>
          <NButton @click="reload">重试</NButton>
        </template>
      </NResult>

      <template v-else-if="plan && metrics">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3 mb-4">
          <NCard size="small">
            <div class="text-xs text-slate-500">总预算</div>
            <div class="text-lg font-semibold mt-1">{{ formatPlanMoney(plan.total_budget, plan) }}</div>
            <div class="text-xs text-slate-500 mt-1">
              已投入 {{ formatPlanMoney(metrics.invested, plan) }}
            </div>
          </NCard>
          <NCard size="small">
            <div class="text-xs text-slate-500">当前市值</div>
            <div class="text-lg font-semibold mt-1">
              {{ metrics.marketValueFilled ? formatPlanMoney(metrics.marketValue, plan) : '未填' }}
            </div>
            <div class="text-xs text-slate-500 mt-1">
              {{ metrics.marketValueFilled }} / {{ metrics.enabledAssets.length }} 个标的已填
            </div>
            <div v-if="metrics.assets.length > metrics.enabledAssets.length" class="text-xs text-slate-400 mt-0.5">
              另有 {{ metrics.assets.length - metrics.enabledAssets.length }} 个已停用，不计入账目
            </div>
          </NCard>
          <NCard size="small">
            <div class="text-xs text-slate-500">总盈亏</div>
            <div class="text-lg font-semibold mt-1" :class="pnlClass(metrics.pnl)">
              {{ metrics.marketValueFilled ? formatSignedMoney(metrics.pnl, plan) : '—' }}
            </div>
            <div class="text-xs mt-1" :class="pnlClass(metrics.pnl)">
              收益率 {{ formatPercent(metrics.returnRate) }}
            </div>
            <div v-if="metrics.unfilledInvested > 0" class="text-xs text-amber-600 mt-0.5">
              另有 {{ formatPlanMoney(metrics.unfilledInvested, plan) }} 未填市值
            </div>
          </NCard>
          <NCard size="small">
            <div class="text-xs text-slate-500">组合加权回撤</div>
            <div class="text-lg font-semibold mt-1">{{ formatPlainPercent(metrics.drawdown) }}</div>
            <div class="text-xs text-slate-500 mt-1">
              {{ plan.weight_method === 'target-weight' ? '目标权重加权' : '实际投入加权' }}
            </div>
          </NCard>
          <NCard size="small">
            <div class="text-xs text-slate-500">本期建议合计</div>
            <div class="text-lg font-semibold mt-1">{{ formatPlanMoney(metrics.suggestionTotal, plan) }}</div>
            <div class="text-xs text-slate-500 mt-1">理论 {{ formatPlanMoney(metrics.theoreticalTotal, plan) }}</div>
          </NCard>
        </div>

        <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_300px] gap-4 items-start">
          <div class="grid grid-cols-1 md:grid-cols-[280px_minmax(0,1fr)] gap-4 xl:col-start-1 xl:row-start-1">
            <NCard size="small" title="资产占比">
              <DonutChart
                :slices="donutSlices"
                :center-value="donutCenter.value"
                :center-label="donutCenter.label"
                :size="180"
              />
              <div v-if="!donutSlices.length" class="text-xs text-slate-400 mt-3">
                填入标的「当前市值」后，这里会显示各标的的占比。
              </div>
            </NCard>

            <NCard size="small" title="收益变化曲线">
              <template #header-extra>
                <div class="flex flex-wrap items-center gap-2">
                  <span class="text-xs" :class="snapshotStale ? 'text-amber-600' : 'text-slate-400'">
                    {{ lastSnapshotLabel }}
                  </span>
                  <NDatePicker v-model:formatted-value="snapshotDate" type="date" value-format="yyyy-MM-dd"
                    size="small" style="width:150px" />
                  <NTooltip :disabled="!recordBlocked">
                    <template #trigger>
                      <span>
                        <NButton size="small" type="primary" :loading="snapshotting" :disabled="recordBlocked"
                          @click="onRecordSnapshot">记录当日净值</NButton>
                      </span>
                    </template>
                    请先保存计划，再记录净值
                  </NTooltip>
                </div>
              </template>
              <div class="flex flex-wrap items-center gap-1.5 mb-2">
                <NButton
                  v-for="opt in SNAPSHOT_RANGES"
                  :key="opt.days"
                  size="tiny"
                  :type="rangeDays === opt.days ? 'primary' : 'default'"
                  :ghost="rangeDays !== opt.days"
                  @click="rangeDays = opt.days"
                >{{ opt.label }}</NButton>
                <span class="text-xs text-slate-400 ml-1">{{ snapshots.length }} 个记录点</span>
              </div>
              <NSpin :show="snapshotLoading">
                <LineChart :points="chartPoints" :height="260" empty-hint="点右上角「记录当日净值」，这里才会画出投入与市值曲线" />
              </NSpin>
              <div class="text-xs text-slate-400 mt-2">
                曲线只在你点「记录当日净值」时留点；同一天重复记录只保存最后一次。
                当前显示：{{ rangeLabel }}。
              </div>
            </NCard>
          </div>

          <NCard size="small" title="计划参数" class="xl:col-start-2 xl:row-start-1 xl:row-span-2">
            <NFormItem label="计划名称" size="small">
              <NInput :value="plan.name" maxlength="80" @update:value="updatePlanName" />
            </NFormItem>
            <div class="grid grid-cols-2 gap-2">
              <NFormItem label="币种" size="small">
                <NSelect :value="plan.currency"
                  :options="CURRENCIES.map((c) => ({ value: c.value, label: c.label }))"
                  style="width:100%"
                  @update:value="(v) => onCurrencyChange(v)" />
              </NFormItem>
              <NFormItem label="总预算" size="small">
                <NInputNumber :value="plan.total_budget" :min="0" :step="plan.unit" style="width:100%"
                  @update:value="(v) => updateBudget(Number(v ?? 0))" />
              </NFormItem>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <NFormItem label="定投频率" size="small">
                <NSelect :value="plan.frequency"
                  :options="Object.entries(FREQUENCY_LABEL).map(([value, label]) => ({ value, label }))"
                  @update:value="(v) => updateFrequency(v as Frequency)" />
              </NFormItem>
              <NFormItem label="下次定投日" size="small">
                <NDatePicker type="date" :formatted-value="plan.next_date || null"
                  :value="plan.next_date ? new Date(`${plan.next_date}T00:00:00`).getTime() : null"
                  value-format="yyyy-MM-dd" style="width:100%"
                  @update:formatted-value="(v) => updateNextDate(String(v ?? ''))" />
              </NFormItem>
            </div>
            <NFormItem label="已投入未分配金额" size="small">
              <NInputNumber :value="plan.unallocated_invested" :min="0" style="width:100%"
                @update:value="(v) => updateUnallocated(Number(v ?? 0))" />
            </NFormItem>
            <NFormItem label="公开范围" size="small">
              <NSelect :value="plan.visibility" :options="VISIBILITY_OPTIONS"
                @update:value="(v) => updateVisibility(v as PlanVisibility)" />
            </NFormItem>
            <NFormItem v-if="plan.visibility === 'specified'" label="指定可见好友" size="small">
              <FriendMultiSelect v-model="visibleUserIds" />
            </NFormItem>
            <p v-if="plan.visibility !== 'private'" class="text-xs text-slate-400 -mt-2 mb-3">
              还要在「主页定制」里打开「对访客展示投资计划」，个人主页才会出现这个页签。
            </p>
            <NFormItem label="组合回撤闸门阈值 %" size="small">
              <NInputNumber :value="plan.gate_threshold" :min="0" :max="100" style="width:100%"
                @update:value="(v) => updateGateThreshold(Number(v ?? 0))" />
            </NFormItem>
            <NFormItem label="闸门模式" size="small">
              <NSelect :value="plan.gate_mode"
                :options="[
                  { value: 'alert', label: '仅提醒' },
                  { value: 'grid-only', label: '仅允许网格加仓' },
                  { value: 'block-all', label: '暂停全部新增投入' },
                ]"
                @update:value="(v) => updateGateMode(v as GateMode)" />
            </NFormItem>
            <NFormItem label="组合回撤加权口径" size="small">
              <NSelect :value="plan.weight_method"
                :options="[
                  { value: 'target-weight', label: '目标权重加权' },
                  { value: 'invested-weight', label: '实际投入加权' },
                ]"
                @update:value="(v) => updateWeightMethod(v as WeightMethod)" />
            </NFormItem>
            <NFormItem label="网格聚合方式" size="small">
              <NSelect :value="plan.grid_aggregation"
                :options="[
                  { value: 'highest', label: '仅最新一层' },
                  { value: 'cumulative', label: '累加所有未执行层' },
                ]"
                @update:value="(v) => updateAggregation(v as 'highest' | 'cumulative')" />
            </NFormItem>
            <div class="mt-3 pt-3 border-t border-slate-100">
              <div class="text-xs text-slate-500 mb-2">初始化模板（会替换全部标的并清零投入/市值）</div>
              <div class="flex flex-wrap gap-2">
                <NButton v-for="tpl in PLAN_TEMPLATES" :key="tpl.key" size="small"
                  @click="onApplyTemplate(tpl.key)">{{ tpl.label }}</NButton>
              </div>
            </div>
          </NCard>

          <div class="flex flex-col gap-3 xl:col-start-1 xl:row-start-2">
            <!-- 新增标的是最高频操作，放在标的列表最上方，不用滚到右侧去找 -->
            <NCard size="small" :bordered="false" class="!bg-slate-50">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div class="min-w-0">
                  <div class="text-sm font-medium text-slate-700">
                    标的（{{ metrics.assets.length }}）
                  </div>
                  <div class="text-xs text-slate-400 mt-0.5">
                    新增后填名称、点位与权重，再点「保存」落库
                  </div>
                </div>
                <div class="flex gap-2">
                  <NButton ghost @click="onUndo">撤销最近执行</NButton>
                  <NButton type="primary" @click="addAsset()">+ 新增标的</NButton>
                </div>
              </div>
            </NCard>
            <NCard v-for="row in metrics.assets" :key="row.asset.id" size="small">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="text-base font-semibold">{{ row.asset.name || '未命名标的' }}</span>
                    <span class="text-xs text-slate-400 font-mono">{{ row.asset.code }}</span>
                    <NTag size="small" :type="statusMeta[row.status].type">{{ statusMeta[row.status].label }}</NTag>
                    <NTag v-if="row.asset.leveraged" size="small" type="error">2× 杠杆</NTag>
                    <NTag v-if="!row.asset.enabled" size="small">已停用</NTag>
                    <NTag v-if="!row.hasMarketValue" size="small" type="warning">未填市值</NTag>
                  </div>
                  <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mt-3">
                    <div>
                      <div class="text-xs text-slate-500">当前 / 参考高点</div>
                      <div class="text-sm font-medium">{{ formatNumber(row.asset.current_point) }} / {{ formatNumber(row.asset.high_point) }}</div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-500">当前回撤</div>
                      <div class="text-sm font-medium">{{ formatPlainPercent(row.dd) }}</div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-500">目标权重</div>
                      <div class="text-sm font-medium">{{ formatPlainPercent(row.asset.weight) }}</div>
                      <div class="text-xs text-slate-400 mt-0.5">
                        应投 {{ formatPlanMoney(row.targetBudget, plan) }}
                      </div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-500">已投入占比</div>
                      <div class="text-sm font-medium">{{ formatPlainPercent(row.actualWeight) }}</div>
                      <div v-if="row.weightDeviation !== null" class="text-xs mt-0.5" :class="deviationClass(row.weightDeviation)">
                        {{ row.weightDeviation > 0 ? '超配' : row.weightDeviation < 0 ? '低配' : '持平' }}
                        {{ formatPlainPercent(Math.abs(row.weightDeviation)) }}
                      </div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-500">已投入 / 市值</div>
                      <div class="text-sm font-medium">
                        {{ formatPlanMoney(row.asset.invested, plan) }} /
                        {{ row.hasMarketValue ? formatPlanMoney(row.asset.market_value, plan) : '—' }}
                      </div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-500">盈亏</div>
                      <div class="text-sm font-medium" :class="pnlClass(row.pnl)">
                        {{ row.hasMarketValue ? formatSignedMoney(row.pnl, plan) : '未填市值' }}
                      </div>
                      <div v-if="row.hasMarketValue" class="text-xs" :class="pnlClass(row.pnl)">
                        {{ formatPercent(row.returnRate) }}
                      </div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-500">目标预算余量</div>
                      <div class="text-sm font-medium">{{ formatPlanMoney(row.remainingIndex, plan) }}</div>
                    </div>
                  </div>
                </div>
                <div class="text-right ml-auto">
                  <div class="text-xs text-slate-500">当前建议投入</div>
                  <div class="text-xl font-bold text-indigo-600">{{ formatPlanMoney(row.suggestion, plan) }}</div>
                  <div class="text-xs text-slate-500 mt-1 max-w-[260px]">{{ row.reason }}</div>
                </div>
              </div>

              <NCard size="small" class="mt-3" :bordered="false">
                <template #header>
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-medium">参数与网格</span>
                    <NTag size="tiny">{{ row.rows.length }} 层</NTag>
                  </div>
                </template>
                <div class="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-5 gap-2 items-start">
                  <NFormItem label="标的" size="small">
                    <NSelect
                      :value="assetSelectValue(row.asset.id)"
                      :options="assetOptions"
                      :render-label="renderAssetLabel"
                      filterable
                      @update:value="(v) => onAssetPresetChange(row.asset.id, String(v))"
                    />
                  </NFormItem>
                  <NFormItem v-if="assetSelectValue(row.asset.id) === CUSTOM_VALUE" label="名称" size="small">
                    <NInput :value="row.asset.name" placeholder="自定义名称"
                      @update:value="(v) => updateAssetText(row.asset.id, 'name', v)" />
                  </NFormItem>
                  <NFormItem label="代码 / 参考标的" size="small">
                    <NInput :value="row.asset.code" @update:value="(v) => updateAssetText(row.asset.id, 'code', v)" />
                  </NFormItem>
                  <NFormItem label="当前点位" size="small">
                    <NInputNumber :value="row.asset.current_point" :min="0" style="width:100%"
                      @update:value="(v) => updateAssetNumber(row.asset.id, 'current_point', v === null ? null : Number(v))" />
                  </NFormItem>
                  <NFormItem label="参考高点" size="small">
                    <NInputNumber :value="row.asset.high_point" :min="0" style="width:100%"
                      @update:value="(v) => updateAssetNumber(row.asset.id, 'high_point', v === null ? null : Number(v))" />
                  </NFormItem>
                  <NFormItem label="目标权重 %" size="small">
                    <NInputNumber :value="row.asset.weight" :min="0" :max="100" :step="0.1" style="width:100%"
                      @update:value="(v) => updateAssetNumber(row.asset.id, 'weight', Number(v ?? 0))" />
                  </NFormItem>
                  <NFormItem label="已投入" size="small">
                    <NInputNumber :value="row.asset.invested" :min="0" style="width:100%"
                      @update:value="(v) => updateAssetNumber(row.asset.id, 'invested', Number(v ?? 0))" />
                  </NFormItem>
                  <NFormItem label="当前市值" size="small">
                    <NInputNumber :value="row.asset.market_value" :min="0" style="width:100%"
                      @update:value="(v) => updateAssetNumber(row.asset.id, 'market_value', Number(v ?? 0))" />
                  </NFormItem>
                  <NFormItem label="常规定投额" size="small">
                    <NInputNumber :value="row.asset.base_amount" :min="0" style="width:100%"
                      @update:value="(v) => updateAssetNumber(row.asset.id, 'base_amount', Number(v ?? 0))" />
                  </NFormItem>
                  <NFormItem label="目标预算" size="small">
                    <div class="w-full">
                      <NInputNumber
                        :value="Math.round(row.targetBudget * 100) / 100"
                        :min="0"
                        :step="plan.unit"
                        :disabled="plan.total_budget <= 0"
                        :show-button="false"
                        :prefix="plan.currency"
                        style="width:100%"
                        @update:value="(v) => updateTargetBudget(row.asset.id, Number(v ?? 0))"
                      />
                      <p v-if="plan.total_budget <= 0" class="text-xs text-slate-400 mt-1">
                        先填计划总预算，才能按金额反推权重
                      </p>
                      <p v-else class="text-xs text-slate-400 mt-1">
                        改金额会反算权重（6 位小数），超上限按剩余额度取最大值
                      </p>
                    </div>
                  </NFormItem>
                  <NFormItem label="参与计算" size="small">
                    <div class="flex items-center gap-2 h-7">
                      <NSwitch size="small" :value="row.asset.enabled"
                        :theme-overrides="{ railColor: '#94a3b8' }"
                        @update:value="(v) => updateAssetEnabled(row.asset.id, v)" />
                      <span class="text-xs" :class="row.asset.enabled ? 'text-slate-500' : 'text-slate-400'">
                        {{ row.asset.enabled ? '计入' : '已停用' }}
                      </span>
                    </div>
                  </NFormItem>
                </div>

                <NDataTable size="small" :bordered="true" :single-line="false"
                  :columns="gridColumns(row.asset.id)" :data="row.rows" :row-key="(r: PlanAssetMetrics['rows'][number]) => r.layer.key"
                  :row-class-name="(r: PlanAssetMetrics['rows'][number]) => (r.voided ? 'grid-row-voided' : '')"
                  :pagination="false" />
                <div class="flex flex-wrap justify-between gap-2 text-xs text-slate-500 mt-2">
                  <span>网格资金比例合计：<b class="text-slate-700">{{ formatPlainPercent(row.gridRatioTotal) }}</b></span>
                  <span v-if="row.voidedCount" class="text-slate-400">
                    已作废 {{ row.voidedCount }} 层（保留历史阈值，不参与触发与合计，可点「生效」恢复）
                  </span>
                  <span v-if="!row.gridRatioValid" class="text-red-600">比例超过 100%，需调整</span>
                  <span v-else-if="row.gridRatioTotal < 100">保留 {{ formatPlainPercent(100 - row.gridRatioTotal) }} 未分配</span>
                  <span v-else>资金比例完整</span>
                  <span v-if="!row.strictPyramid" class="text-amber-600">当前不是严格倒金字塔</span>
                </div>

                <div class="flex flex-wrap gap-2 mt-3">
                  <NButton size="small" @click="addGridLayer(row.asset.id)">新增网格层</NButton>
                  <NButton size="small" type="primary" :disabled="row.suggestion <= 0" @click="executeAsset(row)">
                    记录本期执行
                  </NButton>
                  <NButton size="small" ghost @click="onResetHigh(row)">设为新参考高点</NButton>
                  <NButton size="small" ghost :disabled="!row.asset.executed_layers.length" @click="clearExecutedLayers(row.asset.id)">
                    清除执行状态
                  </NButton>
                  <NButton size="small" ghost type="error" @click="onDeleteAsset(row)">删除标的</NButton>
                </div>

                <div class="flex flex-wrap items-center gap-2 mt-3 pt-3 border-t border-slate-200">
                  <span class="text-xs text-slate-500">一键优化</span>
                  <NInputNumber
                    :value="optimizeLayerCount(row.asset)"
                    :min="MIN_LAYERS"
                    :max="20"
                    size="tiny"
                    style="width:88px"
                    @update:value="(v) => setOptimizeLayerCount(row.asset, Number(v ?? 0))"
                  />
                  <span class="text-xs text-slate-400">层（{{ VOLATILITY_LABEL[volatilityOf(row.asset)] }}默认 {{ suggestLayerCount(row.asset) }} 层）</span>
                  <NButton size="small" ghost type="warning" @click="onOptimizeGrid(row)">一键优化</NButton>
                  <span class="text-xs text-slate-400">
                    浅层没买过的作废，已执行层不动，钱排到更深的档位
                  </span>
                </div>
              </NCard>
            </NCard>

            <NCard v-if="!metrics.assets.length" size="small">
              <div class="text-center text-sm text-slate-500 py-6">还没有标的，点上方「+ 新增标的」开始。</div>
            </NCard>

            <NCard size="small" title="行动与风险提醒">
              <div class="flex flex-col gap-2">
                <div v-for="alert in metrics.alerts" :key="alert.id"
                  class="rounded-md border px-3 py-2 text-sm"
                  :class="{
                    'border-red-200 bg-red-50': alert.level === 'error',
                    'border-amber-200 bg-amber-50': alert.level === 'warn',
                    'border-slate-200 bg-slate-50': alert.level === 'info' || alert.level === 'ok',
                  }">
                  <div class="font-medium">{{ alert.title }}</div>
                  <div class="text-xs text-slate-500 mt-1">{{ alert.detail }}</div>
                </div>
              </div>
            </NCard>
          </div>
        </div>
      </template>
      </div>
    </NSpin>
  </div>
</template>

<style scoped>
:deep(.grid-row-voided) td {
  opacity: 0.45;
  background: #f8fafc;
}
</style>
