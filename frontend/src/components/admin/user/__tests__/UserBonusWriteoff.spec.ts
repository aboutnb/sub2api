import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import UserBonusWriteoff from '../UserBonusWriteoff.vue'

const { get, put, showSuccess } = vi.hoisted(() => ({ get: vi.fn(), put: vi.fn(), showSuccess: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: { get, put } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

describe('UserBonusWriteoff', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    get.mockResolvedValue({ data: { policy: null, items: [], total: 0 } })
    put.mockResolvedValue({ data: { updated: true } })
  })

  it('defaults off and never saves while loading or toggling', async () => {
    const wrapper = mount(UserBonusWriteoff, { props: { userId: 7 } })
    await flushPromises()
    expect((wrapper.get('[role="switch"]').element as HTMLInputElement).checked).toBe(false)
    await wrapper.get('[role="switch"]').setValue(true)
    expect(put).not.toHaveBeenCalled()
    await wrapper.get('fieldset button').trigger('click')
    expect(put).not.toHaveBeenCalled()
    expect(wrapper.get('[role="alert"]').text()).toContain('invalid')
  })

  it('saves only the specified account with explicit consent', async () => {
    const wrapper = mount(UserBonusWriteoff, { props: { userId: 7 } })
    await flushPromises()
    await wrapper.get('[role="switch"]').setValue(true)
    await wrapper.get('textarea').setValue('Agreed in ticket 42')
    await wrapper.get('input[type="number"]').setValue(20)
    await wrapper.get('fieldset button').trigger('click')
    await flushPromises()
    expect(put).toHaveBeenCalledWith('/admin/users/7/bonus-writeoff', expect.objectContaining({ enabled: true, amount_limit: 20, consent_note: 'Agreed in ticket 42' }))
    expect(showSuccess).toHaveBeenCalledOnce()
    expect(wrapper.get('fieldset button').attributes('type')).toBe('button')
  })

  it('does not allow saving after a load failure', async () => {
    get.mockRejectedValueOnce(new Error('offline'))
    const wrapper = mount(UserBonusWriteoff, { props: { userId: 7 } })
    await flushPromises()
    expect(wrapper.get('fieldset button').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[role="alert"]').text()).toContain('loadError')
  })

  it('ignores responses from a previous user', async () => {
    let resolveOld!: (value: unknown) => void
    get.mockImplementationOnce(() => new Promise(resolve => { resolveOld = resolve }))
    const wrapper = mount(UserBonusWriteoff, { props: { userId: 7 } })
    await wrapper.setProps({ userId: 8 })
    await flushPromises()
    resolveOld({ data: { policy: { enabled: true, consent_note: 'old account' }, items: [], total: 0 } })
    await flushPromises()
    expect((wrapper.get('[role="switch"]').element as HTMLInputElement).checked).toBe(false)
    expect((wrapper.get('textarea').element as HTMLTextAreaElement).value).toBe('')
  })
})
