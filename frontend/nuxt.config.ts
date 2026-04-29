// https://nuxt.com/docs/api/configuration/nuxt-config

import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',

  // SPA mode: no SSR, the app boots in the browser. SMO is an
  // organizer tool with no SEO need; SPA simplifies the deploy
  // (single static bundle behind nginx) and keeps the backend free
  // of HTML rendering concerns. See ADR 0005.
  ssr: false,

  devtools: { enabled: true },

  // Global stylesheet that declares Tailwind v4 theme + the design
  // system tokens (colors, spacing, radii, shadows) as CSS
  // variables. The design system file stays the single source of
  // truth.
  css: ['~/assets/css/main.css'],

  // Tailwind v4: configured via the official Vite plugin. CSS-first
  // setup (theme tokens declared inside the global stylesheet using
  // @theme), no separate tailwind.config.ts.
  vite: {
    plugins: [tailwindcss()],
  },

  // Self-hosted fonts via @nuxt/fonts. Inter for UI/display,
  // JetBrains Mono for tabular numerics. NOT Google Fonts CDN —
  // see ADR 0005 (RGPD risk on .fr).
  modules: ['@nuxt/fonts'],
  fonts: {
    families: [
      {
        name: 'Inter',
        provider: 'google',
        weights: [400, 500, 600, 700],
      },
      {
        name: 'JetBrains Mono',
        provider: 'google',
        weights: [400, 500],
      },
    ],
  },

  // Default page metadata. The product is French-only at MVP per
  // the design system voice/tone rules.
  app: {
    head: {
      htmlAttrs: { lang: 'fr' },
      title: 'SMO — Sport Match Organizer',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Organise tes matchs de foot 5v5 entre potes.' },
      ],
    },
  },
})
