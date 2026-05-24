import { describe, it, expect } from 'vitest'
import { playerInitials } from './playerInitials'

describe('playerInitials', () => {
  it('takes first letter of first and last word for multi-word names', () => {
    expect(playerInitials('Alex L.')).toBe('AL')
    expect(playerInitials('Jean Michel Dupont')).toBe('JD')
  })

  it('takes the first two letters for a single-word name', () => {
    expect(playerInitials('Zlatan')).toBe('ZL')
  })

  it('keeps a single character when the name is one letter', () => {
    expect(playerInitials('X')).toBe('X')
  })

  it('uppercases accented initials', () => {
    expect(playerInitials('Éric Cédric')).toBe('ÉC')
  })

  it('returns an empty string for blank input', () => {
    expect(playerInitials('   ')).toBe('')
  })
})
