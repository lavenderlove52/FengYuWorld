<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NAlert, NButton, NCard, NEmpty, NFormItem, NInput, NModal, NRadio, NRadioGroup, NSpin, NTag, useDialog, useMessage } from 'naive-ui'
import { createGridPlan, deleteGridPlan, listGridPlans } from '@/api/gridPlans'
import { formatPlanMoney, formatSignedMoney } from '@/core/gridplan/planFormat'
import { migrateLegacyPlans, readLegacyState } from '@/core/gridplan/legacyMigration'
import { findTemplate, PLAN_TEMPLATES, templatePlanPayload } from '@/core/gridplan/templates'
import type { GridPlanSummary } from '@/types/gridPlan'

const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const plans = ref<GridPlanSummary[]>([])
const loading = ref(false)
const creating = ref(false)
const createOpen = ref(false)
const newName = ref('')
const nameTouched = ref(false)
const newTemplate = ref<'blank' | 'cn' | 'us'>('cn')
const legacyFound = ref(0)
const migrating = ref(false)

async function refresh(): Promise<void> {
  loading.value = true
  try {
    const res = await listGridPlans()
    if (res.code !== 0) {
      message.error(res.message || '加载计划失败')
      return
    }
    plans.value = res.data ?? []
  } catch (err) {
    message.error(err instanceof Error ? err.message : '加载计划失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  legacyFound.value = Object.keys(readLegacyState() ?? {}).length
  refresh()
})

async function onMigrate(): Promise<void> {
  migrating.value = true
  try {
    const result = await migrateLegacyPlans()
    if (result.ok) {
      message.success(result.message)
    } else {
      message.warning(result.message)
    }
    legacyFound.value = Object.keys(readLegacyState() ?? {}).length
    await refresh()
  } catch (err) {
    message.error(err instanceof Error ? err.message : '导入旧计划失败')
  } finally {
    migrating.value = false
  }
}

function openCreate(): void {
  newName.value = ''
  nameTouched.value = false
  createOpen.value = true
}


const nameError = computed(() => (nameTouched.value && !newName.value.trim() ? '请填写计划名称' : ''))
const canCreate = computed(() => newName.value.trim().length > 0)

async function submitCreate(): Promise<void> {
  const name = newName.value.trim()
  if (!name) {
    nameTouched.value = true
    return
  }
  creating.value = true
  try {
    const template = newTemplate.value === 'blank' ? null : findTemplate(newTemplate.value)
    const res = await createGridPlan(
      template
        ? templatePlanPayload(template, name)
        : { name, unit: 100, total_budget: 0, assets: [] },
    )
    if (res.code !== 0 || !res.data) {
      message.error(res.message || '新建计划失败')
      return
    }
    createOpen.value = false
    router.push({ name: 'grid-plan-detail', params: { id: res.data.id } })
  } catch (err) {
    message.error(err instanceof Error ? err.message : '新建计划失败')
  } finally {
    creating.value = false
  }
}

function confirmDelete(plan: GridPlanSummary): void {
  dialog.warning({
    title: `删除计划「${plan.name}」`,
    content: '计划下的标的、净值快照和执行记录会一起删除，无法恢复。',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await deleteGridPlan(plan.id)
        if (res.code !== 0) {
          message.error(res.message || '删除失败')
          return
        }
        message.success('计划已删除')
        await refresh()
      } catch (err) {
        message.error(err instanceof Error ? err.message : '删除失败')
      }
    },
  })
}

function openPlan(plan: GridPlanSummary): void {
  router.push({ name: 'grid-plan-detail', params: { id: plan.id } })
}

function pnlClass(value: number): string {
  if (value > 0) return 'text-red-600'
  if (value < 0) return 'text-emerald-600'
  return 'text-slate-500'
}

/** 列表卡片上的可见范围标记；private 不显示，避免每张卡都挂个「仅自己」。 */
function visibilityTag(visibility: GridPlanSummary['visibility']): {
  label: string
  type: 'default' | 'info' | 'success' | 'warning'
} {
  switch (visibility) {
    case 'public':
      return { label: '公开', type: 'success' }
    case 'login':
      return { label: '登录可见', type: 'info' }
    case 'friends':
      return { label: '好友可见', type: 'info' }
    case 'specified':
      return { label: '指定查看', type: 'warning' }
    default:
      return { label: '仅自己', type: 'default' }
  }
}
</script>

