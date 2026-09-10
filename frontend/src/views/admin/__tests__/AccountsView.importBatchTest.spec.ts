import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AccountsView from '../AccountsView.vue'

const { getAccount, getBatchTodayStats, getAllGroups, getAllProxies, listAccounts, listWithEtag } = vi.hoisted(() => ({
  getAccount: vi.fn(),
  getBatchTodayStats: vi.fn(),
  getAllGroups: vi.fn(),
  getAllProxies: vi.fn(),
  listAccounts: vi.fn(),
  listWithEtag: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      getById: getAccount,
      getBatchTodayStats,
      list: listAccounts,
      listWithEtag
    },
    groups: { getAll: getAllGroups },
    proxies: { getAll: getAllProxies }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showInfo: vi.fn(),
    showSuccess: vi.fn()
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ token: 'test-token' })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

const ImportDataModalStub = {
  props: ['show'],
  emits: ['close', 'imported'],
  template: '<button data-testid="finish-import" @click="$emit(\'imported\', [101, 102])">finish import</button>'
}

const BatchAccountTestModalStub = {
  props: ['show', 'accounts', 'autoStart'],
  template: '<div data-testid="batch-test-modal" :data-show="String(show)" :data-auto-start="String(autoStart)" :data-account-ids="accounts.map(account => account.id).join(\',\')"></div>'
}

describe('AccountsView import batch test flow', () => {
  beforeEach(() => {
    localStorage.clear()
    listAccounts.mockReset()
    listWithEtag.mockReset()
    getAccount.mockReset()
    getBatchTodayStats.mockReset()
    getAllGroups.mockReset()
    getAllProxies.mockReset()

    listAccounts.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20, pages: 0 })
    listWithEtag.mockResolvedValue({ notModified: true, etag: null, data: null })
    getBatchTodayStats.mockResolvedValue({ stats: {} })
    getAllGroups.mockResolvedValue([])
    getAllProxies.mockResolvedValue([])
    getAccount.mockImplementation(async (id: number) => ({
      id,
      name: `imported-${id}`,
      platform: 'openai',
      type: 'oauth',
      status: 'active'
    }))
  })

  it('loads only imported accounts and opens their batch test', async () => {
    const wrapper = mount(AccountsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: true,
          Pagination: true,
          ConfirmDialog: true,
          AccountTableActions: true,
          AccountTableFilters: true,
          AccountBulkActionsBar: true,
          AccountActionMenu: true,
          ImportDataModal: ImportDataModalStub,
          ReAuthAccountModal: true,
          AccountTestModal: true,
          BatchAccountTestModal: BatchAccountTestModalStub,
          AccountStatsModal: true,
          ScheduledTestsPanel: true,
          SyncFromCrsModal: true,
          TempUnschedStatusModal: true,
          ErrorPassthroughRulesModal: true,
          TLSFingerprintProfilesModal: true,
          CreateAccountModal: true,
          EditAccountModal: true,
          BulkEditAccountModal: true,
          PlatformTypeBadge: true,
          AccountCapacityCell: true,
          AccountStatusIndicator: true,
          AccountTodayStatsCell: true,
          AccountGroupsCell: true,
          AccountUsageCell: true,
          Icon: true
        }
      }
    })
    await flushPromises()

    await wrapper.get('[data-testid="finish-import"]').trigger('click')
    await flushPromises()
    await flushPromises()

    expect(getAccount.mock.calls.map(call => call[0])).toEqual([101, 102])
    const batchModal = wrapper.get('[data-testid="batch-test-modal"]')
    expect(batchModal.attributes('data-show')).toBe('true')
    expect(batchModal.attributes('data-account-ids')).toBe('101,102')
  })
})
