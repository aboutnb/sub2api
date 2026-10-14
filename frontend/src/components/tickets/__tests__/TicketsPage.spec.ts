import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useAppStore } from '@/stores/app'
import TicketsPage from '../TicketsPage.vue'
import { createTicketI18n } from './i18n'
import type { SupportTicket } from '@/types/support-ticket'

const mocks = vi.hoisted(() => ({
  list: vi.fn(), get: vi.fn(), create: vi.fn(), reply: vi.fn(), close: vi.fn(),
  adminList: vi.fn(), adminGet: vi.fn(), adminReply: vi.fn(), updateStatus: vi.fn()
}))
vi.mock('@/api/tickets', () => ({ ticketsAPI: { list: mocks.list, getById: mocks.get, create: mocks.create, reply: mocks.reply, close: mocks.close } }))
vi.mock('@/api/admin', () => ({ adminAPI: { tickets: { list: mocks.adminList, getById: mocks.adminGet, reply: mocks.adminReply, updateStatus: mocks.updateStatus } } }))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<div><slot /></div>' } }))

const ticket: SupportTicket = {
  id: 1, user_id: 1, user_email: 'owner@example.test', subject: '充值未到账', category: 'billing', status: 'open', preview: '请核查',
  created_at: '2026-10-14T00:00:00Z', updated_at: '2026-10-14T00:00:00Z', last_message_at: '2026-10-14T00:00:00Z', closed_at: null,
  messages: [{ id: 1, ticket_id: 1, author_id: 1, author_role: 'user', body: '<script>plain text</script>\n请核查', created_at: '2026-10-14T00:00:00Z' }]
}
let wrapper: VueWrapper | undefined
beforeEach(() => {
  vi.resetAllMocks()
  localStorage.clear()
  const list = { items: [{ ...ticket, messages: undefined }], page: 1, page_size: 20, total: 1, pages: 1 }
  mocks.list.mockResolvedValue(list)
  mocks.adminList.mockResolvedValue(list)
  mocks.get.mockResolvedValue(ticket)
  mocks.adminGet.mockResolvedValue(ticket)
})
afterEach(() => { wrapper?.unmount(); document.body.innerHTML = '' })

async function render(admin = false) {
  const pinia = createPinia()
  setActivePinia(pinia)
  const store = useAppStore()
  const success = vi.spyOn(store, 'showSuccess')
  const error = vi.spyOn(store, 'showError')
  wrapper = mount(TicketsPage, { props: { admin }, attachTo: document.body, global: {
    plugins: [pinia, createTicketI18n()], stubs: { Icon: true, teleport: true }
  } })
  await flushPromises()
  return { view: wrapper, success, error }
}
function button(view: VueWrapper, text: string) {
  const result = view.findAll('button').find(item => item.text() === text)
  expect(result, `button ${text}`).toBeDefined()
  return result!
}

