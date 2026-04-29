# SMO Design System

**SMO** (sportpotes.fr) is a web platform for organizing amateur 5v5 football matches between friends. The product is **organizer-facing only** — players don't have accounts; they join via invitation tokens.

This design system locks the visual identity for the entire product. It is **dark by default**, sportive but not corporate, and deliberately restrained.

---

## Sources

This design system was created from a written spec (no codebase or Figma yet). Key references provided by the team:

- **Mon Petit Pronos** (mpp.football) — the closest tonal reference: dark, sportive, friendly, not corporate.
- **Linear** (linear.app) — for sober typography and form quality.
- **Cal.com** — for clean form patterns.
- Production stack target: **Nuxt 3 SPA + Vue 3 + Tailwind CSS + lucide-vue-next**, mobile-first (320px minimum).

When the codebase exists, this README should be updated with paths to the actual Nuxt project and any Figma file.

---

## Index

| File / Folder | What's in it |
|---|---|
| `README.md` | This file — brand, content fundamentals, visual foundations, iconography |
| `SKILL.md` | Agent skill manifest — entry point when this design system is loaded as a skill |
| `colors_and_type.css` | All design tokens as CSS variables: colors, typography, spacing, radii, shadows |
| `assets/` | Logos, wordmarks, brand marks |
| `fonts/` | Local font fallbacks (currently empty — using Google Fonts) |
| `preview/` | HTML cards rendered in the Design System tab |
| `ui_kits/login/` | Login page UI kit — JSX components + interactive `index.html` |

---

## Content Fundamentals

### Voice & tone

SMO is a **tool for friends organizing weekend football**, not a corporate SaaS. Copy is direct, French, and short. No marketing speak. No exclamation marks. No "Hey 👋". No emoji.

- **Language:** French. The product is French-only at MVP.
- **Address form:** **tutoiement** (informal `tu`) for organizers — they're casual users, not enterprise buyers. *"Tu n'as pas encore de compte ?"* — never *"Vous n'avez pas..."*.
- **Casing:** Sentence case everywhere. Never Title Case. Never ALL CAPS (except possibly small uppercase labels with letter-spacing, used sparingly).
- **Density:** Aggressive minimalism. If a label can be removed, remove it. The login page has 2 fields and 1 button — nothing else above the fold.
- **Football register:** Use the right words — *match*, *équipe*, *organisateur*, *invitation*, *5v5*. Avoid generic SaaS terms (*dashboard*, *workspace*, *onboarding*) unless there's no equivalent.

### Examples (login page copy, locked)

| Element | Copy |
|---|---|
| Wordmark | `SMO` |
| Page title | `Connexion organisateur` |
| Email label | `Email` |
| Email placeholder | `toi@exemple.fr` |
| Password label | `Mot de passe` |
| Password placeholder | `••••••••` |
| Primary CTA | `Se connecter` |
| Loading CTA | `Connexion…` |
| Error message | `Identifiants incorrects.` |
| Secondary link | `Pas encore de compte ? S'inscrire` |
| Footer note | `Les joueurs n'ont pas besoin de compte — ils accèdent par lien d'invitation.` |

### What we never write

- ❌ "Welcome back!" / "Bienvenue !" — assumed, redundant.
- ❌ "Heureux de te revoir 🎉" — emoji, fake warmth.
- ❌ "Connectez-vous" — vouvoiement.
- ❌ "Mot de passe oublié ?" — out of scope at MVP.
- ❌ "Se souvenir de moi" — JWT handles persistence.

---

## Visual Foundations

### Mode

**Dark by default. There is no light mode at MVP.** The body fills with deep black `#0E1014` and never reverts. This is non-negotiable — it's part of the brand.

### Color

A strict, narrow palette. **Five neutrals + two action blues + two team colors + one warning yellow.** Anything else is forbidden — especially purple, orange, pink, warm-tinted gradients, and pure white backgrounds.

