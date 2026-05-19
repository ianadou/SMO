import { describe, it, expect } from 'vitest'
import { formatMatchDate, formatMatchTime, parseCapacity } from './inviteFormat'

describe('formatMatchDate', () => {
  it('formats an RFC3339 instant as a capitalised French long date in Europe/Paris', () => {
    expect(formatMatchDate('2026-05-07T19:30:00+02:00')).toBe('Jeudi 7 mai 2026')
  })
})

describe('formatMatchTime', () => {
  it('formats the time as HHhMM zero-padded in Europe/Paris', () => {
    expect(formatMatchTime('2026-05-07T19:30:00+02:00')).toBe('19h30')
    expect(formatMatchTime('2026-05-07T09:05:00+02:00')).toBe('09h05')
  })
})

describe('parseCapacity', () => {
  it('splits "10 (5v5)" into the headcount and the format', () => {
    expect(parseCapacity('10 (5v5)')).toEqual({ count: '10', format: '5v5' })
  })

  it('falls back to the raw string when no parenthesis is present', () => {
    expect(parseCapacity('12')).toEqual({ count: '12', format: '' })
  })
})
