# ADR 0009 — Votes as invitation-token capabilities

**Status:** Accepted (2026-06-10)

## Context

The original `POST /votes` contract accepted
`{match_id, voter_id, voted_id, score}` on a public route. Nothing tied
`voter_id` to the caller: anyone who knew (or enumerated) player and
match identifiers could cast votes on behalf of any player. The flaw was
dormant only because no UI consumed the endpoint. Building the player
vote page would have armed it.

Two adjacent privacy problems surfaced at the same time. The vote page
design promises players that votes are "définitifs et anonymes — vos
coéquipiers ne sauront pas qui les a notés", yet `GET /matches/:id/votes`
and `GET /votes/:id` were public and returned raw votes including
`voter_id`. And the same-team rule ("players rate teammates, never
opponents") was documented in the SQL schema but enforced nowhere.

## Decision

The invitation token becomes the voting capability, exactly as it is
the RSVP capability:

- `POST /votes` now takes `{token, voted_id, score}`. The backend
  resolves the token hash to its invitation, derives the voter from it,
  and verifies that the bearer confirmed attendance and that voter and
  target are on the same team (`Match.TeamOf` / `TeammatesOf`, new pure
  entity queries). Spoofing is impossible by construction: the client
  never names itself.
- `POST /votes/context {token}` is the single read model for the page:
  match block, the bearer's teammates (with the bearer's own scores and
  a "closed matches played together" aggregate), collective progress,
  and — once the match is closed — per-player averages, vote counts,
  delta versus the group's previous decided match, and the bearer's own
  line. Players only ever see aggregates.
- `GET /votes/:id` and `GET /matches/:id/votes` move to the
  organizer-protected group. Raw votes never leave the backend
  unauthenticated.
- Invitation expiry does NOT gate voting. Expiry protects the RSVP
  window; the voting window is the match status alone (`completed`
  opens it, `closed` ends it). Enforcing expiry would lock out players
  invited more than five days before the match ended.

## Consequences

- One token shared by invitation and vote means one link for the player
  for the whole match lifecycle; the invitation context now exposes
  `match_status` so the invite page can route to the vote page once the
  match completes.
- The old vote contract had no client, so it was replaced without a
  compatibility window.
- `CountClosedMatchesTogether` joins invitations of closed matches; it
  runs as one grouped query per context call and is read-only.
- A declined invitation yields 403 on both vote endpoints
  (`ErrNotConfirmedParticipant`); a cross-team target yields 400
  (`ErrNotTeammates`). Both are new domain errors, distinct from
  `ErrSelfVote` which stays owned by the Vote entity.
