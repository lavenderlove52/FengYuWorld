import http from './http'
import type { ApiResponse } from '@/types/auth'
import type { GridAssetPreset } from '@/types/gridPlan'

export const listGridPresets = (market: string) =>
  http
    .get<ApiResponse<GridAssetPreset[]>>('/grid-presets', { params: { market } })
    .then((r) => r.data)

export const deleteGridPreset = (id: number) =>
  http.delete<ApiResponse<null>>(`/grid-presets/${id}`).then((r) => r.data)
