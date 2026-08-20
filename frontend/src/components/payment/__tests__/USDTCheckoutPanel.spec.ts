import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const getUSDTOrder = vi.hoisted(() => vi.fn())
const createUSDTOrder = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: { getUSDTOrder, createUSDTOrder },
}))

import USDTCheckoutPanel from '../USDTCheckoutPanel.vue'

const order = (status = 'PENDING') => ({
  order_id: 42,
  out_trade_no: 'sub2_usdt_42',
  amount: 100,
  pay_amount: 102,
  fee_rate: 2,
  status,
  payment_type: 'usdt' as const,
  fiat_currency: 'CNY',
  fiat_amount: '102',
  crypto_currency: 'USDT' as const,
  crypto_amount: '14.5714',
  network: 'tron',
  trade_type: 'usdt.trc20',
  receiving_address: 'TAddress',
  exchange_rate: '7',
  payment_url: 'https://pay.example/checkout/trade-42',
  expires_at: '2099-01-01T00:30:00Z',
})

describe('USDTCheckoutPanel', () => {
  beforeEach(() => {
    localStorage.clear()
    getUSDTOrder.mockReset()
    createUSDTOrder.mockReset()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('restores a non-terminal order after the page is reopened', async () => {
    localStorage.setItem('usdt.payment.current', JSON.stringify({
      orderId: 42,
      outTradeNo: 'sub2_usdt_42',
      savedAt: Date.now(),
    }))
    getUSDTOrder.mockResolvedValue({ data: order() })

    const wrapper = mount(USDTCheckoutPanel, {
      props: {
        config: {
          enabled: true,
          rate: '7.2',
          networks: [{ network: 'tron', network_name: 'TRON', trade_type: 'usdt.trc20', accepting_orders: true, crypto: 'USDT', wallet_count: 1, rpc_endpoint_set: true }],
        },
      },
      global: { stubs: { Icon: true } },
    })

    await flushPromises()
    expect(getUSDTOrder).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('TAddress')
    expect(localStorage.getItem('usdt.payment.current')).toContain('sub2_usdt_42')
    wrapper.unmount()
  })

  it('clears the recovery snapshot once the restored order is terminal', async () => {
    localStorage.setItem('usdt.payment.current', JSON.stringify({
      orderId: 42,
      outTradeNo: 'sub2_usdt_42',
      savedAt: Date.now(),
    }))
    getUSDTOrder.mockResolvedValue({ data: order('COMPLETED') })

    const wrapper = mount(USDTCheckoutPanel, {
      props: {
        config: { enabled: true, networks: [] },
      },
      global: { stubs: { Icon: true } },
    })

    await flushPromises()
    expect(localStorage.getItem('usdt.payment.current')).toBeNull()
    wrapper.unmount()
  })

  it('offers balance-only quick amounts without subscription controls', async () => {
    const wrapper = mount(USDTCheckoutPanel, {
      props: {
        config: { enabled: true, networks: [] },
      },
      global: { stubs: { Icon: true } },
    })

    await flushPromises()
    const text = wrapper.text()
    const amountInput = wrapper.findComponent({ name: 'AmountInput' })
    expect(text).toContain('payment.quickAmounts')
    expect(text).toContain('₮50')
    expect(text).toContain('payment.usdt.exchangeRate')
    expect(amountInput.props('amounts')).toEqual([10, 20, 50, 100, 200, 500])
    expect(text).not.toContain('payment.usdt.orderType')
    expect(text).not.toContain('payment.usdt.plan')
    wrapper.unmount()
  })

  it('submits the entered value as USDT while showing its CNY estimate', async () => {
    vi.spyOn(window, 'open').mockImplementation(() => null)
    createUSDTOrder.mockResolvedValue({ data: order() })
    const wrapper = mount(USDTCheckoutPanel, {
      props: {
        config: {
          enabled: true,
          rate: '7.2',
          networks: [{ network: 'tron', network_name: 'TRON', trade_type: 'usdt.trc20', accepting_orders: true, crypto: 'USDT', wallet_count: 1, rpc_endpoint_set: true }],
        },
      },
      global: { stubs: { Icon: true } },
    })

    await flushPromises()
    await wrapper.findAll('button').find(button => button.text().includes('TRON'))?.trigger('click')
    const createButton = wrapper.findAll('button').find(button => button.text().includes('payment.usdt.create'))
    await createButton?.trigger('click')
    await flushPromises()

    expect(createUSDTOrder).toHaveBeenCalledWith(expect.objectContaining({ amount: 10, amount_unit: 'USDT', network: 'tron' }))
    wrapper.unmount()
  })
})
