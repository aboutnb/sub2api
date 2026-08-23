/**
 * User Payment API endpoints
 * Handles payment operations for regular users
 */

import { apiClient } from './client'
import type {
  PaymentConfig,
  SubscriptionPlan,
  MethodLimitsResponse,
  CheckoutInfoResponse,
  CreateOrderRequest,
  CreateOrderResult,
  PaymentOrder,
  InvoiceConfig,
  InvoiceDraft,
  InvoiceTaxStatus,
  InvoiceApplication,
  InvoiceApplyRequest,
  USDTConfigResponse,
  USDTOrder,
} from '@/types/payment'
import type { BasePaginationResponse } from '@/types'

export interface PublicOrderVerifyResult {
  out_trade_no: string
  status: string
  paid: boolean
  created_at: string
  expires_at: string
}

export const paymentAPI = {
  /** Get payment configuration (enabled types, limits, etc.) */
  getConfig() {
    return apiClient.get<PaymentConfig>('/payment/config')
  },

  /** Get available subscription plans */
  getPlans() {
    return apiClient.get<SubscriptionPlan[]>('/payment/plans')
  },

  /** Get all checkout page data in a single call */
  getCheckoutInfo() {
    return apiClient.get<CheckoutInfoResponse>('/payment/checkout-info')
  },

  /** Isolated BEpusdt USDT capabilities; independent from RMB provider config. */
  getUSDTConfig() {
    return apiClient.get<USDTConfigResponse>('/usdt/config')
  },

  createUSDTOrder(data: { amount: string; amount_unit?: 'USDT'; network: string; order_type?: string; plan_id?: number; return_url?: string; payment_source?: string }, idempotencyKey: string) {
    return apiClient.post<USDTOrder>('/usdt/orders', data, {
      headers: { 'Idempotency-Key': idempotencyKey }
    })
  },

  getUSDTOrder(id: number) {
    return apiClient.get<USDTOrder>(`/usdt/orders/${id}`)
  },

  cancelUSDTOrder(id: number) {
    return apiClient.post(`/usdt/orders/${id}/cancel`)
  },

  /** Get payment method limits and fee rates */
  getLimits() {
    return apiClient.get<MethodLimitsResponse>('/payment/limits')
  },

  /** Create a new payment order */
  createOrder(data: CreateOrderRequest) {
    return apiClient.post<CreateOrderResult>('/payment/orders', data)
  },

  /** Get current user's orders */
  getMyOrders(params?: { page?: number; page_size?: number; status?: string }) {
    return apiClient.get<BasePaginationResponse<PaymentOrder>>('/payment/orders/my', { params })
  },

  /** Get a specific order by ID */
  getOrder(id: number) {
    return apiClient.get<PaymentOrder>(`/payment/orders/${id}`)
  },

  /** Cancel a pending order */
  cancelOrder(id: number) {
    return apiClient.post(`/payment/orders/${id}/cancel`)
  },

  /** Verify order payment status with upstream provider */
  verifyOrder(outTradeNo: string) {
    return apiClient.post<PaymentOrder>('/payment/orders/verify', { out_trade_no: outTradeNo })
  },

  /** Legacy-compatible public order lookup by out_trade_no */
  verifyOrderPublic(outTradeNo: string) {
    return apiClient.post<PublicOrderVerifyResult>('/payment/public/orders/verify', { out_trade_no: outTradeNo })
  },

  /** Resolve an order from a signed resume token without auth */
  resolveOrderPublicByResumeToken(resumeToken: string) {
    return apiClient.post<PublicOrderVerifyResult>('/payment/public/orders/resolve', { resume_token: resumeToken })
  },

  /** Request a refund for a completed order */
  requestRefund(id: number, data: { reason: string }) {
    return apiClient.post(`/payment/orders/${id}/refund-request`, data)
  },

  /** Get provider instance IDs that allow user refund */
  getRefundEligibleProviders() {
    return apiClient.get<{ provider_instance_ids: string[] }>('/payment/orders/refund-eligible-providers')
  },

  /** Get the server-side invoice integration state (never includes credentials). */
  getInvoiceConfig() {
    return apiClient.get<InvoiceConfig>('/payment/invoices/config')
  },

  /** Validate owned completed orders and create a local invoice draft. */
  validateInvoiceOrders(orderIds: number[], needPayTax: boolean) {
    return apiClient.post<InvoiceDraft>('/payment/invoices/validate', {
      order_ids: orderIds,
      need_pay_tax: needPayTax,
    })
  },

  /** Return the caller's unfinished invoice draft, if one exists. */
  getCurrentInvoiceDraft() {
    return apiClient.get<InvoiceDraft | null>('/payment/invoices/drafts/current')
  },

  /** Confirm one tax checkout and reconcile the original order set. */
  checkInvoiceTaxPayment(draftId: number, taxOrderNo: string) {
    return apiClient.post<InvoiceTaxStatus>(`/payment/invoices/drafts/${draftId}/tax-status`, {
      tax_order_no: taxOrderNo,
    })
  },

  /** Submit buyer data for a validated invoice draft. */
  applyInvoice(draftId: number, data: InvoiceApplyRequest) {
    return apiClient.post<InvoiceApplication>(`/payment/invoices/drafts/${draftId}/apply`, data)
  },

  /** Release an unfinished draft only when it has no invoice-fee checkout. */
  abandonInvoiceDraft(draftId: number) {
    return apiClient.post(`/payment/invoices/drafts/${draftId}/abandon`)
  },

  /** List the current user's locally-owned invoice applications. */
  getInvoices(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<InvoiceApplication>>('/payment/invoices', { params })
  },

  cancelInvoice(id: number) {
    return apiClient.post<InvoiceApplication>(`/payment/invoices/${id}/cancel`)
  },

  downloadInvoicePDF(id: number) {
    return apiClient.get<Blob>(`/payment/invoices/${id}/pdf`, { responseType: 'blob' })
  }
}
