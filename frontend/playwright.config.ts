// Playwright configuration for end-to-end tests.
//
// Boots the Nuxt dev server on port 3001 (avoids colliding with a
// dev server the developer may have running on 3000), then runs
// each test in a fresh browser context.
//
// Run: pnpm e2e            (headless, CI mode)
//      pnpm e2e:headed     (debugging with the browser visible)

import { defineConfig, devices } from '@playwright/test'

const E2E_PORT = 3001
const E2E_HOST = '127.0.0.1'

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL: `http://${E2E_HOST}:${E2E_PORT}`,
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: `pnpm dev --port ${E2E_PORT} --host ${E2E_HOST}`,
    url: `http://${E2E_HOST}:${E2E_PORT}`,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
})
