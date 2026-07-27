import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export interface AdminCheckinConfig {
  enabled: boolean
  normal_enabled: boolean
  lucky_enabled: boolean
  normal_min: string
  normal_max: string
  lucky_reward_type: 'multiplier' | 'amount'
  lucky_positive_probability: string
  lucky_min_multiplier: string
  lucky_max_multiplier: string
  lucky_amount_min: string
  lucky_amount_max: string
  risk_control_enabled: boolean
  min_account_age_hours: number
  ip_window_minutes: number
  ip_max_users: number
  config_version: number
  updated_at: string
}

export interface AdminCheckinConfigUpdate {
  enabled: boolean
  normal_enabled: boolean
  lucky_enabled: boolean
  normal_min: string
  normal_max: string
  lucky_reward_type: 'multiplier' | 'amount'
  lucky_positive_probability: string
  lucky_min_multiplier: string
  lucky_max_multiplier: string
  lucky_amount_min: string
  lucky_amount_max: string
  risk_control_enabled: boolean
  min_account_age_hours: number
  ip_window_minutes: number
  ip_max_users: number
  expected_config_version: number
  change_reason: string
}

export interface AdminCheckinOverview {
  business_date: string
  total: number
  normal_count: number
  lucky_count: number
  positive_total: number
  negative_total: number
}

export interface AdminCheckinRecord {
  id: number
  user_id: number
  user_email: string
  checkin_date: string
  mode: 'normal' | 'lucky'
  reward_type: 'multiplier' | 'amount'
  random_value: number
  reward_amount: number
  balance_before: number
  balance_after: number
  checked_in_at: string
}

export interface AdminCheckinRecordQuery {
  page?: number
  page_size?: number
  date?: string
  user_id?: number
  email?: string
}

export async function getConfig(): Promise<AdminCheckinConfig> {
  const { data } = await apiClient.get<AdminCheckinConfig>('/admin/checkin/config')
  return data
}

export async function updateConfig(payload: AdminCheckinConfigUpdate): Promise<AdminCheckinConfig> {
  const { data } = await apiClient.put<AdminCheckinConfig>('/admin/checkin/config', payload)
  return data
}

export async function getOverview(date?: string): Promise<AdminCheckinOverview> {
  const { data } = await apiClient.get<AdminCheckinOverview>('/admin/checkin/overview', { params: { date } })
  return data
}

export async function getRecords(query: AdminCheckinRecordQuery = {}): Promise<PaginatedResponse<AdminCheckinRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminCheckinRecord>>('/admin/checkin/records', { params: query })
  return data
}

const checkinAdminAPI = { getConfig, updateConfig, getOverview, getRecords }
export default checkinAdminAPI
