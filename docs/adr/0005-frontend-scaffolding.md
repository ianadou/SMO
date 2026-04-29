# ADR 0005 — Frontend scaffolding decisions

**Status:** Accepted (2026-04-29)

## Context

The SMO frontend has not been scaffolded yet. Before running `nuxi init`,
several technical choices need to be locked so the design system can be
ported to production cleanly:

- typography (which font, which weights, which numerics font)
- font hosting (CDN vs self-host)
- CSS framework
- icon library
- package manager
- color mode strategy

The visual direction is already locked by the `SMO Design System/`
folder at the repo root: dark-only, restrained palette, sober animations,
French + tutoiement copy, mobile-first 320px minimum. This ADR captures
the technical decisions that flow from those visual choices.

## Decision

### Stack

| Concern | Choice |
|---|---|
| Framework | Nuxt 3 (SPA mode, no SSR) — already locked in CLAUDE.md |
| Language | TypeScript |
| CSS | Tailwind CSS |
| Display & UI font | Inter Variable (weights 100–900) |
| Numerics font | JetBrains Mono (tabular figures for scores, jersey numbers, timers) |
| Font hosting | Self-host via the official `@nuxt/fonts` module |
| Icon library | `lucide-vue-next` |
| Package manager | pnpm |
| Color mode | Dark only at MVP — no light mode toggle |
| Minimum viewport | 320px (mobile-first) |

### Self-hosted fonts via `@nuxt/fonts`, not Google Fonts CDN

The design system mock in `SMO Design System/colors_and_type.css`
loads Inter and JetBrains Mono via `@import url('https://fonts.googleapis.com/...')`.
That convenience is acceptable for the static HTML mocks Claude Design
generated — it is **not** acceptable for production.

`@nuxt/fonts` downloads the requested families at build time, ships
them inside the bundle, generates `font-display: swap`, and emits the
right `preload` tags. The CDN `@import` line in the design system file
stays where it is (the file is the design reference, not production
CSS) — only the design tokens (colors, sizes, spacing) get lifted into
`tailwind.config.ts`.

**Why self-host rather than CDN:**

- **GDPR risk on `.fr`**: The CNIL and German court rulings consider
  Google Fonts CDN requests as transferring user IPs to Google without
  explicit consent — non-compliant for a French-targeted product like
  sportpotes.fr.
- **One fewer DNS lookup** on first paint, especially on mobile.
- **No third-party runtime dependency**: a Google Fonts outage or a
  network policy block on the user side cannot break the typography.
- **Bundle stays self-contained**: easier to audit and reproduce.

The bundle weight cost is ~80–120 KB for the two families with the
weights we use (400, 500, 600, 700) — acceptable for a portfolio app
where typography is part of the brand.

### Tailwind tokens lifted from CSS variables

The design system ships its tokens as CSS custom properties in
`SMO Design System/colors_and_type.css`. At scaffold time these get
mirrored into `tailwind.config.ts` under `theme.extend.colors`,
`fontFamily`, `borderRadius`, `boxShadow`. The CSS file remains as
the human-readable reference; production styles use Tailwind classes
that resolve to the same hex values.

### JSX kits → Vue 3 SFC at integration

The seven UI kits in `SMO Design System/ui_kits/` are written in JSX
because Claude Design iterates faster in React. The markup, class
names, and CSS files map 1:1 to Vue 3 SFCs. Conversion happens
incrementally as each page is built, not as a single up-front
rewrite — that keeps the design system reference synced with what is
shipped.

### pnpm rather than npm/yarn/bun

- pnpm is what the Nuxt team itself recommends; the ecosystem is
  fully compatible.
- Strict dependency resolution — no surprise hoisting bugs.
- Disk-efficient via global content-addressed store.
- Renovate handles `pnpm-lock.yaml` natively (already configured for
  Go modules in this repo).

bun was considered and rejected: faster but a few Nuxt modules still
have rough edges, and the maturity gap is not worth the speed gain
for a project this size.

## Consequences

- Adding `@nuxt/fonts` as a build dependency from day one — no risk
  of accidentally shipping the Google CDN `@import` to production.
- A new tech-debt item is opened to remind the team to migrate the
  mock CSS away from the Google CDN at scaffold time.
- The Nuxt repo lives alongside the Go backend in the same
  monorepo (under `frontend/`, per CLAUDE.md repository layout).
- Renovate config will need an extra entry for `frontend/package.json`
  — to track in a follow-up after the first scaffold lands.
- The signup-organizer page, generic empty states, and 404/500 screens
  are NOT in the design system kits — they will be designed in Vue at
  the moment they are needed, reusing the locked tokens.

## Alternatives considered and rejected

- **Google Fonts CDN** — RGPD risk on a `.fr` domain, third-party
  runtime dependency, extra DNS lookup. The CDN convenience is not
  worth the cost.
- **System fonts only** — would skip the typography choice entirely
  but lose the brand identity that Inter brings. The design system
  has already locked Inter as the team-chosen typeface.
- **Material UI / Vuetify** — opinionated component libraries that
  would fight the locked design system at every step.
- **CSS Modules / vanilla CSS** instead of Tailwind — more boilerplate,
  no atomic-class velocity, harder to keep tokens consistent across
  components.
- **bun** as package manager — see Decision section above.
