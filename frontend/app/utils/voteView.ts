import type { VoteLoadOutcome, VoteView } from '~/types/vote'

export function resolveVoteView(outcome: VoteLoadOutcome): VoteView {
  if (outcome.kind === 'invalid') return 'invalid'
  if (outcome.kind === 'not-participant') return 'not-participant'
  if (outcome.kind === 'error') return 'error'

  const { status, teammates } = outcome.context
  if (status === 'closed') return 'results'
  if (status !== 'completed') return 'too-early'

  const remaining = teammates.filter((teammate) => teammate.your_score === null)
  return remaining.length === 0 ? 'submitted' : 'rate'
}
