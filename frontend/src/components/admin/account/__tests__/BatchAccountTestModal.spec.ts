import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import BatchAccountTestModal from '../BatchAccountTestModal.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        if (key === 'admin.accounts.batchTest.progressValue') return `${params?.done}/${params?.total}`
        if (key === 'admin.accounts.batchTest.filters.all') return `all-${params?.count}`
        if (key === 'admin.accounts.batchTest.filters.success') return `success-${params?.count}`
        if (key === 'admin.accounts.batchTest.filters.failed') return `failed-${params?.count}`
        if (key === 'admin.accounts.batchTest.filters.unauthorized') return `401-${params?.count}`
        if (key === 'admin.accounts.batchTest.filters.rateLimited') return `429-${params?.count}`
        if (key === 'admin.accounts.batchTest.filters.otherFailed') return `other-${params?.count}`
        return key
      }
    })
  }
})

function createStreamResponse(lines: string[]) {
  const encoder = new TextEncoder()
  const chunks = lines.map((line) => encoder.encode(line))
  let index = 0

  return {
    ok: true,
    body: {
      getReader: () => ({
        read: vi.fn().mockImplementation(async () => {
          if (index < chunks.length) {
            return { done: false, value: chunks[index++] }
          }
          return { done: true, value: undefined }
        }),
        cancel: vi.fn().mockResolvedValue(undefined)
      })
    }
  } as unknown as Response
}

const accounts = [
  { id: 1, name: 'ok-account', platform: 'openai', type: 'oauth', status: 'active' },
  { id: 2, name: 'unauthorized-account', platform: 'openai', type: 'oauth', status: 'active' },
  { id: 3, name: 'limited-account', platform: 'openai', type: 'oauth', status: 'active' }
]

function mountModal() {
  return mount(BatchAccountTestModal, {
    props: {
      show: true,
      accounts
    } as any,
    global: {
      stubs: {
        BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
        Icon: true
      }
    }
  })
}

describe('BatchAccountTestModal', () => {
  beforeEach(() => {
    Object.defineProperty(globalThis, 'localStorage', {
      value: {
        getItem: vi.fn((key: string) => (key === 'auth_token' ? 'test-token' : null))
      },
      configurable: true
    })
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes('/1/test')) {
        return Promise.resolve(createStreamResponse([
          'data: {"type":"test_start","model":"gpt-5.4"}\n',
          'data: {"type":"content","text":"pong"}\n',
          'data: {"type":"test_complete","success":true}\n'
        ]))
      }
      if (url.includes('/2/test')) {
        return Promise.resolve(createStreamResponse([
          'data: {"type":"error","error":"401 unauthorized"}\n'
        ]))
      }
      return Promise.resolve(createStreamResponse([
        'data: {"type":"test_complete","success":false,"error":"upstream returned 429"}\n'
      ]))
    }) as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('runs account tests and classifies success, 401, and 429 results', async () => {
    const wrapper = mountModal()
    await wrapper.get('[data-testid="batch-test-concurrency"]').setValue(2)
    await wrapper.get('[data-testid="batch-test-start"]').trigger('click')
    await flushPromises()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(3)
    const firstRequest = (global.fetch as any).mock.calls[0][1]
    expect(firstRequest.headers.Authorization).toBe('Bearer test-token')
    expect(JSON.parse(firstRequest.body)).toEqual({
      model_id: '',
      prompt: '',
      mode: 'default'
    })

    expect(wrapper.get('[data-testid="batch-test-total-count"]').text()).toBe('3')
    expect(wrapper.get('[data-testid="batch-test-success-count"]').text()).toBe('1')
    expect(wrapper.get('[data-testid="batch-test-failed-count"]').text()).toBe('2')
    expect(wrapper.get('[data-testid="batch-test-progress-count"]').text()).toBe('3/3')

    expect(wrapper.text()).toContain('401-1')
    expect(wrapper.text()).toContain('429-1')
  })

})
