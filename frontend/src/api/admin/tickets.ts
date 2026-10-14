import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'
import type { SupportTicket, TicketStatus } from '@/types/support-ticket'

export async function list(
  page = 1,
  pageSize = 20,
  filters?: { category?: string; queue?: string; search?: string },
  options?: { signal?: AbortSignal }
): Promise<BasePaginationResponse<SupportTicket>> {
  const { data } = await apiClient.get<BasePaginationResponse<SupportTicket>>('/admin/tickets', {
    params: { page, page_size: pageSize, ...filters }, signal: options?.signal
  })
  return data
}

export async function getById(id: number, options?: { signal?: AbortSignal }): Promise<SupportTicket> {
  const { data } = await apiClient.get<SupportTicket>(`/admin/tickets/${id}`, options)
  return data
}

export async function reply(id: number, body: string): Promise<SupportTicket> {
  const { data } = await apiClient.post<SupportTicket>(`/admin/tickets/${id}/messages`, { body })
  return data
}

export async function updateStatus(id: number, status: TicketStatus): Promise<SupportTicket> {
  const { data } = await apiClient.put<SupportTicket>(`/admin/tickets/${id}/status`, { status })
  return data
}

export const adminTicketsAPI = { list, getById, reply, updateStatus }
export default adminTicketsAPI
