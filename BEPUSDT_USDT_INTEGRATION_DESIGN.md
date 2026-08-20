# BEpusdt USDT Multi-network Integration Design

## 1. Decision

Sub2API should integrate BEpusdt as a dedicated `bepusdt` payment provider. The
user-facing method is `usdt`; the user selects a network before the upstream
order is created. Sub2API then calls BEpusdt's fixed-network transaction API.

Do not integrate BEpusdt through the existing EasyPay provider. BEpusdt only
implements the EasyPay redirect endpoint, while Sub2API also depends on the
EasyPay query endpoint for missed-callback recovery. The payment type also must
be shown as USDT rather than Alipay or WeChat Pay.

The initial production network list is the Sub2API `usdt_payment.enabled_networks`
allowlist (YAML or `USDT_PAYMENT_ENABLED_NETWORKS`). The code can support all
BEpusdt USDT networks, but a network is exposed only when:

1. It is enabled in the Sub2API provider instance.
2. BEpusdt reports an enabled receiving wallet for it.
3. Its RPC, scanner, queue, and chain-head lag are healthy.
4. The network has passed the production test matrix in section 11.

BEpusdt currently supports these USDT trade types:

| Network | BEpusdt trade type | Initial recommendation |
| --- | --- | --- |
| TRON | `usdt.trc20` | Phase 1 |
| BSC | `usdt.bep20` | Phase 1 |
| Ethereum | `usdt.erc20` | Phase 1 after fee/latency UX validation |
| Polygon | `usdt.polygon` | Phase 2 |
| Arbitrum | `usdt.arbitrum` | Phase 2 |
| Solana | `usdt.solana` | Phase 2 |
| TON | `usdt.ton` | Phase 2 |
| Aptos | `usdt.aptos` | Phase 2 |
| X Layer | `usdt.xlayer` | Phase 2 |
| Plasma | `usdt.plasma` | Phase 2 |

BEpusdt currently has no USDT-Base trade type. Base is USDC-only and therefore
is outside this integration scope.

## 2. Why Network Selection Belongs in Sub2API

The recommended flow is:

1. Sub2API shows one `USDT` payment method.
2. The user selects one enabled network, for example TRC20.
3. Sub2API creates its local payment order and freezes the final fiat amount.
4. The dedicated provider calls `POST /api/v1/merchant/order/create` with a
   fixed trade type such as `usdt.trc20`.
5. BEpusdt returns the exact USDT amount, address, exchange rate, expiry, and
   upstream `trade_id`.
6. Sub2API persists that quote before returning the checkout response.
7. The network cannot be changed for that order. Selecting another network
   creates a new order.

This is safer than relying on BEpusdt's hosted multi-network selector. The
hosted selector binds a chosen order to a browser fingerprint, while the
current `/api/v1/pay/info` query is also fingerprint-aware and unsigned. A
Sub2API background worker can therefore fail to query an order selected by a
different browser. Fixed-network creation avoids that functional dependency,
although a signed merchant query API is still required for security.

## 3. End-to-end Flow

```text
User              Sub2API                BEpusdt                Chain
 | select USDT/TRC20 |                       |                     |
 |------------------>|                       |                     |
 |                   | create local PENDING  |                     |
 |                   | create-transaction -->|                     |
 |                   |<-- frozen quote -------|                     |
 |<-- amount/address/payment URL -------------|                     |
 |                                                               send
 |----------------------------------------------------------------->|
 |                   |                       |<-- block/receipt -----|
 |                   |<-- signed webhook ----|                     |
 |                   | verify + durable inbox|                     |
 |                   | exact reconciliation  |                     |
 |                   | idempotent fulfillment|                     |
 |<-- COMPLETED ------|                       |                     |

Missed webhook:
Sub2API reconciler -- signed query --> BEpusdt -- persisted paid state --> fulfill
```

Frontend polling only improves perceived responsiveness. It is not a payment
reliability mechanism because the browser may be closed.

## 4. Immutable Amount Contract

The integration must use decimal strings or integer minor units for financial
comparisons. Do not compare the critical fields with binary floating-point
tolerances.

For every order, freeze this tuple:

```text
merchant_order_id = Sub2API out_trade_no
fiat_currency     = configured payment currency, initially CNY
fiat_amount       = Sub2API pay_amount, exact minor units
crypto_currency   = USDT
network           = selected network
trade_type        = BEpusdt trade type
crypto_amount     = exact actual_amount returned by BEpusdt
receiving_address = exact address returned by BEpusdt
exchange_rate     = exact rate returned by BEpusdt
upstream_trade_id = BEpusdt trade_id
upstream_expiry   = BEpusdt expiry
```

BEpusdt may increase `crypto_amount` by one configured USDT atom when another
pending order already uses the same address and amount. The returned amount is
therefore authoritative. Sub2API must persist it and must not recalculate it.

