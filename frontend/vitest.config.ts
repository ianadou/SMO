// Vitest configuration for unit + component tests.
//
// Uses the @nuxt/test-utils preset so tests can mount Nuxt
// components with the real auto-imports, composables, and Tailwind
// classes resolved. happy-dom (lighter than jsdom) provides the DOM.
//
// Run: pnpm test            (CI / one-shot)
//      pnpm test:watch      (TDD)
//      pnpm test:ui         (browser UI)

import { defineVitestConfig } from '@nuxt/test-utils/config'

export default defineVitestConfig({
  test: {
    environment: 'happy-dom',
    globals: true,
    // Component tests live next to the components (foo.vue +
    // foo.test.ts). E2E tests are separate, handled by Playwright.
    include: ['app/**/*.test.ts', 'app/**/*.spec.ts'],
    exclude: ['node_modules', '.nuxt', '.output', 'tests/e2e/**'],
  },
})
