import { describe, expect, it } from 'vitest'
import { toFriendlyError } from './errorMessages'
import { ApiError } from '~/composables/useApi'

describe('toFriendlyError', () => {
  it('maps a known backend public message to its French title and detail', () => {
    const err = new ApiError(400, 'invalid team assignment')
    expect(toFriendlyError(err)).toEqual({
      title: 'Génération impossible',
      message: 'Il n\'y a pas assez de joueurs confirmés pour former deux équipes.',
    })
  })

  it('falls back to HTTP status mapping when message is unknown', () => {
    const err = new ApiError(404, 'something we never mapped')
    expect(toFriendlyError(err)).toEqual({
      title: 'Introuvable',
      message: 'La ressource demandée n\'existe pas.',
    })
  })

  it('falls back to the raw message when both message and status are unknown', () => {
    const err = new ApiError(418, 'kettle is boiling')
    expect(toFriendlyError(err)).toEqual({
      title: 'Erreur',
      message: 'kettle is boiling',
    })
  })

  it('handles native Error by surfacing its message', () => {
    expect(toFriendlyError(new Error('boom'))).toEqual({
      title: 'Erreur',
      message: 'boom',
    })
  })

  it('handles unknown thrown values with a generic message', () => {
    expect(toFriendlyError('not even an error')).toEqual({
      title: 'Erreur',
      message: 'Une erreur inattendue est survenue.',
    })
    expect(toFriendlyError(null)).toEqual({
      title: 'Erreur',
      message: 'Une erreur inattendue est survenue.',
    })
  })

  it('translates rate-limit messages with a wait hint', () => {
    const err = new ApiError(429, 'rate limit exceeded')
    const friendly = toFriendlyError(err)
    expect(friendly.title).toBe('Trop de requêtes')
    expect(friendly.message).toContain('Patiente')
  })

  it('translates 409 conflicts with a meaningful context', () => {
    const err = new ApiError(409, 'match is full')
    expect(toFriendlyError(err).title).toBe('Match complet')
  })
})
