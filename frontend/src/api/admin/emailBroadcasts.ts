import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export type EmailBroadcastStatus =
  | 'scheduled'
  | 'pending'
  | 'running'
  | 'succeeded'
  | 'partially_failed'
  | 'canceled'

export interface EmailBroadcastTask {
  id: number
  title: string
  event: string
  status: EmailBroadcastStatus
  variables: Record<string, string>
  audience: EmailBroadcastAudience
  created_by: number
  scheduled_at: string
  total_recipients: number
  sent_count: number
  failed_count: number
  canceled_by?: number
  canceled_at?: string
  started_at?: string
  finished_at?: string
  created_at: string
  updated_at: string
}

export type EmailBroadcastAudienceMode = 'all' | 'role' | 'groups' | 'selected'

export interface EmailBroadcastAudience {
  mode: EmailBroadcastAudienceMode
  roles?: Array<'admin' | 'user'>
  group_ids?: number[]
  emails?: string[]
}

export interface EmailBroadcastRecipient {
  id: number
  task_id: number
  user_id?: number
  email: string
  recipient_name: string
  locale: string
  status: 'pending' | 'sending' | 'sent' | 'failed'
  attempts: number
  last_error?: string
  sent_at?: string
  created_at: string
  updated_at: string
}

export interface EmailBroadcastPayload {
  title?: string
  scheduled_at?: string
  email?: string
  locale?: 'zh' | 'en'
  subject_zh: string
  heading_zh: string
  body_zh: string
  action_zh?: string
  subject_en?: string
  heading_en?: string
  body_en?: string
  action_en?: string
  audience?: EmailBroadcastAudience
}

function broadcastPayloadSignature(payload: EmailBroadcastPayload): string {
  const raw = JSON.stringify(payload)
  let hash = 5381
  for (let index = 0; index < raw.length; index += 1) hash = ((hash << 5) + hash) ^ raw.charCodeAt(index)
  return (hash >>> 0).toString(36)
}

function pendingBroadcastKey(signature: string): string | null {
  try {
    return globalThis.sessionStorage?.getItem(`sub2api:email-broadcast:${signature}`) ?? null
  } catch {
    return null
  }
}

function storeBroadcastKey(signature: string, key: string | null) {
  try {
    const storageKey = `sub2api:email-broadcast:${signature}`
    if (key) globalThis.sessionStorage?.setItem(storageKey, key)
    else globalThis.sessionStorage?.removeItem(storageKey)
  } catch {
    // A fresh key still protects the common double-submit path when storage is unavailable.
  }
}

export async function estimate(audience?: EmailBroadcastAudience): Promise<{ eligible_recipients: number }> {
  const response = audience
    ? await apiClient.post<{ eligible_recipients: number }>('/admin/email-broadcasts/estimate', { audience })
    : await apiClient.get<{ eligible_recipients: number }>('/admin/email-broadcasts/estimate')
  const { data } = response
  return data
}

export async function list(page = 1, pageSize = 20): Promise<PaginatedResponse<EmailBroadcastTask>> {
  const { data } = await apiClient.get<PaginatedResponse<EmailBroadcastTask>>('/admin/email-broadcasts', {
    params: { page, page_size: pageSize }
  })
  return data
}

export async function getById(id: number): Promise<EmailBroadcastTask> {
  const { data } = await apiClient.get<EmailBroadcastTask>(`/admin/email-broadcasts/${id}`)
  return data
}

export async function create(payload: EmailBroadcastPayload): Promise<EmailBroadcastTask> {
  const signature = broadcastPayloadSignature(payload)
  const requestID = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`
  const idempotencyKey = pendingBroadcastKey(signature) ?? `email-broadcast-${signature}-${requestID}`
  storeBroadcastKey(signature, idempotencyKey)
  const { data } = await apiClient.post<EmailBroadcastTask>('/admin/email-broadcasts', payload, {
    headers: { 'Idempotency-Key': idempotencyKey }
  })
  storeBroadcastKey(signature, null)
  return data
}

export async function sendTest(payload: EmailBroadcastPayload): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/admin/email-broadcasts/test', payload)
  return data
}

export async function listRecipients(
  id: number,
  status = '',
  page = 1,
  pageSize = 50
): Promise<PaginatedResponse<EmailBroadcastRecipient>> {
  const { data } = await apiClient.get<PaginatedResponse<EmailBroadcastRecipient>>(`/admin/email-broadcasts/${id}/recipients`, {
    params: { status, page, page_size: pageSize }
  })
  return data
}

export async function cancel(id: number): Promise<{ id: number; status: EmailBroadcastStatus }> {
  const { data } = await apiClient.post<{ id: number; status: EmailBroadcastStatus }>(`/admin/email-broadcasts/${id}/cancel`)
  return data
}

export async function retryFailed(id: number): Promise<{ id: number; retried_recipients: number }> {
  const { data } = await apiClient.post<{ id: number; retried_recipients: number }>(`/admin/email-broadcasts/${id}/retry-failed`)
  return data
}

export default { estimate, list, getById, create, sendTest, listRecipients, cancel, retryFailed }
