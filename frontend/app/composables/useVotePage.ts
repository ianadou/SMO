import { ref } from 'vue'
import { ApiError, useApi } from './useApi'
import type { CastVoteResult, VoteContext, VoteLoadOutcome } from '~/types/vote'

export interface VoteApi {
  post: <T>(path: string, body: unknown) => Promise<T>
}

export function useVotePage(token: string, api: VoteApi = useApi()) {
  const outcome = ref<VoteLoadOutcome | null>(null)
  const loading = ref(true)
  const submitting = ref(false)

  async function load() {
    loading.value = true
    try {
      const context = await api.post<VoteContext>('/votes/context', { token })
      outcome.value = { kind: 'ok', context }
    }
    catch (err) {
      if (err instanceof ApiError && err.status === 403) {
        outcome.value = { kind: 'not-participant' }
      }
      else if (err instanceof ApiError && (err.status === 404 || err.status === 400)) {
        outcome.value = { kind: 'invalid' }
      }
      else {
        outcome.value = { kind: 'error' }
      }
    }
    finally {
      loading.value = false
    }
  }

  async function submitVotes(votes: CastVoteResult[]): Promise<string[]> {
    submitting.value = true
    const failedIds: string[] = []
    try {
      for (const vote of votes) {
        try {
          await api.post('/votes', { token, voted_id: vote.voted_id, score: vote.score })
        }
        catch (err) {
          const alreadyVoted
            = err instanceof ApiError
              && err.status === 409
              && err.publicMessage.includes('already voted')
          if (!alreadyVoted) failedIds.push(vote.voted_id)
        }
      }
      await load()
    }
    finally {
      submitting.value = false
    }
    return failedIds
  }

  return { outcome, loading, submitting, load, submitVotes }
}
