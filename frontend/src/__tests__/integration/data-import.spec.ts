import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImportDataModal from '@/components/admin/account/ImportDataModal.vue'
import { adminAPI } from '@/api/admin'

const { showError, showSuccess, importData, getAllGroups } = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn(),
  importData: vi.fn(),
  getAllGroups: vi.fn()
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      importData
    },
    groups: {
      getAll: getAllGroups
    }
  }
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

describe('ImportDataModal', () => {
  beforeEach(() => {
    showError.mockReset()
    showSuccess.mockReset()
    importData.mockReset()
    getAllGroups.mockReset()
    getAllGroups.mockResolvedValue([])
  })

  it('未选择文件时提示错误', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    await wrapper.find('form').trigger('submit')
    expect(showError).toHaveBeenCalledWith('admin.accounts.dataImportSelectFile')
  })

  it('无效 JSON 时提示解析失败', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const input = wrapper.find('input[type="file"]')
    const file = new File(['invalid json'], 'data.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: () => Promise.resolve('invalid json')
    })
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await Promise.resolve()

    expect(showError).toHaveBeenCalledWith('admin.accounts.dataImportParseFailed')
  })

  it('导入时会提交所选分组', async () => {
    getAllGroups.mockResolvedValue([
      {
        id: 12,
        name: 'OpenAI',
        platform: 'openai',
        subscription_type: 'standard',
        rate_multiplier: 1,
        account_count: 0
      },
      {
        id: 15,
        name: 'Claude',
        platform: 'anthropic',
        subscription_type: 'standard',
        rate_multiplier: 1,
        account_count: 0
      }
    ])
    importData.mockResolvedValue({
      account_created: 1,
      account_failed: 0,
      proxy_created: 0,
      proxy_reused: 0,
      proxy_failed: 0
    })

    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
          GroupBadge: { template: '<span><slot />{{ name }}</span>', props: ['name'] },
          Icon: true
        }
      }
    })
    await flushPromises()

    const groupCheckboxes = wrapper.findAll('input[type="checkbox"]')
    await groupCheckboxes[0].setValue(true)
    await groupCheckboxes[1].setValue(true)

    const input = wrapper.find('input[type="file"]')
    const payload = JSON.stringify({ type: 'sub2api-data', version: 1, proxies: [], accounts: [] })
    const file = new File([payload], 'data.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: () => Promise.resolve(payload)
    })
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await Promise.resolve()

    expect(adminAPI.accounts.importData).toHaveBeenCalledWith(
      expect.objectContaining({
        group_ids: [12, 15]
      })
    )
  })
})
