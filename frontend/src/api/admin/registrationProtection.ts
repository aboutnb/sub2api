import { apiClient } from '../client'

export interface RegistrationProtectionSettings {
  enabled: boolean
  ip_success_limit: number
  identity_success_limit: number
  success_window_hours: number
  ip_failure_limit: number
  identity_failure_limit: number
  failure_window_minutes: number
  block_minutes: number
  observe_identity_success_limit: number
}

export type RegistrationRecordKind = 'sources' | 'events' | 'blocks' | 'accounts'
export interface RegistrationRiskRecord {
  id: number
  user_id?: number
  email?: string
  ip_address: string
  user_agent: string
  status: string
  reason?: string
  action?: string
  concurrency?: number
  previous_concurrency?: number
  created_at: string
  expires_at?: string
  trigger_path?: string
  observed_count?: number
  failure_count?: number
  review_note?: string
  release_note?: string
  reviewed_at?: string
  reviewed_by?: number
}
export interface RegistrationRiskPage {
  items: RegistrationRiskRecord[]
  total: number
  page: number
  page_size: number
}

const base = '/admin/registration-protection'
export const registrationProtectionAPI = {
  async settings(): Promise<RegistrationProtectionSettings> {
    return (await apiClient.get(`${base}/settings`)).data
  },
  async updateSettings(settings: RegistrationProtectionSettings): Promise<RegistrationProtectionSettings> {
    return (await apiClient.put(`${base}/settings`, settings)).data
  },
  async list(kind: RegistrationRecordKind, params: { page: number; page_size: number; q: string; status: string }): Promise<RegistrationRiskPage> {
    return (await apiClient.get(`${base}/records/${kind}`, { params })).data
  },
  async review(id: number, action: 'restrict' | 'release', note: string): Promise<void> {
    await apiClient.post(`${base}/accounts/${id}/review`, { action, note })
  },
  async releaseBlock(id: number, note: string): Promise<void> {
    await apiClient.post(`${base}/blocks/${id}/release`, { note })
  }
}
