import { apiClient } from './client'

export interface HelpDocument {
  id: number
  slug: string
  title: string
  category: string
  summary: string
  content_markdown: string
  selector_schema: Record<string, unknown>
  status: string
  version: number
  published_at?: string | null
  updated_at: string
  can_rollback?: boolean
}

export async function list(category?: string): Promise<HelpDocument[]> {
  const { data } = await apiClient.get<HelpDocument[]>('/help-docs', { params: category ? { category } : {} })
  return data
}
export async function get(slug: string): Promise<HelpDocument> {
  const { data } = await apiClient.get<HelpDocument>(`/help-docs/${slug.split('/').map(encodeURIComponent).join('/')}`)
  return data
}
export default { list, get }
