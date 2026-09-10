import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export type AuthIPBanStatus = 'active' | 'expired' | 'released'
export type AuthIPBanScope = 'ip' | 'ip_ua'
export type AuthUserAgentCategory = 'browser' | 'automation' | 'empty' | 'other'

export interface AuthIPBan {
  id: number
  ip_address: string
  ban_scope: AuthIPBanScope
  user_agent: string
  ua_category: AuthUserAgentCategory
  source: string
  reason: string
  trigger_path: string
  target_identifier: string
  failure_count: number
  first_seen_at: string
  last_seen_at: string
  banned_at: string
  expires_at: string
  released_at?: string
  released_by_user_id?: number
  released_by_email: string
  release_note: string
  ban_count: number
  created_at: string
  updated_at: string
  status: AuthIPBanStatus
}

export interface AuthIPBanPolicy {
  ua_category: AuthUserAgentCategory
  ban_scope: AuthIPBanScope
  threshold: number
  window_minutes: number
  ban_minutes: number
}

export interface AuthIPBanQuery {
  page?: number
  page_size?: number
  status?: '' | 'all' | AuthIPBanStatus
  q?: string
}

export async function list(params: AuthIPBanQuery): Promise<PaginatedResponse<AuthIPBan>> {
  const { data } = await apiClient.get('/admin/auth-ip-bans', { params })
  return data
}

export async function getPolicy(): Promise<AuthIPBanPolicy[]> {
  const { data } = await apiClient.get('/admin/auth-ip-bans/policy')
  return data
}

export async function release(id: number, note = ''): Promise<AuthIPBan> {
  const { data } = await apiClient.post(`/admin/auth-ip-bans/${id}/release`, { note })
  return data
}

export const authIPBansAPI = {
  list,
  getPolicy,
  release
}

export default authIPBansAPI
