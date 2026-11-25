import http from './http'
import type { ApiResponse } from '@/types/auth'
import type {
  GridExecution,
  GridPlanDetail,
  GridPlanSummary,
  GridSnapshot,
  RecordExecutionPayload,
  SaveGridPlanPayload,
} from '@/types/gridPlan'

export const listGridPlans = () =>
  http.get<ApiResponse<GridPlanSummary[]>>('/grid-plans').then((r) => r.data)

export const getGridPlan = (id: number) =>
  http.get<ApiResponse<GridPlanDetail>>(`/grid-plans/${id}`).then((r) => r.data)

export const createGridPlan = (payload: SaveGridPlanPayload) =>
  http.post<ApiResponse<GridPlanDetail>>('/grid-plans', payload).then((r) => r.data)

export const updateGridPlan = (id: number, payload: SaveGridPlanPayload) =>
  http.put<ApiResponse<GridPlanDetail>>(`/grid-plans/${id}`, payload).then((r) => r.data)

export const deleteGridPlan = (id: number) =>
  http.delete<ApiResponse<null>>(`/grid-plans/${id}`).then((r) => r.data)

/** days=0 表示全部历史；缺省返回最近 180 天。 */
export const listGridSnapshots = (id: number, days = 180) =>
  http
    .get<ApiResponse<GridSnapshot[]>>(`/grid-plans/${id}/snapshots`, { params: { days } })
    .then((r) => r.data)

/** date 省略时后端按服务器当天算；同一天重复记录只保留最后一次。 */
export const recordGridSnapshot = (id: number, date?: string) =>
  http.post<ApiResponse<GridSnapshot>>(`/grid-plans/${id}/snapshots`, date ? { date } : {}).then((r) => r.data)

export const addGridExecution = (id: number, payload: RecordExecutionPayload) =>
  http.post<ApiResponse<GridExecution>>(`/grid-plans/${id}/executions`, payload).then((r) => r.data)

export const undoLastGridExecution = (id: number) =>
  http.delete<ApiResponse<GridExecution>>(`/grid-plans/${id}/executions/last`).then((r) => r.data)

// ---------- 个人主页（对外展示，按可见性过滤） ----------

export const listUserGridPlans = (username: string) =>
  http.get<ApiResponse<GridPlanSummary[]>>(`/users/${username}/grid-plans`).then((r) => r.data)

export const getUserGridPlan = (username: string, id: number) =>
  http.get<ApiResponse<GridPlanDetail>>(`/users/${username}/grid-plans/${id}`).then((r) => r.data)

export const listUserGridSnapshots = (username: string, id: number, days = 180) =>
  http
    .get<ApiResponse<GridSnapshot[]>>(`/users/${username}/grid-plans/${id}/snapshots`, { params: { days } })
    .then((r) => r.data)
