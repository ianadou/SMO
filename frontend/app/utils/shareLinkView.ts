import type { JoinView, ShareLinkLoadOutcome } from '~/types/shareLink'

export function resolveJoinView(outcome: ShareLinkLoadOutcome): JoinView {
  if (outcome.kind === 'invalid') return 'invalid'
  if (outcome.kind === 'error') return 'error'

  const { match_status } = outcome.context
  return match_status === 'draft' || match_status === 'open' ? 'roster' : 'locked'
}
