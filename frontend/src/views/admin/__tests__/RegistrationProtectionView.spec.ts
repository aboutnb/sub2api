import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RegistrationProtectionView from '../RegistrationProtectionView.vue'

const api = vi.hoisted(() => ({ settings: vi.fn(), updateSettings: vi.fn(), list: vi.fn(), review: vi.fn(), releaseBlock: vi.fn() }))
vi.mock('@/api/admin/registrationProtection', () => ({ registrationProtectionAPI: api }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: vi.fn() }) }))
vi.mock('vue-i18n', async () => ({ ...(await vi.importActual<typeof import('vue-i18n')>('vue-i18n')), useI18n: () => ({ t: (key: string) => key, te: () => false }) }))
const layout = defineComponent({ template: '<main><slot /></main>' })
const dialog = defineComponent({ props: ['show', 'title', 'pending'], template: '<section v-if="show" role="dialog"><slot /><slot name="footer" /></section>' })
const defaults = { enabled: true, ip_success_limit: 10, identity_success_limit: 2, success_window_hours: 24, ip_failure_limit: 100, identity_failure_limit: 20, failure_window_minutes: 10, block_minutes: 15, observe_identity_success_limit: 4 }
const record = { id: 7, user_id: 42, email: 'user@example.com', ip_address: '203.0.113.1', user_agent: 'browser', status: 'restricted', concurrency: -1, previous_concurrency: 3, created_at: '2026-09-12T00:00:00Z' }
const create = () => mount(RegistrationProtectionView, { global: { stubs: { AppLayout: layout, BaseDialog: dialog } } })
function button(wrapper: ReturnType<typeof create>, suffix: string) {
  const found = wrapper.findAll('button').find(b => b.text() === `admin.registrationProtection.${suffix}`)
  if (!found) throw new Error(`Missing button ${suffix}`)
  return found
}
beforeEach(() => {
  vi.resetAllMocks()
  api.settings.mockResolvedValue({ ...defaults })
  api.list.mockResolvedValue({ items: [record], total: 1, page: 1, page_size: 20 })
  api.review.mockResolvedValue(undefined)
})
describe('RegistrationProtectionView', () => {
  it('releases a source restriction without changing an account', async () => {
    api.list.mockResolvedValue({ items: [{ ...record, status: 'active' }], total: 1, page: 1, page_size: 20 })
    const wrapper = create(); await flushPromises()
    await button(wrapper, 'tabs.sources').trigger('click'); await flushPromises()
    await wrapper.find('select').setValue('blocks'); await flushPromises()
    await button(wrapper, 'release').trigger('click')
    await wrapper.get('textarea').setValue('shared network verified')
    await button(wrapper, 'confirm').trigger('click'); await flushPromises()
    expect(api.releaseBlock).toHaveBeenCalledWith(7, 'shared network verified')
    expect(api.review).not.toHaveBeenCalled()
    wrapper.unmount()
  })
  it('saves the edited quotas and preserves disabled protection', async () => {
    const wrapper = create(); await flushPromises()
    await wrapper.find('input[type="checkbox"]').setValue(false)
    await wrapper.findAll('input[type="number"]')[2].setValue('48')
    api.updateSettings.mockResolvedValue({ ...defaults, enabled: false, success_window_hours: 48 })
    await wrapper.find('form').trigger('submit'); await flushPromises()
    expect(api.updateSettings).toHaveBeenCalledWith({ ...defaults, enabled: false, success_window_hours: 48 })
    wrapper.unmount()
  })
  it('requires a review note and retains it after a conflict', async () => {
    const wrapper = create(); await flushPromises()
    await button(wrapper, 'tabs.accounts').trigger('click'); await flushPromises()
    await button(wrapper, 'release').trigger('click')
    expect(button(wrapper, 'confirm').attributes('disabled')).toBeDefined()
    await wrapper.find('textarea').setValue('verified owner')
    api.review.mockRejectedValueOnce(new Error('state changed'))
    await button(wrapper, 'confirm').trigger('click'); await flushPromises()
    expect(api.review).toHaveBeenCalledWith(7, 'release', 'verified owner')
    expect(wrapper.get('[role="dialog"]').text()).toContain('state changed')
    expect((wrapper.get('textarea').element as HTMLTextAreaElement).value).toBe('verified owner')
    wrapper.unmount()
  })
  it('ignores stale source responses after switching to risk accounts', async () => {
    let resolveSource!: (value: unknown) => void
    api.list.mockImplementationOnce(() => new Promise(resolve => { resolveSource = resolve }))
    const wrapper = create(); await flushPromises()
    await button(wrapper, 'tabs.sources').trigger('click'); await flushPromises()
    await button(wrapper, 'tabs.accounts').trigger('click'); await flushPromises()
    resolveSource({ items: [{ ...record, email: 'stale@example.com' }], total: 1 })
    await flushPromises()
    expect(wrapper.text()).toContain('user@example.com')
    expect(wrapper.text()).not.toContain('stale@example.com')
    wrapper.unmount()
  })
  it('shows a load failure without offering to overwrite settings with defaults', async () => {
    api.settings.mockRejectedValueOnce(new Error('unavailable'))
    const wrapper = create(); await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('unavailable')
    expect(wrapper.find('input[type="number"]').exists()).toBe(false)
    expect(button(wrapper, 'retry').exists()).toBe(true)
    wrapper.unmount()
  })
})