On a paid callback or query response, all of these checks are mandatory:

1. Callback signature and timestamp are valid.
2. `order_id` equals the Sub2API `out_trade_no`.
3. `trade_id` equals the frozen upstream trade ID.
4. Fiat amount equals `pay_amount` in exact fiat minor units.
5. `actual_amount` equals the frozen USDT decimal string after canonical decimal
   normalization.
6. Trade type, network, receiving address, and fiat currency equal the quote.
7. The transaction hash is present and has not fulfilled another order.
8. The chain transfer timestamp is inside the upstream order's payment window.

A mismatch must keep the order unfulfilled, write a payment audit record, and
raise an operator alert. The amount check must never be relaxed to make a
callback pass.

### Site Fee and Balance Credit

Sub2API calculates its fee before calling BEpusdt. The exact final
`pay_amount` is sent as BEpusdt's fiat `amount`.

For a 100 balance recharge with a 2 percent site fee:

```text
base amount       = 100.00
site fee          = 2.00
BEpusdt fiat amount = 102.00
```

With `BALANCE_RECHARGE_MULTIPLIER=1` and fee credit enabled, the credited
balance is 102.00. With fee credit disabled, it is 100.00. The USDT amount is
only the chain settlement amount and must never be used directly as the site
balance credit.

Subscription fee participation continues to use the existing
`SUBSCRIPTION_FEE_ENABLED` setting.

## 5. Required BEpusdt Fork Changes

The current upstream behavior is not sufficient for a production payment SLO.
Implement these changes on `codex/local-sub2api-integration`:

### Merchant APIs

Add signed server-to-server endpoints that do not depend on browser
fingerprints:

```text
POST /api/v1/merchant/order/query
POST /api/v1/merchant/capabilities
GET  /api/v1/merchant/readiness
POST /api/v1/merchant/rate
```

The query response must include `order_id`, `trade_id`, status, fiat currency,
fiat amount, exact USDT amount, rate, trade type, network, receiving address,
transaction hash, chain transfer timestamp, creation time, expiry, and
confirmation information.

The capabilities response must return only networks with enabled wallets and
must include network readiness. Sub2API intersects it with its own allowlist.
The rate response returns the current `USDT/CNY` quote. Sub2API reads it before
each order, converts the user-entered USDT amount to the provider fiat amount,
and sends the same rate back when creating the order so the crypto quantity is
frozen instead of being recomputed in the browser.

### Authentication and Idempotency

Add an HMAC-SHA256 protocol version with request timestamp, nonce, key ID, and
body digest. Reject stale timestamps and replayed nonces. Keep the current MD5
scheme only for backward compatibility.

Repeated create requests with the same `order_id` and identical immutable
parameters must return the same order and quote. A request with the same
`order_id` but different amount, network, currency, or callback target must be
rejected. It must not silently rebuild a waiting payment order.

Restrict `notify_url` and `redirect_url` to configured host allowlists. The
current validation checks only the URL scheme and host.

### Webhook Contract

Add a versioned callback payload using strings for both fiat and crypto amounts:

```json
{
  "event_id": "unique-delivery-event-id",
  "event_type": "payment.succeeded",
  "occurred_at": 0,
  "order_id": "sub2api-out-trade-no",
  "trade_id": "bepusdt-trade-id",
  "fiat": "CNY",
  "amount": "102.00",
  "crypto": "USDT",
  "actual_amount": "14.17",
  "trade_type": "usdt.trc20",
  "network": "tron",
  "token": "receiving-address",
  "block_transaction_id": "chain-transaction-hash",
  "transfer_at": 0,
  "signature_version": "v2",
  "signature": "hmac-sha256"
}
```

The delivery is successful only when Sub2API returns HTTP 200 with the exact
body `success`. The current native callback treats any HTTP 200 as success.

Use fast durable retries after an initial failure: 1s, 2s, 5s, 10s, 20s, 30s,
then 1m, 2m, 5m, 10m, and 30m. Persist `next_attempt_at`, attempt count, last
status, last error, and response summary. Do not derive retries only from the
confirmation timestamp.

### Scanner Reliability

Confirmation thresholds must be configurable per network. A single global
boolean with hard-coded offsets is not enough for production operations.

Support at least two RPC endpoints per enabled network with health scoring and
automatic failover. New orders must be disabled for a network when:

- the chain head has not advanced within its threshold;
- scanner lag exceeds the network threshold;
- the scan queue approaches its safety limit;
- all RPC endpoints are unhealthy;
- there is no enabled receiving wallet;
- the callback backlog exceeds its threshold.

Expose chain head, last scanned height, lag, queue depth, last successful scan,
RPC health, confirming order count, and callback backlog through readiness and
metrics endpoints. The existing root-page health check proves only that HTTP is
running; it does not prove payment readiness.