<template>
  <div class="max-w-[1200px] mx-auto">
    <div class="flex flex-wrap items-center justify-between gap-3 mb-4">
      <div>
        <h1 class="text-2xl font-bold m-0">定投计划</h1>
        <p class="text-sm text-slate-500 mt-1 mb-0">
          每个计划是一份独立的资金池：自己的预算、权重、网格与回撤闸门。
        </p>
      </div>
      <NButton type="primary" @click="openCreate">新建计划</NButton>
    </div>

    <NAlert v-if="legacyFound" type="info" class="mb-4" title="检测到旧版本地计划">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="text-sm">
          浏览器里还存着 {{ legacyFound }} 份旧版计划（A 股 / 美股），可以一键导入到账号下。重复点击不会重复创建。
        </span>
        <NButton size="small" type="primary" :loading="migrating" @click="onMigrate">导入旧计划</NButton>
      </div>
    </NAlert>

    <NSpin :show="loading">
      <div v-if="plans.length" class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <NCard v-for="plan in plans" :key="plan.id" size="small" hoverable class="cursor-pointer" @click="openPlan(plan)">
          <template #header>
            <div class="flex items-center gap-2 min-w-0">
              <span class="font-semibold truncate">{{ plan.name }}</span>
              <NTag size="tiny">{{ plan.currency_name }}</NTag>
              <NTag v-if="plan.visibility !== 'private'" size="tiny" :type="visibilityTag(plan.visibility).type">
                {{ visibilityTag(plan.visibility).label }}
              </NTag>
            </div>
          </template>
          <template #header-extra>
            <NButton size="tiny" quaternary type="error" @click.stop="confirmDelete(plan)">删除</NButton>
          </template>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <div class="text-xs text-slate-500">总预算</div>
              <div class="text-base font-semibold mt-1">{{ formatPlanMoney(plan.total_budget, plan) }}</div>
            </div>
            <div>
              <div class="text-xs text-slate-500">已投入</div>
              <div class="text-base font-semibold mt-1">{{ formatPlanMoney(plan.invested, plan) }}</div>
            </div>
            <div>
              <div class="text-xs text-slate-500">当前市值</div>
              <div class="text-base font-semibold mt-1">
                {{ plan.market_value > 0 ? formatPlanMoney(plan.market_value, plan) : '未填' }}
              </div>
            </div>
            <div>
              <div class="text-xs text-slate-500">盈亏</div>
              <div class="text-base font-semibold mt-1" :class="pnlClass(plan.pnl)">
                {{ plan.market_value > 0 ? formatSignedMoney(plan.pnl, plan) : '—' }}
              </div>
            </div>
          </div>

          <div class="flex items-center justify-between text-xs text-slate-400 mt-3 pt-3 border-t border-slate-100">
            <span>{{ plan.asset_count }} 个标的</span>
            <span>更新于 {{ plan.updated_at }}</span>
          </div>
        </NCard>
      </div>

      <NCard v-else-if="!loading" size="small">
        <NEmpty description="还没有投资计划">
          <template #extra>
            <NButton type="primary" @click="openCreate">新建第一个计划</NButton>
          </template>
        </NEmpty>
      </NCard>
    </NSpin>

    <NModal
      v-model:show="createOpen"
      preset="dialog"
      title="新建投资计划"
      positive-text="创建"
      negative-text="取消"
      :loading="creating"
      :positive-button-props="{ disabled: !canCreate }"
      @positive-click="submitCreate"
    >
      <div class="flex flex-col gap-3">
        <NFormItem :show-label="false" :validation-status="nameError ? 'error' : undefined" :feedback="nameError">
          <NInput
            v-model:value="newName"
            placeholder="例如：A 股网格 / 美股指数"
            maxlength="80"
            show-count
            autofocus
            :status="nameError ? 'error' : undefined"
            @blur="nameTouched = true"
            @keydown.enter="submitCreate"
          />
        </NFormItem>
        <div class="text-xs text-slate-500">选择模板可一次建好标的、权重和 10 档网格；稍后也能在详情页改。</div>
        <NRadioGroup v-model:value="newTemplate">
          <div class="flex flex-col gap-2">
            <NRadio value="blank">
              <div class="text-sm">空白计划</div>
              <div class="text-xs text-slate-400">不预置任何标的</div>
            </NRadio>
            <NRadio v-for="tpl in PLAN_TEMPLATES" :key="tpl.key" :value="tpl.key">
              <div class="text-sm">{{ tpl.label }}</div>
              <div class="text-xs text-slate-400">{{ tpl.description }}</div>
            </NRadio>
          </div>
        </NRadioGroup>
      </div>
    </NModal>
  </div>
</template>
