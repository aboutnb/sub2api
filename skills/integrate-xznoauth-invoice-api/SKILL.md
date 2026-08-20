---
name: integrate-xznoauth-invoice-api
description: Guide developers and coding agents through integrating XZNOAuth's self-service invoice API, including tax-mode scoping, client_credentials token acquisition, paid-order validation, tax checkout and reconciliation, payment-status queries, invoice submission, application cancellation, status listing, and PDF download. Use when adding, generating, testing, or debugging an integration with `/api/v1/invoice-orders/validate`, `/api/v1/invoice-tax-payments/status`, `/api/v1/invoices/:id/cancel`, `/api/v1/invoices`, or the `invoice.apply` scope.
---

# Integrate XZNOAuth Self-Service Invoicing

Build a server-side integration that validates paid orders, optionally collects invoice tax, submits an invoice application, and retrieves the completed PDF. Keep credentials and access tokens on the server, make the base URL configurable, and treat the API response envelope as part of the contract.

Read [references/api-contract.md](references/api-contract.md) when implementing request models, response parsing, field validation, or error handling. Do not duplicate its detailed examples in generated project documentation unless the target project needs them.

## Workflow

1. Inspect the target project and identify its HTTP client, configuration system, secret storage, retry policy, and test conventions. Reuse those patterns instead of adding a second API client or configuration layer.
2. When the request does not already state the required coverage, ask the user to choose exactly one integration scope before changing code: only invoices that do not require invoice-tax payment, only invoices that require invoice-tax payment, or both modes. Mark the both-modes option as recommended, but do not silently select it. Use these explicit terms instead of the ambiguous shorthand "tax-inclusive/tax-exclusive." Treat this as an integration-scope decision; when both modes are selected, choose the branch per invoice request through `needPayTax`, not through a deployment-wide setting.
3. Require an approved XZNOAuth developer application with the `invoice.apply` scope. Store `client_id`, `client_secret`, and the configurable `XZNOAUTH_BASE_URL` as server-side configuration. Use `https://oauth.xzncraft.cn` only as the production default documented by this skill; never hardcode credentials or put the secret in browser code.
4. Implement a small token client for `POST /api/oauth/token` using `application/x-www-form-urlencoded` and `grant_type=client_credentials`. Treat `expires_in` as dynamic, cache the token until shortly before it expires, and do not expect a Refresh Token. On an authentication failure, obtain a new token once and retry the original request only when the target client's conventions permit it. Application updates, approval changes, and secret rotation can invalidate an unexpired client Token. Redact the secret, token, and Authorization header from logs.
5. Implement `POST /api/v1/invoice-orders/validate` with one to twenty order numbers. Accept platform order numbers, external order numbers, and WeChat/Alipay transaction IDs. For an invoice that does not require invoice-tax payment, send `needPayTax: false` or omit it. When the selected scope includes invoices that require invoice-tax payment, send `needPayTax: true`, display a `taxPayments` map containing Alipay and WeChat checkout options, and persist only the tax order number for the channel the user pays. Preserve any already paid tax orders in `taxOrderNos` when revalidating.
6. When the selected scope includes invoices that require invoice-tax payment, do not submit until payment is confirmed. Poll `POST /api/v1/invoice-tax-payments/status` with the `invoice.apply` client Token and one `taxOrderNo`; proceed only when it returns `{ "paid": true }`. Then revalidate the original source orders with `needPayTax: true` and every paid `taxOrderNos`, and require `taxDueAmount` to be `"0.00"`. Do not infer payment from a checkout redirect, client-side state, or a successful order-creation response.
7. Implement `POST /api/v1/invoices` with the exact order list and buyer data used by the validated flow. When invoice-tax payment is required, send `needPayTax: true` and every verified, paid `taxOrderNos`; when it is not required, omit both tax-order fields. `taxOrderNo` remains a single-order compatibility alias only. Treat a `201` response with `code: 0` as a successful pending application.
8. Implement `GET /api/v1/invoices?page=&pageSize=` for application status and `GET /api/v1/invoices/:id/pdf` for completed documents. Parse the list as JSON; stream the PDF response as binary and only offer it when the application status is `completed`. The platform also sends the completed PDF directly to `recipientEmail` as an attachment, so do not require the recipient to sign in just to receive the invoice.
9. Implement `POST /api/v1/invoices/:id/cancel` for an application-owned `pending` or `approved` invoice. On success, treat the returned application as `canceled` and refresh order eligibility. Treat another Client ID's application as not found, and refresh the application list before retrying after a timeout.
10. Add focused tests for token form encoding and redaction, envelope parsing, unpaid or already-claimed order rejection, application status handling, application cancellation success and rejection, and PDF binary download. If the selected scope includes invoice-tax payment, also test both checkout channels, payment-status polling, tax reconciliation, payment-required submission, and duplicate tax-order rejection. If both modes are selected, test both `needPayTax` branches. Use mocked HTTP responses; do not call production payment services from tests.

## Guardrails

- Preserve the exact order set between validation and submission. A validation response is not a reservation; another application can claim an order before submission.
- Let the server calculate tax from its configured rate. Display `taxAmount`, `taxPaidAmount`, `taxDueAmount`, and the selected checkout `payUrl` from the validation response; do not calculate a second amount in a way that could drift by a cent.
- Treat `INVOICE_TAX_PAYMENT_NOT_ELIGIBLE`, `INVOICE_TAX_PAYMENT_EXCEEDS_REQUIRED`, `INVOICE_TAX_ORDER_ALREADY_CLAIMED`, `INVOICE_ORDER_NOT_ELIGIBLE`, and `INVOICE_ORDER_ALREADY_CLAIMED` as user-action or conflict outcomes, not retryable transport failures. Never silently discard a paid tax order to bypass an overpayment error.
- Retry only according to the target project's existing policy for timeouts, `429`, and transient `5xx` responses. Do not blindly retry a timed-out submission without checking the resulting application state.
- Preserve `requestId` from error envelopes in application logs and support responses, but redact order-sensitive credentials and tokens. Do not expose raw payment-provider responses to end users.
- Use the client integration routes documented here, not the equivalent `/api/v1/me/...` user-center routes, when authenticating with `client_credentials`.

## Delivery Checklist

- [ ] Base URL, Client ID, Client Secret, token cache, and timeout are configurable; the cache follows returned `expires_in` and does not depend on a Refresh Token.
- [ ] The user explicitly selected only invoices requiring invoice-tax payment, only invoices not requiring it, or both modes; both-mode integrations branch per request with `needPayTax`.
- [ ] The application has been approved with `invoice.apply` and the token request uses that scope.
- [ ] Paid-order validation is completed before submission and matches the selected tax-mode coverage.
- [ ] When invoice-tax payment is supported, checkout displays both channels, records the user-selected paid tax order, and reconciles `taxOrderNos` until `taxDueAmount` is `0.00`.
- [ ] When invoice-tax payment is supported, payment status is polled through the client endpoint with the `invoice.apply` Token before final reconciliation.
- [ ] Submission reuses the validated order list and handles validation, overpaid tax, conflict, rate-limit, and payment errors distinctly.
- [ ] Application listing and completed-PDF download are implemented with authorization and ownership checks delegated to the API.
- [ ] Application cancellation handles `pending`, `approved`, `canceled`, not-found, and conflict outcomes without retrying the mutation blindly.
- [ ] Secrets and tokens are absent from source control, browser bundles, URLs, logs, and error messages.
- [ ] Unit or integration tests cover both successful and rejection paths.
