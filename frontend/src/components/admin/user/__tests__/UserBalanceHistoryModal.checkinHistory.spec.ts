import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import UserBalanceHistoryModal from '@/components/admin/user/UserBalanceHistoryModal.vue'
import type { AdminUser } from '@/types'

const { getUserBalanceHistory } = vi.hoisted(() => ({
  getUserBalanceHistory: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: { users: { getUserBalanceHistory } },
}))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})
const BaseDialogStub = defineComponent({
  props: ['show'],
  template: '<section v-if="show"><slot /></section>',
})
const IconStub = defineComponent({ props: ['name'], template: '<i :data-icon="name" />' })
const SelectStub = defineComponent({ template: '<button type="button">select</button>' })
const user = {
  id: 1,
  email: 'user@example.com',
  username: 'user',
  balance: 10,
  created_at: '2026-07-01T08:00:00Z',
} as AdminUser

describe('admin user balance history check-in tags', () => {
  beforeEach(() => {
    getUserBalanceHistory.mockReset()
    getUserBalanceHistory.mockResolvedValue({
      items: [
        {
          id: 1,
          code: 'SYS-CHECKIN-1',
          type: 'checkin',
          value: -0.08,
          status: 'used',
          used_by: 1,
          used_at: '2026-07-31T08:00:00Z',
          created_at: '2026-07-31T08:00:00Z',
          group_id: null,
          validity_days: 30,
          notes: '',
          checkin_mode: 'lucky',
        },
        {
          id: 2,
          code: 'SYS-CHECKIN-2',
          type: 'checkin',
          value: 0.05,
          status: 'used',
          used_by: 1,
          used_at: '2026-07-30T08:00:00Z',
          created_at: '2026-07-30T08:00:00Z',
          group_id: null,
          validity_days: 30,
          notes: '',
          checkin_mode: 'normal',
        },
      ],
      total: 2,
      page: 1,
      page_size: 15,
      pages: 1,
      total_recharged: 0,
    })
  })

  it('labels only lucky check-ins', async () => {
    const wrapper = mount(UserBalanceHistoryModal, {
      props: { show: false, user },
      global: {
        stubs: { BaseDialog: BaseDialogStub, Icon: IconStub, Select: SelectStub },
      },
    })
    await wrapper.setProps({ show: true })
    await flushPromises()

    const tags = wrapper.findAll('[data-testid="lucky-checkin-tag"]')
    expect(tags).toHaveLength(1)
    expect(tags[0].text()).toBe('checkin.lucky')
    expect(wrapper.text()).toContain('+$0.05')
  })
})
