export type TicketCategory = 'billing' | 'api_models' | 'account' | 'other'
export type TicketStatus = 'open' | 'pending_admin' | 'pending_user' | 'closed'
export type TicketQueue = 'unanswered' | 'awaiting_user' | 'closed'

export interface SupportTicketMessage {
  id: number
  ticket_id: number
  author_id: number
  author_role: 'user' | 'admin'
  body: string
  created_at: string
}

export interface SupportTicket {
  id: number
  user_id: number
  user_email?: string
  subject: string
  category: TicketCategory
  status: TicketStatus
  preview: string
  last_message_at: string
  closed_at: string | null
  created_at: string
  updated_at: string
  messages?: SupportTicketMessage[]
}

export interface CreateTicketRequest {
  subject: string
  body: string
  category: TicketCategory
}