## 6. Required Sub2API Changes

### Provider and Configuration

Add `payment.TypeBepusdt = "bepusdt"` and user-facing type `usdt`. Implement a
dedicated provider using BEpusdt native JSON APIs.

Provider instance configuration:

| Field | Purpose |
| --- | --- |
| `apiBase` | Private/server-to-server BEpusdt API base |
| `publicBaseUrl` | Public checkout origin |
| `keyId` | HMAC key identifier |
| `apiSecret` | HMAC secret, sensitive |
| `fiat` | Initially `CNY` |
| `enabledNetworks` | Admin network allowlist |
| `orderTimeoutSeconds` | 180 to 3600, aligned with Sub2API |
| `latePaymentWindowMinutes` | Must not exceed BEpusdt lookback |

The provider must not implement automatic refunds. Blockchain refunds require
manual operator review and a separately authorized outbound wallet process.

Do not expose user cancellation after a crypto quote has been shown. A user may
have broadcast a transaction that is not yet visible; cancelling the upstream
order would stop BEpusdt from matching that transfer. Let the order expire and
retain late-confirmation recovery instead.

### Exact Quote Storage

Add a one-to-one crypto quote record for each payment order. Critical monetary
values remain strings:

```text
payment_order_id        unique
provider_trade_id       unique
fiat_currency
fiat_amount
crypto_currency
network
trade_type
crypto_amount
exchange_rate
receiving_address
upstream_created_at
upstream_expires_at
provider_status
transaction_hash        nullable, unique when present
chain_transfer_at        nullable
last_reconciled_at       nullable
```

Persist the provider trade ID and complete quote atomically before returning the
checkout response. If the create request times out, retry with the same
`out_trade_no`; BEpusdt idempotency must return the same quote.

### Durable Webhook Inbox

The isolated USDT webhook endpoint is:

```text
POST /api/v1/usdt/webhook/bepusdt
```

After signature verification, insert the event into a durable inbox using
`event_id` and payload hash as idempotency keys. Acknowledge only after the event
is durable. Attempt fulfillment immediately, while retaining the inbox item for
background retry if fulfillment fails.

The existing payment state transition and fulfillment lease remain the final
idempotency boundary. Webhook, active query, and background reconciliation may
race safely but may credit only once.

### Durable Reconciliation

Add a BEpusdt reconciliation worker independent of the browser and independent
of the current 60-second expiry sweep. It must claim due work durably so multiple
Sub2API instances do not query the same order concurrently.

Recommended schedule:

| Order age/state | Query interval |
| --- | --- |
| First 2 minutes, pending | 2 seconds |
| Until upstream expiry | 5 seconds |
| Expired but inside late-payment window | 15 seconds |
| Provider/RPC error | bounded exponential backoff, maximum 30 seconds while order is payable |

Store next attempt, attempt count, last query result, and last error. Process in
bounded batches with `SKIP LOCKED` or an equivalent durable lease.

The current Sub2API background job runs every 60 seconds and reconciles only
WeChat Pay orders. It does not protect BEpusdt orders. The result page polls
every 2 seconds for only 15 attempts, so it also cannot be used as a reliability
guarantee.

## 7. Expiry and Late Confirmation

BEpusdt matches a transfer only when its chain timestamp is strictly inside the
order creation and expiry window. It can discover that transfer later through
its lookback scanner. The default lookback is three hours.

Sub2API currently accepts an expired order for only a five-minute grace period.
That is too short for an RPC outage, lookback recovery, or a confirmation policy
that waits for multiple blocks.

For BEpusdt only, accept a later success when all conditions hold:

1. The signed query or callback proves the chain transfer timestamp was before
   the frozen upstream expiry.
2. All immutable quote fields match.
3. Current time is inside the configured late-payment window.
4. The late-payment window is no longer than BEpusdt's scanner lookback, with a
   small operational buffer handled before the upstream boundary.

After that window, do not auto-credit. Create a manual reconciliation incident
containing the transaction hash and full comparison result.

## 8. Real-time SLO

No blockchain system can guarantee a fixed wall-clock completion time from the
moment a wallet submits a transaction. Inclusion time, network finality, and
third-party RPC availability are outside Sub2API's control. The enforceable
guarantee is an SLO measured from chain eligibility and BEpusdt state changes.

```text
total latency = chain inclusion
              + configured confirmation window
              + scanner detection
              + callback/query delivery
              + Sub2API fulfillment
```

Production targets under healthy RPC and without an added confirmation window:

| Segment | Target |
| --- | --- |
| BEpusdt success -> Sub2API durable receipt | p95 <= 3s, p99 <= 8s |
| Durable receipt -> completed balance/subscription | p95 <= 2s |
| Missed callback -> background recovery | p95 <= 5s while pending |
| TRON eligible block -> BEpusdt detected | p95 <= 10s |
| Ethereum eligible block -> BEpusdt detected | p95 <= 20s |
| Other enabled networks | set only after network-specific load testing |

