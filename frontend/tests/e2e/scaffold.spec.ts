// Sanity e2e for the scaffold: boot the SPA, the wordmark and the
// design-token preview must render. Real e2e flows (login, group
// CRUD) live in dedicated spec files alongside the corresponding
// pages.

import { test, expect } from '@playwright/test'

test('scaffold: home page renders the wordmark and the token preview', async ({ page }) => {
  await page.goto('/')

  await expect(page.getByRole('heading', { name: /SMO/ })).toBeVisible()
  await expect(page.getByText('Scaffold prêt')).toBeVisible()
  await expect(page.getByText('action-primary')).toBeVisible()
})
