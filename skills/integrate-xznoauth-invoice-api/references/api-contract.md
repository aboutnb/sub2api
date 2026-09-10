# XZNOAuth Self-Service Invoice API Contract

Use this reference when writing a client, request models, fixtures, or integration tests. The production base URL is `https://oauth.xzncraft.cn`; make it configurable for staging and private deployments.

## Common Conventions

Authenticated invoice endpoints require:

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

Most successful responses use:

```json
{ "code": 0, "message": "ok", "data": {} }
```

Errors use:

```json
{ "code": "ERROR_CODE", "message": "...", "requestId": "..." }
```

The OAuth token endpoint is the exception: it returns the OAuth token object directly rather than the API envelope.

## Prerequisites and Token

Create a developer application, request the `invoice.apply` scope, and wait for administrator approval. A server-to-server invoicing application does not need a redirect URI. Keep the Client Secret server-side and request a token as follows:

```http
POST https://oauth.xzncraft.cn/api/oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&client_id=<client_id>&client_secret=<client_secret>&scope=invoice.apply
```

Successful response:

```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 900,
  "scope": "invoice.apply"
}
```

`expires_in` is configuration-dependent; the example value is illustrative. Cache the token in memory or the target service's protected secret cache until shortly before the returned lifetime, and obtain a new client Token rather than expecting a Refresh Token. Application configuration or approval changes and Client Secret rotation can invalidate a Token before its nominal expiry. Do not persist it in a browser or log it.

## Choose Integration Tax Coverage

Before implementing a new integration, ask the user to select one scope unless the request already states it:

- Only invoices that do not require invoice-tax payment: send `needPayTax: false` or omit it, and do not implement checkout or tax-status polling.
- Only invoices that require invoice-tax payment: send `needPayTax: true` and implement checkout, payment-status polling, reconciliation, and `taxOrderNos` submission.
- Both modes (recommended): select the branch per invoice request through `needPayTax` and test both branches.

Do not describe these choices only as “tax-inclusive” and “tax-exclusive”; those terms can be interpreted as invoice-price semantics rather than whether this workflow must collect invoice tax. This choice limits the client integration, not the server API, which supports both modes. Do not infer the mode from `buyerType`, order amount, or environment configuration.

## Validate Orders and Create Tax Checkout

```http
POST /api/v1/invoice-orders/validate
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "orderNos": ["YOUR_PAID_ORDER_NO"],
  "needPayTax": true
}
```

`orderNos` must contain one to twenty unique, non-empty order numbers, each no longer than 128 characters. Platform order numbers, external order numbers, and WeChat/Alipay transaction IDs are accepted. Only paid orders with valid CNY amounts and no existing invoice claim are returned. When available, each returned order includes the matched payment-channel identifier as `transactionId`. To reconcile a previously paid tax checkout, pass its exact `INV-TAX-...` value(s) in `taxOrderNos`; the server verifies them, returns the accepted list as `taxOrderNos`, and deducts their paid amount before creating any new checkout.

Without tax, the response data contains:

```json
{
  "orders": [
    {
      "platformOrderId": 123,
      "orderNo": "ORDER-001",
      "externalNo": "YOUR_PAID_ORDER_NO",
      "productName": "商品",
      "amount": "5.00",
      "currency": "CNY",
      "paidAt": "2026-01-01T12:00:00Z",
      "verifiedAt": "2026-01-01T12:00:01Z"
    }
  ],
  "totalAmount": "5.00",
  "currency": "CNY"
}
```

With `needPayTax: true`, the server calculates tax using the administrator-configured rate (default 6%, rounded to cents), returns `taxAmount`, `taxPaidAmount`, and `taxDueAmount`, and creates new payment orders only when `taxDueAmount` is greater than `0.00`:

```json
{
  "taxAmount": "0.30",
  "taxPaidAmount": "0.00",
  "taxDueAmount": "0.30",
  "taxPayments": {
    "alipay": { "taxOrderNo": "INV-TAX-...", "payUrl": "https://pay.xzncraft.cn/cashier/..." },
    "wxpay": { "taxOrderNo": "INV-TAX-...", "payUrl": "https://pay.xzncraft.cn/cashier/..." }
  },
  "taxOrderNo": "INV-TAX-...",
  "payUrl": "https://pay.xzncraft.cn/cashier/..."
}
```

The checkout response only creates tax payment orders. It does not reserve the source orders or create an invoice application. Present one selected channel's `payUrl` to the user and keep that exact channel's `taxOrderNo` for the later validation and submission. The top-level `taxOrderNo` and `payUrl` are retained as an Alipay compatibility alias.

If previously paid tax orders total more than the tax currently required by `orderNos`, validation returns `INVOICE_TAX_PAYMENT_EXCEEDS_REQUIRED`. Preserve those paid tax order numbers and restore the original source-order set or restart the invoice flow; never drop a paid tax order and proceed with a different claim.