When block-offset confirmation is enabled, add the actual network confirmation
time. For example, BEpusdt currently hard-codes 30 TRON blocks, 12 Ethereum
blocks, 15 BSC blocks, and 40 blocks for several other EVM networks.

The system preserves the SLO by failing closed: if readiness is degraded, the
network is removed from new checkout choices before a user is given an address.
Existing orders continue scanning, querying, and alerting until their recovery
window closes.

## 9. Failure Handling

| Failure | Required behavior |
| --- | --- |
| Create response times out | Retry same idempotency key; return same quote |
| User closes page | Webhook and background reconciler continue |
| Webhook is lost | Signed query recovers within polling SLO |
| Sub2API webhook returns 500 | BEpusdt fast durable retry; inbox remains absent until accepted |
| Fulfillment fails after webhook | Durable inbox and fulfillment lease retry; no duplicate credit |
| BEpusdt restarts | Persistent DB, scanner lookback, callback queue, and quote survive |
| Primary RPC fails | Switch to healthy secondary; disable new orders if all fail |
| Scanner lags | Disable affected network, alert, keep recovery scan running |
| Amount/network/address mismatch | Do not credit; audit and alert |
| Payment sent just before expiry | Accept after confirmation if chain timestamp is in-window |
| Payment sent after expiry | Do not auto-credit; manual incident |
| Wrong network used | Do not auto-credit; manual wallet recovery process |
| Duplicate callback/query result | Deduplicate event and transaction hash; fulfillment CAS credits once |
| BEpusdt database unavailable | Disable all new USDT orders; existing recovery raises a critical alert |

MQTT may be used as a low-latency wake-up signal and monitoring feed. It is not
a source of truth because it is emitted before final order matching and may be
duplicated or unavailable during a subscriber outage.

## 10. Deployment Requirements

- Use PostgreSQL for production BEpusdt state, not the local SQLite setup.
- Run one active scanner unless BEpusdt gains explicit distributed task leases.
- Keep a hot standby and tested restore procedure rather than two uncontrolled
  active scanners.
- Put Sub2API and the BEpusdt API on a private network; expose only the checkout
  and required webhook routes through HTTPS.
- Configure BEpusdt `api_app_uri` or rewrite returned checkout URLs to the
  configured public origin. Never return a Docker hostname to the browser.
- Store HMAC secrets in the existing sensitive provider configuration path and
  never return them through the admin API.
- Use paid, rate-limited RPC accounts with independent providers for primary and
  secondary endpoints. Public default RPCs are not a production reliability
  plan.
- Back up BEpusdt and Sub2API payment tables and test point-in-time recovery.
- Synchronize hosts with NTP because signatures, expiry, and replay windows are
  time-dependent.

## 11. Production Test Gate

Each enabled network must pass all tests with small real transfers before it is
shown to users:

1. Exact amount, collision-adjusted amount, and consecutive same-amount orders.
2. Balance recharge with no fee, fee not credited, and fee credited.
3. Subscription with its fee switch off and on.
4. Browser closed immediately after displaying the address.
5. Callback blocked, callback returns 500, callback times out, and callback is
   delivered more than once.
6. Sub2API restarts before callback, during fulfillment, and after paid state but
   before completion.
7. BEpusdt restarts before detection, while confirming, and before callback.
8. Primary RPC is cut while the transaction is pending; secondary failover must
   detect it without manual intervention.
9. Transfer just before expiry, confirmation after expiry, and recovery through
   lookback.
10. Transfer after expiry, underpayment, overpayment, wrong network, wrong token,
    wrong address, and replayed signed payload.
11. Two Sub2API instances reconcile the same order concurrently.
12. Transaction hash reuse is rejected and balance/subscription is credited once.

Release gates:

- 100 percent pass on the failure matrix for every enabled network.
- No unexplained amount mismatch in a minimum 72-hour soak test.
- Scanner lag, RPC failover, callback backlog, and reconciliation backlog alerts
  are visible and tested.
- The p95 and p99 targets in section 8 are met under expected concurrent order
  volume and injected callback/RPC failures.
- A manual transaction-hash reconciliation procedure is documented and tested.

## 12. Current Local Baseline

The local BEpusdt container is healthy at `http://127.0.0.1:18080`, but its
health check only loads the root page. Its local configuration currently uses:

- exact/classic amount matching;
- USDT atom `0.01`;
- 20-minute upstream timeout;
- three-hour lookback;
- block-offset confirmation disabled;
- no configured receiving wallets.

This local instance is suitable for API and failure-injection development after
test wallets are added. It is not yet a payment-ready production baseline.
