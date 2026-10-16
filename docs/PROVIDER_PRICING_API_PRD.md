# Provider Pricing API PRD

## Objective

Expose the gateway's standard group prices to Hvoy AI through `GET /api/provider/pricing`.
The endpoint is public, uses the Provider Pricing API v1.1 response contract, and is refreshed at most once every 30 minutes per process.

## Business Rules

- Currency is `CNY`.
- The conversion rate is fixed at `7.5 CNY / USD`.
- Token prices are converted from USD/token to CNY/1M tokens.
- Per-call prices are converted from USD/call to CNY/call.
- Every numeric price is rounded to one decimal place.
- Prices use the group's standard configured price only. Group, user-specific, and peak multipliers are not applied.
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
- Token, cache, image, and per-call prices follow the conversion and rounding rules.
- Inactive entries remain in the response with `enabled: false` when they have a valid price.
- Invalid or price-less entries are not emitted because the Hvoy contract would discard them.
