import { test as base, expect } from '@playwright/test'

export const test = base.extend({
  page: async ({ page }, use) => {
    await page.route('**/api/v1/**', route => route.abort('failed'))
    await use(page)
  },
})

export { expect }
export type { Page } from '@playwright/test'
