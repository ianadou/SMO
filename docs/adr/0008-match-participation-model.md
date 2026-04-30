# ADR 0008 — Match participation model

**Status:** Accepted (2026-04-30)

## Context

Before this change, an SMO `Match` had no notion of "who is in it"
until teams were generated. The `invitations` table carried a token
hash, an expiration, and a `used_at` timestamp — but no link to a
specific player. The Match entity itself stored no participant list
either; a player was only attached to a match transitively, via
`team_players` after teams existed.

That gap blocked everything downstream. The match-detail page cannot
answer "who is coming?" without joining invitations, but invitations
have no `player_id`. Team generation cannot run a strategy on the
"confirmed players" set because that set is not addressable. SF05.3
(validation présence via lien d'invitation) loses its grounding.

This ADR records the decision to make a confirmed invitation the
source of truth for participation, the constraints around it, and the
trade-offs accepted along the way.

## Decision

### One invitation = one player, picked at creation time

Migration `005_add_player_id_to_invitations.sql` adds `player_id` as
a `NOT NULL` foreign key on `invitations`, with `ON DELETE CASCADE`
and an index. Every invitation is now created **for** a specific
player; the organizer picks the recipient when generating the link.

`CreateInvitationUseCase` requires the player_id in its input and
verifies that `player.GroupID == match.GroupID` before persisting.
This closes a cross-group leak: an organizer cannot invite a player
from another group to one of their matches, even with a hand-crafted
HTTP request.

### Confirmed invitation = participant, capped at 10 (FCFS)

A confirmed invitation is one whose `used_at` is not NULL. We cap the
number of confirmed invitations per match at
`entities.MaxParticipantsPerMatch = 10`, which matches the 5-vs-5
team layout. The 11th attempt to accept is rejected with
`ErrMatchFull` (mapped to HTTP 409 by the existing error mapper).

The selection policy is **first-come-first-served** (FCFS): the first
ten players to click their invitation link are the ten who play.
There is no organizer-side review step; if the organizer needed
control, they would not invite more than ten players.

### Migration is fail-loud, not destructive

A naive `ALTER TABLE invitations ADD COLUMN player_id NOT NULL ...`
fails on a non-empty table because Postgres has no value to fill in.
The two options are:

1. Silently `DELETE FROM invitations` before the alter. Discards
   user data without warning if anyone deployed the previous schema
   to a real environment.
2. Fail-loud with `RAISE EXCEPTION` if the table is non-empty,
   forcing a manual backfill decision before the migration runs.

We chose (2). The current dev environment has zero rows, so the
guard never triggers there; in any environment that does have data,
the operator gets a clear error and decides whether to backfill or
truncate. Bruyant beats silencieux for a destructive operation.

### Race window on `count + accept` is acknowledged, not fixed

`AcceptInvitation` reads the count of confirmed invitations, then
calls `MarkAsUsed`. Between the two, a concurrent acceptance could
slip in, taking the count above ten. We accept this for the MVP:
SMO's scale (one match, two-digit invitee count) makes the race
unlikely, and the cost of fixing it (advisory lock, partial unique
index, or SERIALIZABLE transaction) is disproportionate to the risk.

If real attack data shows this is exploitable, the proper fix is a
Postgres advisory lock keyed on `match_id` around the count + update
sequence.

### `GET /matches/:id/participants` is authenticated, not public

§3 Personas of the product spec is explicit: an invited player must
not see other matches' rosters. A public endpoint here would let any
caller enumerate participants across the platform, which is both a
privacy issue (player names) and a competitive issue (other
organizers can see who you've recruited).

The cleanest implementation is dual-mode authentication: accept a
JWT organizer who owns the match's group, OR an invitation token
whose `match_id` matches the URL. Both modes are described in the
follow-up issue; this PR ships behind a plain JWT-organizer gate
(garde-fou) so the endpoint is not public, with the cross-organizer
leak as a documented temporary trade-off until the dual-mode
middleware lands.

## Alternatives rejected

### NULL-able `player_id` with backfill

Rejected. There is no legacy data to preserve. NULL-able + backfill
introduces a window where the schema allows orphan invitations,
which the application code would have to defensively handle
forever. NOT NULL from the start is simpler and stricter.

### Organizer chooses N participants among M confirmed

The product alternative considered for over-confirmation. Rejected
because:

- It requires a new flow (organizer review step) that does not
  exist in the current backend.
- It splits the participation model into "accepted" vs "selected"
  states that the rest of the system would have to know about.
- FCFS is operationally equivalent at this scale: the organizer
  invites exactly the number they need, and the first to confirm
  are the ones who play.

If the product evolves toward over-recruitment as a feature, the
selection step belongs in its own ADR.

### Trigger or partial unique index for the race fix

Rejected for the MVP. A `CREATE UNIQUE INDEX ... WHERE used_at IS NOT
NULL` partial index on `(match_id, slot_number)` would require
allocating a slot at acceptance time, which complicates the use case
without buying anything until concurrent acceptance is a real
problem. The advisory-lock fix described above is cheaper to add
later.

### Public `/participants` with token in body

Rejected. Putting the invitation token in the request body for a
GET breaks REST conventions, and a query string token gets logged in
HTTP access logs, increasing the surface for accidental disclosure.
The dual-auth middleware (planned follow-up) reads the token from a
designated query parameter and explicitly redacts it before logging.

## Consequences

### Targeted DoS via lockout-style attack is mitigated

A confirmed invitation cannot be re-used (`MarkAsUsed` rejects
`AlreadyUsed`). An attacker would need to *steal* a player's token
to confirm on their behalf — a separate threat model already
addressed by the token's high entropy.

### Cross-organizer enumeration leak is open until PR A.1

Any authenticated organizer can call
`GET /api/v1/matches/<other-organizer-match>/participants` today.
This is documented in the PR description and tracked as a follow-up
issue; the user-facing impact is limited (the leak is between
organizers, not toward players) and the dual-auth middleware that
closes it is the entire scope of PR A.1.

### `CreateInvitation` contract change

`POST /api/v1/invitations` now requires `player_id` in the body.
This is breaking, but no production client uses the endpoint yet:
the frontend never integrated the invitation flow, and there are no
external consumers.

### The fail-loud migration must run before any deploy

If the migration runs against a non-empty `invitations` table, the
deploy aborts with a clear error. The runbook entry is: inspect the
existing rows, backfill `player_id` for each, then re-run.

## Future work

- **PR A.1** — Dual-mode middleware on `/matches/:id/participants`
  accepting JWT-organizer-of-group OR invitation token. Closes the
  cross-organizer leak.
- **PR B** — Team generation use case + endpoint, taking confirmed
  participants as input.
- **Race fix** — Advisory lock around `count + MarkAsUsed` if real
  data shows the race matters.
- **Bulk invite** — Endpoint to create N invitations at once for a
  match, returning N plain tokens. Saves the organizer N round-trips.
