import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useTicketNotificationsStore } from '../ticketNotifications'
const summary = vi.hoisted(() => vi.fn())
vi.mock('@/api/admin/tickets', () => ({ adminTicketsAPI: { summary } }))
beforeEach(() => { setActivePinia(createPinia()); summary.mockReset() })
describe('ticket notifications', () => {
 it('counts new tickets even when another ticket closes and pending count stays unchanged', async () => {
  const store = useTicketNotificationsStore(); store.setOwner(1)
  summary.mockResolvedValueOnce({ pending_count: 2, total_count: 3 })
  expect(await store.refresh()).toEqual({ initial: true, added: 0 })
  summary.mockResolvedValue({ pending_count: 2, total_count: 4 })
  expect(await store.refresh()).toEqual({ initial: false, added: 1 })
  expect(await store.refresh()).toEqual({ initial: false, added: 0 })
  store.setOwner(1)
  expect(await store.refresh()).toEqual({ initial: false, added: 0 })
 })
 it('keeps counts on failure and rejects stale responses after switching users', async () => {
  const store = useTicketNotificationsStore(); store.setOwner(1)
  summary.mockResolvedValueOnce({ pending_count: 3, total_count: 4 }); await store.refresh()
  summary.mockRejectedValueOnce(new Error('offline')); expect(await store.refresh()).toBeNull()
  expect(store.pendingCount).toBe(3)
  let resolve!: (value: unknown) => void
  summary.mockImplementationOnce(() => new Promise(r => { resolve = r }))
  const pending = store.refresh(); store.setOwner(2)
  resolve({ pending_count: 99, total_count: 99 }); await pending
  expect(store.pendingCount).toBe(0)
  store.setOwner(null); expect(await store.refresh()).toBeNull()
 })
})
