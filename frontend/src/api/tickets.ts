import { apiClient } from './client'
import type { BasePaginationResponse } from '@/types'
import type { CreateTicketRequest, SupportTicket } from '@/types/support-ticket'

export async function list(
  page = 1,
  pageSize = 20,
  filters?: { category?: string; status?: string; search?: string },
  options?: { signal?: AbortSignal }
): Promise<BasePaginationResponse<SupportTicket>> {
  const { data } = await apiClient.get<BasePaginationResponse<SupportTicket>>('/tickets', {
    params: { page, page_size: pageSize, ...filters }, signal: options?.signal
  })
  return data
}

export async function getById(id: number, options?: { signal?: AbortSignal }): Promise<SupportTicket> {
  const { data } = await apiClient.get<SupportTicket>(`/tickets/${id}`, options)
  return data
}

export async function create(request: CreateTicketRequest): Promise<SupportTicket> {
  const { data } = await apiClient.post<SupportTicket>('/tickets', request)
  return data
}

export async function reply(id: number, body: string): Promise<SupportTicket> {
  const { data } = await apiClient.post<SupportTicket>(`/tickets/${id}/messages`, { body })
  return data
}

export async function close(id: number): Promise<SupportTicket> {
  const { data } = await apiClient.post<SupportTicket>(`/tickets/${id}/close`)
  return data
}

export const ticketsAPI = { list, getById, create, reply, close }
export default ticketsAPI
