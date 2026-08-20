import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import UserOrdersView from '../UserOrdersView.vue'
import type { PaymentOrder } from '@/types/payment'

const {
  getMyOrders,
  getRefundEligibleProviders,
  getInvoiceConfig,
  getCurrentInvoiceDraft,
  validateInvoiceOrders,
  checkInvoiceTaxPayment,
  applyInvoice,
  abandonInvoiceDraft,
  getInvoices,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getMyOrders: vi.fn(),
  getRefundEligibleProviders: vi.fn(),
  getInvoiceConfig: vi.fn(),
  getCurrentInvoiceDraft: vi.fn(),
  validateInvoiceOrders: vi.fn(),
  checkInvoiceTaxPayment: vi.fn(),
  applyInvoice: vi.fn(),
  abandonInvoiceDraft: vi.fn(),
  getInvoices: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getMyOrders,
    getRefundEligibleProviders,
    getInvoiceConfig,
    getCurrentInvoiceDraft,
    validateInvoiceOrders,
    checkInvoiceTaxPayment,
    applyInvoice,
    abandonInvoiceDraft,
    getInvoices,
    cancelOrder: vi.fn(),
    requestRefund: vi.fn(),
    cancelInvoice: vi.fn(),
    downloadInvoicePDF: vi.fn(),
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('@/utils/apiError', () => ({
  extractI18nErrorMessage: (_err: unknown, _t: unknown, _prefix: string, fallback: string) => fallback,
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>, fallback?: string) => {
        if (key === 'payment.invoice.selectedCount') return `${params?.count}/${params?.max}`
        return fallback || key
      },
    }),
  }
})

const completedOrder: PaymentOrder = {
  id: 101,
  user_id: 7,
  amount: 100,
  pay_amount: 102,
  fee_rate: 2,
  payment_type: 'alipay',
  out_trade_no: 'ORDER-101',
  status: 'COMPLETED',
  order_type: 'balance',
  created_at: '2026-08-05T00:00:00Z',
  expires_at: '2026-08-05T00:30:00Z',
  refund_amount: 0,
}

const AppLayoutStub = { template: '<div><slot /></div>' }
const BaseDialogStub = defineComponent({
  props: { show: Boolean, title: String },
  setup(props, { slots }) {
    return () => props.show ? h('section', { 'data-dialog-title': props.title }, [slots.default?.(), slots.footer?.()]) : null
  },
})
const OrderTableStub = defineComponent({
  props: { orders: { type: Array, default: () => [] } },
  setup(props, { slots }) {
    return () => h('div', { 'data-test': 'orders' }, (props.orders as PaymentOrder[]).map(row => h('div', { 'data-order-id': row.id }, slots.actions?.({ row }))))
  },
})
const SelectStub = defineComponent({
  inheritAttrs: false,
  props: {
    modelValue: { type: String, default: '' },
    options: { type: Array, default: () => [] },
  },
  emits: ['update:modelValue', 'change'],
  template: '<select :value="modelValue" />',
})

function mountView() {
  return mount(UserOrdersView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        BaseDialog: BaseDialogStub,
        OrderTable: OrderTableStub,
        Select: SelectStub,
        Pagination: true,
        Icon: true,
      },
    },
  })
}

