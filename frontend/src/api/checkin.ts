import { apiClient } from './client'
import type { BasePaginationResponse, CheckinRecord, CheckinStatus } from '@/types'

const pendingCheckinKeys = new Map<string, string>()

function currentUserID(): string | null {
  try {
    const raw = globalThis.localStorage?.getItem('auth_user')
    if (!raw) return null
    const id = JSON.parse(raw)?.id
    return typeof id === 'number' && Number.isSafeInteger(id) && id > 0 ? String(id) : null
  } catch {
    return null
  }
}

function operationScope(mode: 'normal' | 'lucky', businessDate: string): string | null {
  const userID = currentUserID()
  if (!userID || !/^\d{4}-\d{2}-\d{2}$/.test(businessDate)) return null
  return `${userID}:${businessDate}:${mode}`
}

function storageKey(scope: string): string {
  return `sub2api:user:checkin:${scope}`
}

function getPendingKey(scope: string | null): string | null {
  if (!scope) return null
  const memoryKey = pendingCheckinKeys.get(scope)
  if (memoryKey) return memoryKey
  try {
    return globalThis.sessionStorage?.getItem(storageKey(scope)) ?? null
  } catch {
    return null
  }
}

function storePendingKey(scope: string | null, key: string | null): void {
  if (!scope) return
  if (key) pendingCheckinKeys.set(scope, key)
  else pendingCheckinKeys.delete(scope)
  try {
    if (key) globalThis.sessionStorage?.setItem(storageKey(scope), key)
    else globalThis.sessionStorage?.removeItem(storageKey(scope))
  } catch {
    // In-memory retry protection still works when browser storage is unavailable.
  }
}

export async function getStatus(): Promise<CheckinStatus> {
  const { data } = await apiClient.get<CheckinStatus>('/user/checkin/status')
  return data
}

export async function checkIn(mode: 'normal' | 'lucky', businessDate: string, turnstileToken?: string): Promise<{
  newly_checked_in: boolean
  record: CheckinRecord
}> {
  const scope = operationScope(mode, businessDate)
  const requestID = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`
  const idempotencyKey = getPendingKey(scope) ?? `checkin-${businessDate}-${requestID}`
  storePendingKey(scope, idempotencyKey)
  const headers: Record<string, string> = { 'Idempotency-Key': idempotencyKey }
  if (turnstileToken) headers['X-Turnstile-Token'] = turnstileToken
  const { data } = await apiClient.post('/user/checkin', { mode }, { headers })
  storePendingKey(scope, null)
  return data
}

export async function getRecords(page = 1, pageSize = 20): Promise<BasePaginationResponse<CheckinRecord>> {
  const { data } = await apiClient.get<BasePaginationResponse<CheckinRecord>>('/user/checkin/records', {
    params: { page, page_size: pageSize }
  })
  return data
}

export const checkinAPI = { getStatus, checkIn, getRecords }

export default checkinAPI