## Check Tax Payment Status

An application can poll one tax order with its `invoice.apply` client Token:

```http
POST /api/v1/invoice-tax-payments/status
Authorization: Bearer <access_token>
Content-Type: application/json

{ "taxOrderNo": "INV-TAX-..." }
```

HTTP `200` uses the normal successful envelope and contains `{ "paid": true }` or `{ "paid": false }`. This endpoint only checks status, does not create an order, and is rate limited by the OAuth `client_id` represented by the Token. Use bounded polling with backoff and stop or slow down on `429`; do not assume a fixed polling interval. A valid `INV-TAX-...` value is an unguessable checkout identifier, but status alone is not submission proof: call `/api/v1/invoice-orders/validate` again with the original `orderNos`, `needPayTax: true`, and every paid `taxOrderNos`, then require `taxDueAmount: "0.00"` before submitting.

The separate `/api/v1/me/invoice-tax-payments/status` route belongs to the logged-in user center and requires a user Access Token. Do not call a `/me/...` route with a client Token.

## Submit an Invoice Application

```http
POST /api/v1/invoices
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "orderNos": ["YOUR_PAID_ORDER_NO"],
  "needPayTax": true,
  "taxOrderNos": ["INV-TAX-..."],
  "buyerType": "company",
  "title": "示例科技有限公司",
  "taxpayerId": "91310000677833266F",
  "buyerAddress": "上海市示例路 1 号",
  "buyerPhone": "021-12345678",
  "buyerBank": "示例银行上海分行",
  "buyerBankAccount": "6222000000000000000",
  "recipientEmail": "invoice@example.com"
}
```

The server re-verifies every source order before creating the application. When tax is required it also re-queries every `taxOrderNos` value and requires all of the following: paid status, the current configured tax amount in total, currency `CNY`, product name `发票税费`, and the system-generated `INV-TAX-` external-number prefix. Send the exact paid tax order numbers unchanged from validation; `taxOrderNo` is accepted as a single-order compatibility alias. Do not treat the browser's return from the payment page as proof of payment.

Field rules:

| Field | Rule |
| --- | --- |
| `orderNos` | One to twenty unique order numbers; preserve the validated list. |
| `needPayTax` | Optional boolean, default `false`. |
| `taxOrderNos` | One to twenty verified, paid tax order numbers; required when `needPayTax=true`; omit otherwise. |
| `taxOrderNo` | Legacy single-order alias; prefer `taxOrderNos` for new integrations. |
| `buyerType` | `individual` or `company`. |
| `title` | Required, maximum 255 characters; use the buyer's name or company name. The administrator reviews it manually. |
| `taxpayerId` | Required, maximum 32 characters; send the supplied resident ID or company taxpayer ID. The administrator reviews it manually. |
| `buyerAddress` | Optional, maximum 255 characters. |
| `buyerPhone` | Optional valid mobile or landline. |
| `buyerBank` | Optional bank name, maximum 255 characters. |
| `buyerBankAccount` | Optional 8-32 digit account; spaces and hyphens are normalized. |
| `recipientEmail` | Required valid email address. |

A successful submission returns HTTP `201` and the normal envelope with an application whose initial status is `pending`. The source order and tax order are claimed transactionally, so a duplicate submission can return a conflict.

After an administrator uploads the completed invoice PDF, the platform sends it directly as an attachment to `recipientEmail`. The completion email includes the buyer name, issue date, and total amount and does not require the recipient to sign in.

## List Applications and Download PDFs

```http
GET /api/v1/invoices?page=1&pageSize=20
Authorization: Bearer <access_token>
```

The data object is:

```json
{ "items": [], "total": 0, "page": 1, "pageSize": 20 }
```

Only records submitted by the current Client ID are returned; owner user IDs and email addresses are omitted. Valid statuses are `pending`, `approved`, `rejected`, `completed`, and `canceled`. Use the application ID from the list to download a completed invoice:

```http
GET /api/v1/invoices/<application_id>/pdf
Authorization: Bearer <access_token>
Accept: application/pdf
```

This response is binary `application/pdf`, not the JSON envelope. Download is allowed only for the current application and `completed` status; stream it to a file or object store with a private content disposition.

## Cancel an Application

An application can cancel an invoice that it submitted:

```http
POST /api/v1/invoices/<application_id>/cancel
Authorization: Bearer <access_token>
```

This endpoint has no request body. It accepts only applications in `pending` or `approved` status. HTTP `200` uses the normal JSON envelope and returns the updated application with `status: "canceled"`. Cancellation transactionally releases all source-order and tax-order claims, so the orders can become eligible for a new application. Cancellation is not idempotent: a repeated request after success returns `INVOICE_CANNOT_CANCEL` because the application is already `canceled`.

