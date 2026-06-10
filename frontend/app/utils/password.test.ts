import { describe, it, expect } from 'vitest'
import { passwordStrength, passwordStrengthLabel, isEmailFormat } from './password'

function compose(parts: { lower?: number, upper?: number, digit?: number, special?: number }): string {
  return [
    'a'.repeat(parts.lower ?? 0),
    'A'.repeat(parts.upper ?? 0),
    '1'.repeat(parts.digit ?? 0),
    '!'.repeat(parts.special ?? 0),
  ].join('')
}

describe('passwordStrength', () => {
  it('returns 0 for an empty string', () => {
    expect(passwordStrength('')).toBe(0)
  })

  it('returns 1 for fewer than 12 characters', () => {
    expect(passwordStrength(compose({ lower: 1 }))).toBe(1)
    expect(passwordStrength(compose({ lower: 3, digit: 4 }))).toBe(1)
    expect(passwordStrength(compose({ lower: 5, digit: 5 }))).toBe(1)
  })

  it('returns 2 for 12+ chars with a single character class', () => {
    expect(passwordStrength(compose({ lower: 12 }))).toBe(2)
    expect(passwordStrength(compose({ digit: 12 }))).toBe(2)
  })

  it('returns 3 for 12+ chars with 2 classes', () => {
    expect(passwordStrength(compose({ lower: 6, digit: 6 }))).toBe(3)
    expect(passwordStrength(compose({ upper: 4, lower: 8 }))).toBe(3)
  })

  it('returns 4 for 12+ chars with 3+ classes', () => {
    expect(passwordStrength(compose({ upper: 1, lower: 7, digit: 4 }))).toBe(4)
    expect(passwordStrength(compose({ upper: 1, lower: 7, digit: 2, special: 2 }))).toBe(4)
  })

  it('does not return 4 for 12+ chars with only 2 classes', () => {
    expect(passwordStrength(compose({ lower: 8, digit: 4 }))).toBe(3)
  })
})

describe('passwordStrengthLabel', () => {
  it('returns an empty string for level 0', () => {
    expect(passwordStrengthLabel(0)).toBe('')
  })

  it('returns the matching French label for each strength level', () => {
    expect(passwordStrengthLabel(1)).toBe('Faible')
    expect(passwordStrengthLabel(2)).toBe('Moyen')
    expect(passwordStrengthLabel(3)).toBe('Fort')
    expect(passwordStrengthLabel(4)).toBe('Très fort')
  })
})

describe('isEmailFormat', () => {
  it('accepts well-formed addresses', () => {
    expect(isEmailFormat('alex@example.fr')).toBe(true)
    expect(isEmailFormat('a.b+c@sub.domain.co.uk')).toBe(true)
  })

  it('rejects strings without @', () => {
    expect(isEmailFormat('plainaddress')).toBe(false)
  })

  it('rejects strings without a domain dot', () => {
    expect(isEmailFormat('foo@bar')).toBe(false)
  })

  it('rejects strings with whitespace', () => {
    expect(isEmailFormat('foo bar@example.fr')).toBe(false)
    expect(isEmailFormat('foo@example .fr')).toBe(false)
  })
})
