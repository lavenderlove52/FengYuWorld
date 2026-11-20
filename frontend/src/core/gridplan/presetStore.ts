import { reactive, readonly } from 'vue'
import { deleteGridPreset, listGridPresets } from '@/api/gridPresets'
import type { GridAssetPreset } from '@/types/gridPlan'

/**
 * 账号级自建标的的本地缓存。
 * 后端在保存计划时自动收集，这里只负责拉取与本地删除，避免详情页每次打开都打接口。
 */

interface PresetState {
  cn: GridAssetPreset[]
  us: GridAssetPreset[]
  loading: Record<'cn' | 'us', boolean>
}

const state = reactive<PresetState>({
  cn: [],
  us: [],
  loading: { cn: false, us: false },
})

export const presetState = readonly(state)

function bucketOf(market: string): 'cn' | 'us' {
  return market === 'us' ? 'us' : 'cn'
}

/** 拉一次并覆盖缓存。失败时保留旧数据，下拉退化成只有预置项。 */
export async function refreshPresets(market: string): Promise<void> {
  const bucket = bucketOf(market)
  state.loading[bucket] = true
  try {
    const res = await listGridPresets(bucket)
    if (res.code === 0) state[bucket] = res.data ?? []
  } catch {
    // 联想只是便利功能，失败不打断主流程
  } finally {
    state.loading[bucket] = false
  }
}

/** 本地已有的就不用再打接口。 */
export async function ensurePresets(market: string): Promise<void> {
  const bucket = bucketOf(market)
  if (state.loading[bucket]) return
  if (state[bucket].length > 0) return
  await refreshPresets(bucket)
}

export async function removePreset(market: string, id: number): Promise<boolean> {
  const bucket = bucketOf(market)
  try {
    const res = await deleteGridPreset(id)
    // 404 说明后端已经没有了，本地也当删除成功，不要让下拉一直留着一个死项
    if (res.code !== 0 && res.code !== 40400) return false
  } catch {
    return false
  }
  state[bucket] = state[bucket].filter((item) => item.id !== id)
  return true
}

export function presetsFor(market: string): GridAssetPreset[] {
  return state[bucketOf(market)]
}

export function findPreset(market: string, name: string): GridAssetPreset | undefined {
  return state[bucketOf(market)].find((item) => item.name === name)
}

export function clearPresets(): void {
  state.cn = []
  state.us = []
}
