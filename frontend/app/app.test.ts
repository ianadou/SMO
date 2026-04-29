// @vitest-environment nuxt
// Sanity test for the scaffold: the root app renders the wordmark
// and the design-token preview without throwing. Real component
// tests for the login and other pages land alongside them in
// follow-up PRs.
//
// The `nuxt` environment is required so mountSuspended has access
// to the Nuxt app context (auto-imports, composables).

import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import App from './app.vue'

describe('App scaffold', () => {
  it('renders the wordmark and the token preview', async () => {
    const wrapper = await mountSuspended(App)

    expect(wrapper.text()).toContain('SMO')
    expect(wrapper.text()).toContain('Scaffold prêt')
    expect(wrapper.text()).toContain('action-primary')
  })
})
