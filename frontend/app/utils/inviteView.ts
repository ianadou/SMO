import type { InviteView, LoadOutcome } from '~/types/invitation'

export function resolveInviteView(outcome: LoadOutcome): InviteView {
  if (outcome.kind === 'invalid') return 'invalid'
  if (outcome.kind === 'error') return 'error'

  const { state, response } = outcome.context
  if (state === 'expired') return 'expired'
  if (state === 'locked') return response === 'pending' ? 'locked-pending' : 'result'
  return response === 'pending' ? 'initial' : 'result'
}