| Role | Token | Hex | Use |
|---|---|---|---|
| Body background | `--bg-base` | `#0E1014` | Page background, only |
| Elevated surface | `--bg-elevated` | `#1A1F26` | Cards, input fills, popovers |
| Subtle surface | `--bg-subtle` | `#141821` | Hover/pressed on elevated |
| Border / divider | `--border-default` | `#222831` | Hairline separators |
| Mid grey | `--fg-muted` | `#4A5560` | Secondary text, placeholder, icon-default |
| Off-white | `--fg-default` | `#F5F6F8` | Primary text |
| Pure white | `--fg-emphasis` | `#FFFFFF` | Reserved — wordmark, key numbers |
| Action primary | `--action-primary` | `#2080FF` | Primary button, focused field ring, active link |
| Action primary hover | `--action-primary-hover` | `#1452C9` | Primary button hover/pressed |
| Team red | `--team-red` | `#DC2A3B` | Team A, also error state |
| Team green | `--team-green` | `#30D158` | Team B, also success |
| Warning | `--warn` | `#F5C518` | Alerts only |

> The error state and the red team share a hue. That's intentional — it keeps the palette tight. Errors use `--team-red` directly.

### Typography

Two type levels maximum on any screen — **a title and body**. Numerical content (scores, jersey numbers, timers) uses a mono variant for tabular alignment.

- **Display & UI:** **Inter Variable** (wght 100–900). Loaded via Google Fonts. The team's chosen typeface — locked.
- **Numerics:** **JetBrains Mono** (tabular figures, used for scores / jersey numbers / timers).
- Inter is configured with `font-feature-settings: 'cv11', 'ss01', 'tnum'` (alternate single-storey `a`, slashed `0`, tabular numerals) for a slightly more geometric read in UI.

> Inter is generally a "commodity" stack — but it's the team's call here, so it's locked. We lean on the variable axis (use 400/500/600 most; reserve 700 for the wordmark only) and the cv/ss features to keep it from looking generic.

Type scale (mobile-first, all sizes in px / line-height in rem):

| Token | Size / Line-height | Weight | Use |
|---|---|---|---|
| `--type-display` | 28 / 1.2 | 600 | Page titles like "Connexion organisateur" |
| `--type-body` | 15 / 1.5 | 400 | Default text, labels, paragraphs |
| `--type-input` | 16 / 1.4 | 400 | Form inputs (16px prevents iOS auto-zoom) |
| `--type-button` | 15 / 1 | 500 | Buttons |
| `--type-caption` | 13 / 1.4 | 400 | Footer notes, micro-labels |
| `--type-mono-num` | 15 / 1 | 500 | Scores, numbers |

### Spacing

A **4px base** scale. Avoid arbitrary px values; always use a token.

`4 · 8 · 12 · 16 · 20 · 24 · 32 · 40 · 56 · 80`

The login form sits on a **vertical rhythm of 16px between groups**, **8px between label and field**, with **24px below the title**.

### Backgrounds

- **Solid only.** No gradients, no images, no textures, no noise.
- The only "depth" comes from a 1-level surface system: base (`#0E1014`) → elevated (`#1A1F26`) → subtle hover (`#141821`).
- No protection gradients, no scrims, no blur overlays. The product never overlays text on imagery at MVP.

### Borders

- **Used sparingly.** Prefer separating elements with spacing or a fill change rather than a border.
- When a border is used: **1px solid `#222831`**. That's the only border color.
- Inputs in the **default state have no border** — they sit on `--bg-elevated` against the `--bg-base` page, and the contrast is enough. They get a **2px ring** on focus, not a border swap.

### Corner radii

| Token | Value | Use |
|---|---|---|
| `--radius-sm` | 6px | Small chips, badges |
| `--radius-md` | 10px | Inputs, buttons (default) |
| `--radius-lg` | 14px | Cards, larger surfaces |
| `--radius-pill` | 999px | Pills only |

The login uses `--radius-md` consistently for inputs and the button, so they read as a single stack.

### Shadows

