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
// 积分来源
export type CreditSource = 'gift' | 'subscription' | 'topup'

export interface CreditSourceView {
  source: CreditSource
  remaining: number // quota
  remaining_unit: number // 积分
  nearest_expire: number // 到期时间戳；0=含永久
}

export interface CreditExpiringSoon {
  remaining: number
  remaining_unit: number
  expire_at: number
  source: CreditSource
}

export interface CreditBatchesData {
  sources: CreditSourceView[]
  total: number
  total_unit: number
  currency_unit: string
  expiring_soon?: CreditExpiringSoon
}

export interface CreditBatchesResponse {
  success: boolean
  message: string
  data: CreditBatchesData
}