describe('TicketsPage', () => {
  it('creates, replies to, and closes a user ticket with text-only message rendering', async () => {
    mocks.create.mockResolvedValue(ticket)
    mocks.reply.mockResolvedValue({ ...ticket, status: 'pending_admin' })
    mocks.close.mockResolvedValue({ ...ticket, status: 'closed', closed_at: '2026-10-14T01:00:00Z' })
    const { view, success } = await render()
    expect(mocks.list).toHaveBeenCalledWith(1, 20, { category: '', status: '', search: '' }, expect.any(Object))
    expect(view.find('[role="dialog"]').exists()).toBe(false)
    await button(view, '新建工单').trigger('click')
    await view.get('#ticket-subject').setValue('充值未到账')
    await view.get('#ticket-body').setValue('请核查')
    await view.get('#ticket-create-form').trigger('submit')
    await flushPromises()
    expect(mocks.create).toHaveBeenCalledWith({ subject: '充值未到账', body: '请核查', category: 'billing' })
    expect(success).toHaveBeenCalledWith('工单已提交')
    expect(view.find('article script').exists()).toBe(false)
    expect(view.get('article').text()).toContain('<script>plain text</script>')
    await view.get('#ticket-reply').setValue('补充信息')
    await view.get('#ticket-reply-form').trigger('submit')
    await flushPromises()
    expect(mocks.reply).toHaveBeenCalledWith(1, '补充信息')
    expect((view.get('#ticket-reply').element as HTMLTextAreaElement).value).toBe('')
    await button(view, '关闭工单').trigger('click')
    await flushPromises()
    expect(mocks.close).toHaveBeenCalledWith(1)
    expect(view.find('#ticket-reply').exists()).toBe(false)
    expect(view.text()).toContain('工单已关闭。如需继续沟通，请创建新工单。')
  })

  it('opens a row and preserves drafts when a localized API error occurs', async () => {
    mocks.reply.mockRejectedValue({ reason: 'TICKET_BODY_INVALID' })
    const { view, error } = await render()
    await view.get('[data-ticket-id="1"]').trigger('click')
    await flushPromises()
    expect(mocks.get).toHaveBeenCalledWith(1, expect.any(Object))
    await view.get('#ticket-reply').setValue('保留回复')
    await view.get('#ticket-reply-form').trigger('submit')
    await flushPromises()
    expect(error).toHaveBeenCalledWith('正文不能超过 4000 个字符')
    expect((view.get('#ticket-reply').element as HTMLTextAreaElement).value).toBe('保留回复')
  })

  it('uses admin queues, saves status, and permits replying to a closed ticket', async () => {
    mocks.adminGet.mockResolvedValue({ ...ticket, status: 'closed', closed_at: '2026-10-14T01:00:00Z' })
    mocks.updateStatus.mockResolvedValue({ ...ticket, status: 'open' })
    mocks.adminReply.mockResolvedValue({ ...ticket, status: 'pending_user', closed_at: null })
    const { view } = await render(true)
    expect(mocks.adminList).toHaveBeenCalledWith(1, 20, { category: '', queue: 'unanswered', search: '' }, expect.any(Object))
    expect(view.text()).not.toContain('新建工单')
    await view.get('[data-ticket-id="1"]').trigger('click')
    await flushPromises()
    expect(view.text()).toContain('发送回复将重新打开此工单')
    const statusSelect = view.get('.ticket-status-editor [role="combobox"]')
    await statusSelect.trigger('click')
    document.querySelectorAll<HTMLElement>('[role="option"]')[0].click()
    await flushPromises()
    await button(view, '保存').trigger('click')
    await flushPromises()
    expect(mocks.updateStatus).toHaveBeenCalledWith(1, 'open')
    await view.get('#ticket-reply').setValue('客服回复')
    await view.get('#ticket-reply-form').trigger('submit')
    await flushPromises()
    expect(mocks.adminReply).toHaveBeenCalledWith(1, '客服回复')
    expect(view.get('.conversation-header').text()).toContain('待用户回复')
  })

  it('shows the limit from error metadata while retaining the creation form', async () => {
    mocks.create.mockRejectedValue({ reason: 'TICKET_OPEN_LIMIT', metadata: { max: '10' } })
    const { view, error } = await render()
    await button(view, '新建工单').trigger('click')
    await view.get('#ticket-subject').setValue('问题')
    await view.get('#ticket-body').setValue('描述')
    await view.get('#ticket-create-form').trigger('submit')
    await flushPromises()
    expect(error).toHaveBeenCalledWith('最多同时保留 10 个未关闭工单，请先关闭已解决的工单')
    expect((view.get('#ticket-subject').element as HTMLInputElement).value).toBe('问题')
  })

  it('hides pagination when no tickets match the filters', async () => {
    mocks.list.mockResolvedValue({ items: [], page: 1, page_size: 20, total: 0, pages: 1 })
    const { view } = await render()
    expect(view.text()).toContain('暂无工单')
    expect(view.find('nav[aria-label="Pagination"]').exists()).toBe(false)
  })

  it('filters unclosed and closed tickets, searches by ID and clears filters while keeping the conversation', async () => {
    const { view } = await render()
    await button(view, '未关闭').trigger('click')
    await flushPromises()
    expect(mocks.list).toHaveBeenLastCalledWith(1, 20, { category: '', status: 'unclosed', search: '' }, expect.any(Object))
    await view.get('input[aria-label="搜索"]').setValue('#1')
    await view.get('form[role="search"]').trigger('submit')
    await flushPromises()
    expect(mocks.list).toHaveBeenLastCalledWith(1, 20, { category: '', status: 'unclosed', search: '#1' }, expect.any(Object))
    mocks.list.mockResolvedValue({ items: [], page: 1, page_size: 20, total: 0, pages: 0 })
    await button(view, '已关闭').trigger('click')
    await flushPromises()
    expect(view.get('.conversation-header').text()).toContain(ticket.subject)
    await button(view, '清除筛选').trigger('click')
    await flushPromises()
    expect(mocks.list).toHaveBeenLastCalledWith(1, 20, { category: '', status: '', search: '' }, expect.any(Object))
  })

  it('preserves separate reply drafts when switching tickets and returning to the mobile list', async () => {
    const second = { ...ticket, id: 2, subject: '模型问题' }
    mocks.list.mockResolvedValue({ items: [ticket, second], page: 1, page_size: 20, total: 2, pages: 1 })
    mocks.get.mockImplementation((id: number) => Promise.resolve(id === 1 ? ticket : second))
    const { view } = await render()
    expect(view.get('.tickets-workspace').classes()).not.toContain('show-conversation')
    await view.get('[data-ticket-id="1"]').trigger('click')
    await flushPromises()
    expect(view.get('.tickets-workspace').classes()).toContain('show-conversation')
    await view.get('#ticket-reply').setValue('第一条的草稿')
    await view.get('[data-ticket-id="2"]').trigger('click')
    await flushPromises()
    expect((view.get('#ticket-reply').element as HTMLTextAreaElement).value).toBe('')
    await view.get('#ticket-reply').setValue('第二条的草稿')
    await button(view, '返回工单列表').trigger('click')
    expect(view.get('.tickets-workspace').classes()).not.toContain('show-conversation')
    await view.get('[data-ticket-id="1"]').trigger('click')
    await flushPromises()
    expect((view.get('#ticket-reply').element as HTMLTextAreaElement).value).toBe('第一条的草稿')
  })

  it('ignores stale detail responses when the user changes tickets', async () => {
    const second = { ...ticket, id: 2, subject: '第二条工单' }
    mocks.list.mockResolvedValue({ items: [ticket, second], page: 1, page_size: 20, total: 2, pages: 1 })
    let resolveFirst!: (value: SupportTicket) => void
    mocks.get.mockImplementation((id: number) => id === 1 ? new Promise<SupportTicket>(resolve => { resolveFirst = resolve }) : Promise.resolve(second))
    const { view } = await render()
    await view.get('[data-ticket-id="2"]').trigger('click')
    await flushPromises()
    resolveFirst(ticket)
    await flushPromises()
    expect(view.get('.conversation-header').text()).toContain(second.subject)
    expect(view.get('.conversation-header').text()).not.toContain(ticket.subject)
  })

  it('prevents changing the conversation during a reply submission', async () => {
    let resolveReply!: (value: SupportTicket) => void
    mocks.reply.mockImplementation(() => new Promise<SupportTicket>(resolve => { resolveReply = resolve }))
    const { view } = await render()
    await view.get('#ticket-reply').setValue('提交中的回复')
    await view.get('#ticket-reply-form').trigger('submit')
    expect(view.get('[data-ticket-id="1"]').attributes('disabled')).toBeDefined()
    expect(button(view, '新建工单').attributes('disabled')).toBeDefined()
    await view.get('#ticket-reply-form').trigger('submit')
    expect(mocks.reply).toHaveBeenCalledTimes(1)
    resolveReply({ ...ticket, status: 'pending_admin' })
    await flushPromises()
    expect(view.get('[data-ticket-id="1"]').attributes('disabled')).toBeUndefined()
  })
})
