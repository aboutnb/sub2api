# Provider Pricing API PRD

## Objective

Expose public group prices in CNY to Hvoy AI through `GET /api/provider/pricing`.
The endpoint is public, uses the Provider Pricing API v1.1 response contract, and is refreshed at most once every 30 minutes per process.

## Business Rules

- Currency is `CNY`.
- Official USD price numbers map 1:1 to the site's CNY balance units. No currency exchange multiplier is applied.
- Token rates prefer the system's official model catalog (dynamic prices, then bundled fallback prices). Models without an official catalog rate fall back to configured group/channel prices.
- Token prices in CNY/1M tokens are `official USD/token rate * 1,000,000 * group.RateMultiplier`.
- Example: official output price `30 USD/1M tokens` with group rate `0.25` is exported as `7.5 CNY/1M tokens`.
- Per-call prices are the configured base price multiplied by the group rate. Image prices use the independent image rate when enabled, matching the site's image billing rule.
- Group multipliers are applied exactly once. User-specific and peak multipliers are not applied to this public standard quote.
- Token prices are rounded to one decimal place; per-call prices are rounded to two decimal places so low-cost image prices such as `0.06` remain representable. Required prices that round to zero are omitted, as the Hvoy contract rejects zero prices.
- All public (non-exclusive) groups and their configured model entries are enumerated, including inactive entries. `enabled` is true only when the group and a supporting channel are active.
- `group_name` is the current public group name. `model_name` is the stable model ID within that group.
- Each `group_name + model_name` pair appears at most once; an enabled candidate wins when several platforms offer the same model ID.
- Image models use per-call pricing; the group's 2K default image price is preferred, then 1K/4K, then the model's configured per-call price.
- Video pricing is intentionally omitted.

## Contract

- `schema_version` is `1.1`.
- `data.price_unit` is `per_1m_tokens`.
- Per-call entries override the unit with `price_unit: "per_call"` and `unit_price`.
- `updated_at` is the UTC time at which the cached snapshot was generated.
- HMAC authentication is disabled.

## Acceptance Criteria

- The unauthenticated endpoint returns a valid v1.1 JSON document.
- A second request within 30 minutes does not rebuild the database snapshot.
- Token, cache, image, and per-call prices follow the 1:1 CNY balance, group-rate, and rounding rules.
- A catalog output price of 30 and group rate of 0.25 yields 7.5, even when group/channel token price overrides exist.
- Inactive entries remain in the response with `enabled: false` when they have a valid price.
- Invalid or price-less entries are not emitted because the Hvoy contract would discard them.
