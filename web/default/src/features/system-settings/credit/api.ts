/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { api } from '@/lib/api'

// ============================================================================
// 充值档位 (Recharge Tier) — /api/admin/recharge-tier/*
// ============================================================================

export interface RechargeTier {
  id: string
  name: string
  amount: number
  base_credits: number
  gift_credits: number
  validity_days: number
  status: number // 1=上架 0=下架
  sort_order: number
}

export async function listRechargeTiers(): Promise<RechargeTier[]> {
  const res = await api.get('/api/admin/recharge-tier/')
  return res.data?.data ?? []
}

export async function saveRechargeTier(tier: RechargeTier): Promise<RechargeTier> {
  const res = await api.post('/api/admin/recharge-tier/', tier)
  return res.data?.data
}

export async function deleteRechargeTier(id: string): Promise<void> {
  await api.delete(`/api/admin/recharge-tier/${id}`)
}

// ============================================================================
// 模型价格定时生效 (Model Price Schedule) — /api/admin/model-price-schedule/*
// ============================================================================

export interface ModelPriceSchedule {
  id: number
  model_name: string
  model_ratio: number
  completion_ratio: number
  effective_at: number
  applied: boolean
  applied_at: number
  operator_id: number
  created_at: number
}

export interface CreateModelPriceScheduleRequest {
  model_name: string
  model_ratio: number
  completion_ratio: number
  effective_at: number
}

export async function listModelPriceSchedules(
  applied?: boolean
): Promise<ModelPriceSchedule[]> {
  const params = applied === undefined ? '' : `?applied=${applied}`
  const res = await api.get(`/api/admin/model-price-schedule/${params}`)
  return res.data?.data ?? []
}

export async function createModelPriceSchedule(
  request: CreateModelPriceScheduleRequest
): Promise<ModelPriceSchedule> {
  const res = await api.post('/api/admin/model-price-schedule/', request)
  return res.data?.data
}

export async function deleteModelPriceSchedule(id: number): Promise<void> {
  await api.delete(`/api/admin/model-price-schedule/${id}`)
}

// ============================================================================
// 流量池总览 (Credit Overview) — /api/admin/credit/*
// ============================================================================

export interface CreditOverviewData {
  total_tokens: number
  total_quota: number
  total_credits: number
  cost_estimate: number
  total_recharge: number
  user_count: number
  active_user_count: number
  model_tokens: Record<string, number>
  model_credits: Record<string, number>
  currency_unit: string
}

export interface UserCreditStats {
  total: number
  normal: number
  low_balance: number
  exhausted: number
}

export async function getCreditOverview(
  startTimestamp?: number,
  endTimestamp?: number
): Promise<CreditOverviewData> {
  const params = new URLSearchParams()
  if (startTimestamp) params.append('start_timestamp', String(startTimestamp))
  if (endTimestamp) params.append('end_timestamp', String(endTimestamp))
  const qs = params.toString()
  const res = await api.get(`/api/admin/credit/overview${qs ? `?${qs}` : ''}`)
  return res.data?.data
}

export async function getUserCreditStats(): Promise<UserCreditStats> {
  const res = await api.get('/api/admin/credit/user-stats')
  return res.data?.data
}

/** 返回 CSV 导出的下载 URL（直接交给浏览器下载）。 */
export function buildCreditExportUrl(
  type: 'model' | 'user',
  startTimestamp?: number,
  endTimestamp?: number
): string {
  const params = new URLSearchParams({ type })
  if (startTimestamp) params.append('start_timestamp', String(startTimestamp))
  if (endTimestamp) params.append('end_timestamp', String(endTimestamp))
  return `/api/admin/credit/export?${params.toString()}`
}