**Almost none.** Dark UI doesn't benefit from drop shadows the way light UI does — they look like grime. We use:

- `--shadow-focus`: `0 0 0 2px rgba(32, 128, 255, 0.45)` — focus ring on inputs and buttons.
- `--shadow-elevated`: `0 8px 24px -8px rgba(0, 0, 0, 0.6)` — only for popovers / modals, never on form cards.

There is no card-floating-on-page shadow. The login form has **no shadow**.

### Hover states

- **Buttons (primary):** background shifts from `#2080FF` → `#1452C9`. No scale, no shadow change.
- **Buttons (ghost / icon):** background `transparent → rgba(255,255,255,0.06)`.
- **Links:** color stays the same, underline appears (or, for already-underlined links, opacity goes from 0.85 → 1).
- **Inputs:** no hover state (focus is the only state change).

### Press states

- **Buttons:** background goes one notch darker than hover. No transform, no shrink — those feel toy-like for a sober UI.
- Icon buttons: background `rgba(255,255,255,0.10)`.

### Focus states

- **Inputs:** 2px ring in `--action-primary` (`#2080FF`), no border swap. Inset shadow for clarity on dark.
- **Buttons:** same 2px ring on `:focus-visible`.
- **Links:** underline + 2px ring offset 2px.
- Always visible on keyboard focus. Never hidden (`:focus-visible` only — mouse clicks don't show the ring).

### Animation

Sober, never bouncy.

- **Default duration:** 150ms.
- **Default easing:** `cubic-bezier(0.4, 0, 0.2, 1)` (standard ease-out).
- **What we animate:** background color on buttons, opacity on the password-show toggle, the loading spinner.
- **What we don't:** page transitions, slide-ins, modal entrances at MVP. Kept boring on purpose.

### Transparency & blur

Not used in this product at MVP. No glass, no frosted modals, no backdrop blur. If we need a modal later, it's a solid `#1A1F26` with a `rgba(0,0,0,0.6)` scrim — no blur.

### Layout rules

- **Mobile-first.** Designs target 320px minimum. Single column, full-width controls.
- **Desktop:** the login form sits **centered, max-width 380px, vertically centered in the viewport**. No sidebar. No hero. The page is the form.
- The **wordmark sits ABOVE the form** at all sizes — never floated in a corner.
- The **footer note pins to the bottom of the viewport** on mobile (or sits 32px below the form on short screens), and sits **directly below the form** with 32px gap on desktop.

### Imagery

**None at MVP.** The product is functional, not marketing. There are no hero images, no decorative illustrations, no team avatars yet. Field/pitch imagery may show up in deeper screens (match cards) but is out of scope here.

---

## Iconography

- **Library:** **Lucide** (`lucide-vue-next` in production). No alternatives.
- **Style:** stroke-based, 2px stroke, 24px default canvas, 20px in compact UI, 16px inside inputs.
- **Color:** icons inherit `currentColor`. They're never colored standalone — color comes from the surrounding text token.
- **Loading via CDN in mocks:** the `<lucide-icon>` web component from `https://unpkg.com/lucide@latest`, OR raw SVG paths copied from lucide.dev. We don't ship a custom icon font.
- **Emoji:** **never.** No emoji in product UI, copy, error messages, or marketing.
- **Unicode chars:** the ` · ` middle dot is acceptable as a separator; the `→` arrow is acceptable inline in CTAs as a directional cue, but we prefer Lucide's `arrow-right` for buttons.
- **Icons used on the login page:** `eye`, `eye-off` (password toggle), `loader-2` (spinner), `alert-circle` (error message).

---

## Component principles

- **One thing per screen** — the login page exists to log you in, nothing else.
- **No floating cards over a complex background.** The form sits directly on the page.
- **No left-side decoration on desktop.** No Stripe-mode hero. No marketing imagery.
- **Max 2 type levels** (display + body). Caption is a third only when necessary (footer notes, helper text), and it's always smaller and `--fg-muted`.
- **Borders are a last resort.** Use spacing or a fill change first.
