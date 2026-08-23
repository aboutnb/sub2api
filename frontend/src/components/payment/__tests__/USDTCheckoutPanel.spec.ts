import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const getUSDTOrder = vi.hoisted(() => vi.fn())
const createUSDTOrder = vi.hoisted(() => vi.fn())
const cancelUSDTOrder = vi.hoisted(() => vi.fn())
const toCanvas = vi.hoisted(() => vi.fn().mockResolvedValue(undefined))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: { getUSDTOrder, createUSDTOrder, cancelUSDTOrder },
}))

vi.mock('qrcode', () => ({
  default: { toCanvas },
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
    cancelUSDTOrder.mockReset()
    toCanvas.mockClear()
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
    expect(wrapper.text()).toContain('USDT · TRON')
    expect(wrapper.text()).toContain('TRC-20')
    expect(toCanvas).toHaveBeenCalledWith(expect.anything(), 'TAddress', expect.objectContaining({ width: 160 }))
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
        config: { enabled: true, rate: '7.2', networks: [] },
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
        config: { enabled: true, rate: '7.2', networks: [] },
      },
      global: { stubs: { Icon: true } },
    })

    await flushPromises()
    const text = wrapper.text()
    const amountInput = wrapper.findComponent({ name: 'AmountInput' })
    expect(text).toContain('payment.quickAmounts')
    expect(text).toContain('₮50')
    expect(text).toContain('payment.usdt.exchangeRate')
    expect(text).toContain('¥7.200000 / USDT')
    expect(amountInput.props('amounts')).toEqual([10, 20, 50, 100, 200, 500])
    expect(text).not.toContain('payment.usdt.orderType')
    expect(text).not.toContain('payment.usdt.plan')
    wrapper.unmount()
  })

  it('submits the entered value as USDT while showing its CNY estimate', async () => {
    createUSDTOrder.mockResolvedValue({ data: order() })
    const wrapper = mount(USDTCheckoutPanel, {
      props: {
        config: {
          enabled: true,
          rate: '7.2',
          bonus_percent: 10,
          networks: [{ network: 'tron', network_name: 'TRON', trade_type: 'usdt.trc20', accepting_orders: true, crypto: 'USDT', wallet_count: 1, rpc_endpoint_set: true }],
        },
      },
      global: { stubs: { Icon: true } },
    })

    await flushPromises()
    expect(wrapper.text()).toContain('USDT · TRON')
    expect(wrapper.text()).toContain('TRC-20')
    expect(wrapper.text()).toContain('payment.usdt.bonusLine')
    expect(wrapper.text()).toContain('payment.usdt.estimatedBalance')
    await wrapper.findAll('button').find(button => button.text().includes('TRON'))?.trigger('click')
    const createButton = wrapper.findAll('button').find(button => button.text().includes('payment.usdt.confirmPayment'))
    await createButton?.trigger('click')
    await flushPromises()

    expect(createUSDTOrder).toHaveBeenCalledWith(
      expect.objectContaining({ amount: '10', amount_unit: 'USDT', network: 'tron' }),
      expect.stringMatching(/^usdt-order-/),
    )
    expect(wrapper.text()).toContain('TAddress')
    expect(wrapper.find('a[target="_blank"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('opens native cashier checkout in a popup instead of embedding it', async () => {
    const popup = { closed: false, location: { href: '' }, focus: vi.fn(), close: vi.fn() } as unknown as Window
    vi.spyOn(window, 'open').mockReturnValue(popup)
    createUSDTOrder.mockResolvedValue({ data: { ...order(), payment_mode: 'cashier' } })
    const wrapper = mount(USDTCheckoutPanel, {
      props: {
        config: { enabled: true, checkout_mode: 'cashier', rate: '7.2', networks: [] },
      },
      global: { stubs: { Icon: true } },
    })

    await wrapper.findAll('button').find(button => button.text().includes('payment.usdt.confirmPayment'))?.trigger('click')
    await flushPromises()

    expect(window.open).toHaveBeenCalledWith('about:blank', 'sub2api-usdt-cashier', expect.stringContaining('popup'))
    expect(wrapper.find('iframe').exists()).toBe(false)
    expect(wrapper.text()).toContain('payment.usdt.cashierPopupTitle')
    wrapper.unmount()
  })

  it('does not show network availability prompts in native cashier mode', async () => {
    const wrapper = mount(USDTCheckoutPanel, {
      props: {
        config: { enabled: true, checkout_mode: 'cashier', rate: '7.2', networks: [] },
      },
      global: { stubs: { Icon: true } },
    })

    await flushPromises()
    expect(wrapper.text()).not.toContain('payment.usdt.unavailable')
    expect(wrapper.text()).not.toContain('payment.usdt.networkInstruction')
    wrapper.unmount()
  })

  it('cancels a pending USDT order and stops showing it as payable', async () => {
    createUSDTOrder.mockResolvedValue({ data: order() })
    cancelUSDTOrder.mockResolvedValue({ data: { cancelled: true } })
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
    await wrapper.findAll('button').find(button => button.text().includes('payment.usdt.confirmPayment'))?.trigger('click')
    await flushPromises()
    await wrapper.findAll('button').find(button => button.text().includes('payment.usdt.cancel'))?.trigger('click')
    expect(wrapper.text()).toContain('payment.usdt.cancelPrompt')
    await wrapper.findAll('button').find(button => button.text().includes('payment.usdt.confirmCancel'))?.trigger('click')
    await flushPromises()

    expect(cancelUSDTOrder).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('payment.status.cancelled')
    expect(wrapper.find('canvas').exists()).toBe(false)
    wrapper.unmount()
  })
})