Use the same `invoice.apply` client Token as the other application endpoints. “Current Client ID” means the OAuth `client_id` represented by that Token. Only records whose `sourceClientId` exactly matches it can be canceled; treat another application's record as not found. Do not retry a timed-out cancellation blindly; refresh the application's invoice list first.

The separate `/api/v1/me/invoices/<application_id>/cancel` route belongs to the logged-in user center and requires a user Access Token. It is not part of the application integration contract.

## Important Error Codes

| Code | Meaning | Client action |
| --- | --- | --- |
| `OAUTH_INVALID_CLIENT` / OAuth token error | Client credentials are invalid, app is not approved, or the requested scope is unavailable. | Fix server configuration or app approval; do not retry unchanged credentials. |
| `AUTH_INVALID_TOKEN` / `AUTH_INSUFFICIENT_SCOPE` / `AUTH_FORBIDDEN` | The client Token expired or was invalidated, lacks `invoice.apply`, or its application owner is no longer an active developer. | Obtain a new Token once; if the error remains, fix application approval, scope, owner status, or credentials. Treat the actual HTTP status and `code` as authoritative. |
| `INVOICE_INVALID_ORDERS` | Empty, oversized, or more than 20 order numbers. | Correct the input. |
| `INVOICE_DUPLICATE_ORDER` | The request repeats an order number. | Deduplicate before calling. |
| `INVOICE_ORDER_NOT_ELIGIBLE` | An order is missing, unpaid, invalid, or already claimed during validation. | Ask the user to select eligible orders; do not blindly retry. |
| `INVOICE_TAX_ORDER_REQUIRED` | Tax is requested without a tax order number. | Complete tax reconciliation and submit the exact paid `taxOrderNos`. |
| `INVOICE_INVALID_TAX_ORDERS` / `INVOICE_DUPLICATE_TAX_ORDER` | Tax order list is empty, malformed, too large, or contains duplicates. | Normalize and deduplicate the exact tax order numbers before retrying. |
| `INVOICE_TAX_PAYMENT_NOT_ELIGIBLE` | Tax order is unpaid, wrong amount/currency/purpose, or not a system tax order. | Return the user to payment or start a new validation flow. |
| `INVOICE_TAX_PAYMENT_EXCEEDS_REQUIRED` | Previously paid tax exceeds the amount required by the current source-order set. | Preserve the tax orders and restore the original source orders or restart the flow; do not discard a paid tax order. |
| `INVOICE_TAX_ORDER_ALREADY_CLAIMED` | A tax order was already attached to an invoice application. | Refresh application state and do not reuse that tax order. |
| `INVOICE_ORDER_ALREADY_CLAIMED` | A source order or tax order was claimed by another application. | Refresh application state and do not resubmit the same claim. |
| `INVOICE_INVALID_FORM` | Required buyer data is missing or a field fails the remaining format or length validation. | Correct the buyer data before resubmitting. |
| `INVOICE_CANNOT_CANCEL` | The application is not `pending` or `approved`. | Refresh the application and disable cancellation for its current state. |
| `NOT_FOUND` | The cancellation target does not exist or does not belong to the current Client ID. | Refresh the application's invoice list; do not reveal ownership information. |
| `INVOICE_ORDER_RATE_LIMITED` / `INVOICE_RATE_LIMITED` | Order validation or invoice submission exceeded its API limit. | Back off until the limit window resets; do not immediately retry the mutation. |
| `PAYMENT_NOT_CONFIGURED` / `PAYMENT_MERCHANT_NOT_CONFIGURED` / `INVOICE_ADMIN_EMAIL_NOT_CONFIGURED` | The server lacks required order-verification, checkout, or notification configuration. | Report a service configuration failure; do not fabricate a successful result. |
| `PAYMENT_UNAVAILABLE` / `PAYMENT_INVALID_RESPONSE` | The payment dependency is unavailable or returned an invalid order response. | Apply bounded transient-error handling and preserve `requestId`; after a submission timeout, check application state before retrying. |
| `INVOICE_TAX_STATUS_RATE_LIMITED` | Tax status polling exceeded the current Client ID's hourly limit. | Back off polling, then reconcile through order validation before submission. |
| `429` or transient `5xx` | Rate limit or temporary service failure. | Apply the target project's bounded retry policy and preserve request IDs. |

## End-to-End Sequence

1. Obtain and cache a `client_credentials` token with `invoice.apply`.
2. Validate the exact paid source orders.
3. If `needPayTax=true`, send the user to one selected `taxPayments.*.payUrl`, poll the client status endpoint, then revalidate with the paid `taxOrderNos` and require `taxDueAmount: "0.00"`.
4. Submit the same source orders and, when applicable, every unchanged paid `taxOrderNos` value.
5. Store the returned application ID and poll the application list from a backend job or user refresh.
6. Download the PDF only after status becomes `completed`.
7. Cancel a current-client `pending` or `approved` application when required, then refresh application and order eligibility state.