describe('UserOrdersView invoice workflow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getMyOrders.mockResolvedValue({ data: { items: [completedOrder], total: 1, page: 1, page_size: 20, pages: 1 } })
    getRefundEligibleProviders.mockResolvedValue({ data: { provider_instance_ids: [] } })
    getInvoices.mockResolvedValue({ data: { items: [], total: 0, page: 1, page_size: 20, pages: 1 } })
    applyInvoice.mockResolvedValue({ data: { id: 1, status: 'pending' } })
    getCurrentInvoiceDraft.mockResolvedValue({ data: null })
    abandonInvoiceDraft.mockResolvedValue({ data: null })
  })

  it('hides invoice controls when the backend integration is disabled', async () => {
    getInvoiceConfig.mockResolvedValue({ data: { enabled: false, supports_tax_payment: true, max_orders: 20 } })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).not.toContain('payment.invoice.apply')
    expect(wrapper.text()).not.toContain('payment.invoice.records')
  })

  it('defaults to customer-paid fees without showing a payer selector', async () => {
    getInvoiceConfig.mockResolvedValue({
      data: { enabled: true, supports_tax_payment: true, max_orders: 20, fee_payer: 'customer' },
    })
    validateInvoiceOrders.mockResolvedValue({
      data: {
        draft_id: 8,
        order_ids: [101],
        need_pay_tax: true,
        tax_order_nos: [],
        validation: {
          totalAmount: '102.00',
          invoiceAmount: '108.12',
          currency: 'CNY',
          taxAmount: '6.12',
          taxDueAmount: '6.12',
        },
      },
    })

    const wrapper = mountView()
    await flushPromises()
    await wrapper.find('[data-order-id="101"] button').trigger('click')

    expect(wrapper.find('[data-test="invoice-tax-mode"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="invoice-fixed-fee-payer"]').text()).toContain('payment.invoice.taxRequired')
    await wrapper.findAll('button').find(button => button.text() === 'common.next')!.trigger('click')
    await flushPromises()

    expect(validateInvoiceOrders).toHaveBeenCalledWith([101], true)
  })

  it('marks an already invoiced order and excludes it from merged selection', async () => {
    getInvoiceConfig.mockResolvedValue({
      data: { enabled: true, supports_tax_payment: true, max_orders: 20, fee_payer: 'customer' },
    })
    getMyOrders.mockResolvedValue({
      data: {
        items: [{ ...completedOrder, invoice_status: 'completed' }],
        total: 1,
        page: 1,
        page_size: 20,
        pages: 1,
      },
    })

    const wrapper = mountView()
    await flushPromises()

    const row = wrapper.find('[data-order-id="101"]')
    expect(row.find('[data-test="invoice-order-status"]').text()).toContain('payment.invoice.orderStatus.completed')
    expect(row.text()).not.toContain('payment.invoice.apply')

    await wrapper.findAll('button').find(button => button.text() === 'payment.invoice.apply')!.trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-order-id="101"] input[type="checkbox"]').exists()).toBe(false)
    expect(wrapper.find('[data-order-id="101"] [data-test="invoice-order-status"]').exists()).toBe(true)
  })

  it('requires confirmed tax reconciliation before submitting buyer data', async () => {
    getInvoiceConfig.mockResolvedValue({ data: { enabled: true, supports_tax_payment: true, max_orders: 20, fee_payer: 'user_choice' } })
    validateInvoiceOrders.mockResolvedValue({
      data: {
        draft_id: 9,
        order_ids: [101],
        need_pay_tax: true,
        tax_order_nos: [],
        validation: {
          totalAmount: '102.00',
          invoiceAmount: '108.12',
          currency: 'CNY',
          taxAmount: '6.12',
          taxPaidAmount: '0.00',
          taxDueAmount: '6.12',
          taxPayments: {
            alipay: { taxOrderNo: 'INV-TAX-9', payUrl: 'https://pay.example.test/invoice-9' },
          },
        },
      },
    })
    checkInvoiceTaxPayment.mockResolvedValue({
      data: {
        paid: true,
        ready: true,
        tax_order_nos: ['INV-TAX-9'],
        validation: {
          totalAmount: '102.00',
          invoiceAmount: '108.12',
          currency: 'CNY',
          taxAmount: '6.12',
          taxPaidAmount: '6.12',
          taxDueAmount: '0.00',
        },
      },
    })
    const open = vi.spyOn(window, 'open').mockImplementation(() => null)

    const wrapper = mountView()
    await flushPromises()

    const rowApply = wrapper.find('[data-order-id="101"] button')
    expect(rowApply.exists()).toBe(true)
    await rowApply.trigger('click')
    expect(wrapper.find('[data-test="invoice-selection-bar"]').text()).toContain('1/20')
    await wrapper.findAll('button').find(button => button.text() === 'payment.invoice.taxRequired')!.trigger('click')
    await wrapper.findAll('button').find(button => button.text() === 'common.next')!.trigger('click')
    await flushPromises()

    expect(validateInvoiceOrders).toHaveBeenCalledWith([101], true)
    expect(wrapper.find('[data-test="invoice-progress"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="invoice-tax-step"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="invoice-final-amount"]').text()).toContain('CNY 108.12')
    expect(wrapper.find('[data-test="invoice-amount-summary"]').text()).toContain('CNY 102.00')
    expect(wrapper.find('[data-test="invoice-amount-summary"]').text()).toContain('CNY 6.12')
    expect(wrapper.find('#invoice-title').exists()).toBe(false)

    const taxPaymentOption = wrapper.find('[data-test="invoice-tax-payment-option"]')
    expect(taxPaymentOption.attributes('aria-pressed')).toBe('false')
    await taxPaymentOption.trigger('click')
    expect(taxPaymentOption.attributes('aria-pressed')).toBe('true')
    expect(open).toHaveBeenCalledWith('https://pay.example.test/invoice-9', '_blank', 'noopener,noreferrer')
    await wrapper.findAll('button').find(button => button.text() === 'payment.invoice.checkTaxPayment')!.trigger('click')
    await flushPromises()

    expect(checkInvoiceTaxPayment).toHaveBeenCalledWith(9, 'INV-TAX-9')
    expect(wrapper.find('[data-test="invoice-buyer-form"]').exists()).toBe(true)
    expect(wrapper.find('#invoice-title').exists()).toBe(true)

    await wrapper.find('#invoice-title').setValue('Example Technology Ltd.')
    await wrapper.find('#invoice-taxpayer-id').setValue('91310000677833266F')
    await wrapper.find('#invoice-email').setValue('invoice@example.test')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(applyInvoice).toHaveBeenCalledWith(9, expect.objectContaining({
      buyer_type: 'company',
      title: 'Example Technology Ltd.',
      taxpayer_id: '91310000677833266F',
      recipient_email: 'invoice@example.test',
    }))
    expect(showSuccess).toHaveBeenCalledWith('payment.invoice.applicationSubmitted')
    expect(getMyOrders).toHaveBeenCalledTimes(2)
  })

  it('renders invoice records in mobile cards and a desktop table', async () => {
    getInvoiceConfig.mockResolvedValue({ data: { enabled: true, supports_tax_payment: true, max_orders: 20, fee_payer: 'customer' } })
    getInvoices.mockResolvedValue({
      data: {
        items: [{
          id: 33,
          external_id: 'INV-33',
          order_ids: [101],
          order_nos: ['ORDER-101'],
          need_pay_tax: false,
          tax_order_nos: [],
          status: 'pending',
          title: 'Example Technology Ltd.',
          total_amount: '102.00',
          currency: 'CNY',
          created_at: '2026-08-05T00:00:00Z',
          updated_at: '2026-08-05T00:00:00Z',
        }],
        total: 1,
        page: 1,
        page_size: 20,
        pages: 1,
      },
    })

    const wrapper = mountView()
    await flushPromises()
    await wrapper.findAll('button').find(button => button.text() === 'payment.invoice.records')!.trigger('click')
    await flushPromises()

    const cards = wrapper.find('[data-test="invoice-record-cards"]')
    const table = wrapper.find('[data-test="invoice-record-table"]')
    expect(cards.text()).toContain('INV-33')
    expect(cards.text()).toContain('payment.invoice.cancelApplication')
    expect(table.text()).toContain('Example Technology Ltd.')
  })

  it('uses the platform-paid branch without showing a user override', async () => {
    getInvoiceConfig.mockResolvedValue({ data: { enabled: true, supports_tax_payment: true, max_orders: 20, fee_payer: 'platform' } })
    validateInvoiceOrders.mockResolvedValue({
      data: {
        draft_id: 12,
        order_ids: [101],
        need_pay_tax: false,
        tax_order_nos: [],
        validation: { totalAmount: '102.00', invoiceAmount: '102.00', currency: 'CNY' },
      },
    })

    const wrapper = mountView()
    await flushPromises()
    await wrapper.find('[data-order-id="101"] button').trigger('click')
    expect(wrapper.find('[data-test="invoice-tax-mode"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="invoice-fixed-fee-payer"]').text()).toContain('payment.invoice.taxNotRequired')
    await wrapper.findAll('button').find(button => button.text() === 'common.next')!.trigger('click')
    await flushPromises()

    expect(validateInvoiceOrders).toHaveBeenCalledWith([101], false)
    expect(wrapper.find('[data-test="invoice-tax-step"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="invoice-buyer-form"]').exists()).toBe(true)
  })

  it('keeps the selected WeChat tax checkout order for reconciliation', async () => {
    getInvoiceConfig.mockResolvedValue({ data: { enabled: true, supports_tax_payment: true, max_orders: 20, fee_payer: 'user_choice' } })
    validateInvoiceOrders.mockResolvedValue({
      data: {
        draft_id: 18,
        order_ids: [101],
        need_pay_tax: true,
        tax_order_nos: [],
        validation: {
          totalAmount: '102.00',
          invoiceAmount: '108.12',
          currency: 'CNY',
          taxAmount: '6.12',
          taxPaidAmount: '0.00',
          taxDueAmount: '6.12',
          taxPayments: {
            alipay: { taxOrderNo: 'INV-TAX-ALIPAY-18', payUrl: 'https://pay.example.test/alipay-18' },
            wxpay: { taxOrderNo: 'INV-TAX-WXPAY-18', payUrl: 'https://pay.example.test/wxpay-18' },
          },
        },
      },
    })
    checkInvoiceTaxPayment.mockResolvedValue({
      data: {
        paid: true,
        ready: true,
        tax_order_nos: ['INV-TAX-WXPAY-18'],
        validation: {
          totalAmount: '102.00',
          invoiceAmount: '108.12',
          currency: 'CNY',
          taxAmount: '6.12',
          taxPaidAmount: '6.12',
          taxDueAmount: '0.00',
        },
      },
    })
    const open = vi.spyOn(window, 'open').mockImplementation(() => null)

    const wrapper = mountView()
    await flushPromises()
    await wrapper.find('[data-order-id="101"] button').trigger('click')
    await wrapper.findAll('button').find(button => button.text() === 'payment.invoice.taxRequired')!.trigger('click')
    await wrapper.findAll('button').find(button => button.text() === 'common.next')!.trigger('click')
    await flushPromises()

    const options = wrapper.findAll('[data-test="invoice-tax-payment-option"]')
    expect(options).toHaveLength(2)
    await options[1].trigger('click')
    expect(open).toHaveBeenCalledWith('https://pay.example.test/wxpay-18', '_blank', 'noopener,noreferrer')
    await wrapper.findAll('button').find(button => button.text() === 'payment.invoice.checkTaxPayment')!.trigger('click')
    await flushPromises()

    expect(checkInvoiceTaxPayment).toHaveBeenCalledWith(18, 'INV-TAX-WXPAY-18')
    expect(wrapper.find('[data-test="invoice-buyer-form"]').exists()).toBe(true)
  })

  it('restores an unfinished draft without creating a new checkout', async () => {
    getInvoiceConfig.mockResolvedValue({ data: { enabled: true, supports_tax_payment: true, max_orders: 20, fee_payer: 'customer' } })
    getCurrentInvoiceDraft.mockResolvedValue({
      data: {
        draft_id: 17,
        order_ids: [101],
        need_pay_tax: true,
        tax_order_nos: [],
        validation: {
          totalAmount: '102.00',
          invoiceAmount: '108.12',
          currency: 'CNY',
          taxAmount: '6.12',
          taxDueAmount: '6.12',
          taxPayments: {
            alipay: { taxOrderNo: 'INV-TAX-17', payUrl: 'https://pay.example.test/invoice-17' },
          },
        },
      },
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-test="invoice-draft-resume"]').exists()).toBe(true)
    await wrapper.findAll('button').find(button => button.text() === 'payment.invoice.resumeDraft')!.trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-test="invoice-tax-step"]').exists()).toBe(true)
    expect(validateInvoiceOrders).not.toHaveBeenCalled()
  })
})
