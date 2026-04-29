// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import StrengthMeter from './StrengthMeter.vue'

describe('StrengthMeter', () => {
  it('renders 4 segments', async () => {
    const wrapper = await mountSuspended(StrengthMeter, { props: { level: 0 } })
    expect(wrapper.findAll('span').filter(s => s.classes('h-1')).length).toBe(4)
  })

  it('shows no label when level is 0', async () => {
    const wrapper = await mountSuspended(StrengthMeter, { props: { level: 0 } })
    expect(wrapper.text()).toContain('Au moins 8 caractères')
    expect(wrapper.text()).not.toContain('Faible')
    expect(wrapper.text()).not.toContain('Très fort')
  })

  it('shows the matching French label for each level', async () => {
    for (const [level, label] of [[1, 'Faible'], [2, 'Moyen'], [3, 'Fort'], [4, 'Très fort']] as const) {
      const wrapper = await mountSuspended(StrengthMeter, { props: { level } })
      expect(wrapper.text()).toContain(label)
    }
  })
})
