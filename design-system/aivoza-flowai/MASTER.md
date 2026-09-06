# Aivoza FlowAI UI system

This file is the implementation source of truth for the `sub2api-flowai-theme` branch. The visual reference is `aivoza-home-pixel.html`; product usability and existing business behavior take precedence over decorative styling.

## Direction

- Warm, friendly SaaS product interface with restrained pixel accents.
- Pixel character comes from 2px outlines, square details, a subtle 32px grid, and 2–5px offset shadows.
- Do not use oversized poster typography, arcade fonts, noisy tiled backgrounds, or heavy animation.
- Keep dense admin and account workflows readable. Styling must not reduce information capacity or alter behavior.
- Always use the configured site logo through `resolveBrandLogo`; never replace it with a decorative fake logo.

## Core tokens

| Role | Light | Dark | Notes |
| --- | --- | --- | --- |
| Page | `#fffaf4` | `#0d1422` | Warm paper / blue-black |
| Surface | `#ffffff` | `#172033` | Cards, menus, dialogs |
| Ink | `#172033` | `#f7f5f2` | Main copy and outlines |
| Muted ink | `#566074` | `#a6a8ad` | Supporting text |
| Line | `#eadfd4` | `#344056` | Default borders |
| Control line | `#8f7968` | `#6c7d9b` | Input/button boundaries, at least 3:1 against surfaces |
| Focus | `#087c74` | `#4acbbb` | Solid 3px keyboard outline, at least 3:1 |
| Teal | `#2fb9aa` | `#4acbbb` | Decorative accent |
| Teal solid | `#087c74` | `#2fb9aa` | White-text controls and links |
| Teal soft | `#dff8f1` | teal at 16% | Selected and information states |
| Coral | `#ff8a5c` | `#ff9c72` | Primary CTA, paired with ink text |
| Coral soft | `#fff0e8` | coral at 14% | Warm highlights |
| Yellow | `#f7c95c` | `#f7c95c` | Sparse status detail |

The matching CSS variables live in `frontend/src/style.css`. They are split into
primitive values, mode-aware semantic roles, and component contracts. Product UI
consumes Tailwind's `canvas`, `surface`, `ink`, `line`, `brand`, and `action`
roles. `primary` and `accent` remain compatibility names for existing Aivoza
controls while they are migrated.

Tailwind's stock `gray`, `blue`, `teal`, and other palettes are not globally
overridden. They are available only when a page intentionally needs a neutral,
data-series, provider-brand, or semantic status color. Likewise, ordinary
`shadow-sm` through `shadow-2xl` retain Tailwind's soft elevation; the named
`shadow-pixel*` utilities are reserved for brand marks, primary actions, and
selected states.

## Typography

- Sans stack: `Avenir Next`, `Trebuchet MS`, system UI, PingFang SC, Microsoft YaHei.
- Body: 14–16px. Supporting labels: 12–14px.
- Page title: 20px mobile, 24px desktop.
- Marketing hero: maximum 36px mobile / 40px desktop unless a page-specific review approves otherwise.
- Prefer weight and spacing over size. Avoid all-caps paragraphs and decorative monospace body text.

## Components

### Buttons

- Default minimum height: 44px; compact table actions may use 36px.
- 2px ink outline, 10–14px radius, 3px offset shadow.
- Primary CTA: coral fill with ink text.
- Navigation/selected controls: teal solid with white text, or teal soft with dark teal text.
- Hover movement is at most 1px; active state compresses the offset shadow.

### Cards and dialogs

- 2px line or ink-adjacent border, 14–18px radius.
- Default card and overlay shadows use soft elevation without a pixel offset.
- Do not make every card interactive. Only actionable cards receive lift on hover.
- Dialogs and dropdowns use stronger soft elevation and visible focus containment.

### Inputs

- Minimum 44px height, 2px border, 10–14px radius.
- Control boundaries and the solid focus outline retain at least 3:1 contrast against their surface.
- Focus uses a teal border plus an external outline; never depend on color alone for errors.
- Labels remain visible. Placeholders do not replace labels.

### Tables

- Preserve density and sticky-column behavior.
- Header uses a very light teal/paper tint, bold 12px labels, and a clear bottom rule.
- Sort controls must be keyboard reachable and retain `aria-sort` semantics.
- Mobile card rendering must not remove horizontal access to desktop tables where still used.

## Layout

- Supported checks: 375px, 768px, 1024px, 1440px.
- App header remains 64px; desktop sidebar remains 256px / 72px collapsed.
- Main content width is capped at 1680px and remains fluid below it.
- Background decoration must be pointer-free and never reduce text contrast.

## Motion and accessibility

- Transitions: 150–220ms for hover/focus; up to 300ms for overlays and sidebar changes.
- Respect `prefers-reduced-motion`; non-essential transitions and animations collapse to 1ms.
- Normal text contrast target is at least 4.5:1. Do not pair light teal with white text.
- Control boundaries and keyboard focus indicators target at least 3:1 against adjacent surfaces.
- Every interactive element needs a visible keyboard focus state and semantic HTML.
- Use the existing SVG icon system. Do not use emojis as structural icons.

## Fixed homepage integration

- `aivoza-home-pixel.html` is the single reviewed homepage source and is imported by `HomeView` with Vite `?raw` at build time. It is no longer configured from the admin panel.
- Its logo source must remain the sole quoted `{{SITE_LOGO}}` token. `HomeView` replaces that token with the escaped, validated `site_logo` value from public settings before rendering the static asset.
- The template must remain dependency-free and contain no `script`, inline event attributes, `iframe`, or `javascript:` URLs. Its structure, copy, links, spacing, and pixel details are release-controlled and are not rebuilt in Vue.
- Light/dark behavior is limited to the `.dark .aivoza-home` variable override. It follows the shared theme state and has no homepage-only theme control.
- `home_content` and `compact_home_enabled` remain deprecated compatibility fields. Admin reads and ordinary saves preserve their historical database values; public settings return `""` and `false`, and neither value participates in rendering or CSP generation.
- Both repository Dockerfiles must copy `aivoza-home-pixel.html` to `/app/aivoza-home-pixel.html` before `pnpm run build`, matching `HomeView`'s relative raw-import path.

## Deliberate exceptions

- Provider-owned payment controls may retain official Stripe, Airwallex, Alipay, and WeChat Pay colors.
- Semantic success, warning, and danger colors remain distinct from brand accents.
- A gradient is allowed only when it encodes data, such as the two-part latency bar in `UsageTable`; decorative surface and headline gradients are not part of this system.

## Review checklist

- [ ] No headline reads like an oversized poster.
- [ ] Configured official logo is visible in public, auth, and app shells.
- [ ] Light and dark modes retain contrast and hierarchy.
- [ ] Keyboard focus, table sorting, dialogs, and dropdowns remain usable.
- [ ] No mobile horizontal page overflow at 375px.
- [ ] Reduced motion is honored.
- [ ] Existing business logic, routes, permissions, and settings behavior are unchanged.
