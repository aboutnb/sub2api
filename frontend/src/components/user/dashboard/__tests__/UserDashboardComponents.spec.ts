import { ref } from 'vue'
import { mount, shallowMount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import UserDashboardCharts from '../UserDashboardCharts.vue'
import UserDashboardQuickActions from '../UserDashboardQuickActions.vue'
import UserDashboardRecentUsage from '../UserDashboardRecentUsage.vue'
import UserDashboardStats from '../UserDashboardStats.vue'

const { push, refreshBatchImageAccess } = vi.hoisted(() => ({
  push: vi.fn(),
  refreshBatchImageAccess: vi.fn().mockResolvedValue(undefined)
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-router')>()
  return {
    ...actual,
    useRouter: () => ({ push })
  }
})

vi.mock('@/composables/useBatchImageAccess', () => ({
  useBatchImageAccess: () => ({
    canUseBatchImage: ref(true),
    refreshBatchImageAccess
  })
}))

const RouterLinkStub = {
  props: ['to'],
  template: '<a :href="to"><slot /></a>'
}

describe('user dashboard themed components', () => {
  beforeEach(() => {
    push.mockReset()
    refreshBatchImageAccess.mockClear()
  })

  it('keeps dashboard statistics and platform details visible', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        balance: 24.5,
        isSimple: false,
        stats: {
          total_api_keys: 3,
          active_api_keys: 2,
          today_requests: 18,
          total_requests: 240,
          today_actual_cost: 0.25,
          today_cost: 0.5,
          total_actual_cost: 3,
          total_cost: 4,
          today_tokens: 1200,
          today_input_tokens: 700,
          today_output_tokens: 300,
          today_cache_creation_tokens: 100,
          today_cache_read_tokens: 100,
          total_tokens: 9000,
          total_input_tokens: 5000,
          total_output_tokens: 2500,
          total_cache_creation_tokens: 750,
          total_cache_read_tokens: 750,
          rpm: 8,
          tpm: 1200,
          average_duration_ms: 420,
          by_platform: [
            {
              platform: 'openai',
              total_actual_cost: 3,
              today_actual_cost: 0.25,
              total_requests: 240,
              total_tokens: 9000
            }
          ]
        } as any
      }
    })

    expect(wrapper.text()).toContain('$24.50')
    expect(wrapper.text()).toContain('OpenAI')
    expect(wrapper.text()).toContain('dashboard.platformBreakdown')
    expect(wrapper.find('.grid-cols-1').exists()).toBe(true)
  })

  it('keeps quick-action routes and batch-image visibility intact', async () => {
    const wrapper = mount(UserDashboardQuickActions)
    const buttons = wrapper.findAll('button')

    expect(buttons).toHaveLength(4)
    expect(refreshBatchImageAccess).toHaveBeenCalledTimes(1)

    await buttons[0].trigger('click')
    await buttons[2].trigger('click')

    expect(push).toHaveBeenNthCalledWith(1, '/keys')
    expect(push).toHaveBeenNthCalledWith(2, '/batch-image')
    expect(wrapper.html()).not.toContain('group-hover:scale')
  })

  it('renders recent usage values and retains the usage destination', () => {
    const wrapper = mount(UserDashboardRecentUsage, {
      props: {
        loading: false,
        data: [
          {
            id: 1,
            model: 'gpt-test',
            created_at: '2026-09-04T08:00:00Z',
            actual_cost: 0.1,
            total_cost: 0.2,
            input_tokens: 12,
            output_tokens: 8
          } as any
        ]
      },
      global: {
        stubs: {
          RouterLink: RouterLinkStub
        }
      }
    })

    expect(wrapper.text()).toContain('gpt-test')
    expect(wrapper.text()).toContain('$0.1000')
    expect(wrapper.get('a').attributes('href')).toBe('/usage')
  })

  it('retains chart filtering inputs and refresh emission', async () => {
    const wrapper = shallowMount(UserDashboardCharts, {
      props: {
        loading: false,
        startDate: '2026-09-01',
        endDate: '2026-09-04',
        granularity: 'day',
        trend: [],
        models: []
      }
    })

    await wrapper.get('button').trigger('click')

    expect(wrapper.emitted('refresh')).toHaveLength(1)
    expect(wrapper.text()).toContain('dashboard.timeRange')
    expect(wrapper.text()).toContain('dashboard.granularity')
  })
})
