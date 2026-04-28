# ADR 0003 — Discord webhook URL storage on Group

**Status:** Accepted (2026-04-28)

## Context

Per the project specifics in CLAUDE.md, "Discord webhooks replace
real-time notifications": when a Match transitions to `teams_ready`,
the system posts a notification to a Discord channel. The webhook URL
holds the authority to post on that channel and is therefore a secret.

This ADR records the decisions that govern **where the URL is stored
and how it travels through the system**, independent of the notifier
adapter that consumes it. The notifier wiring is the subject of a
separate ADR (created if non-trivial decisions emerge during its
implementation in the follow-up PR).

## Decisions

### D1 — One webhook URL per Group, optional

Each `Group` aggregate carries a single optional `webhookURL` field.
Empty value = the group has no Discord channel configured and no
notification is sent.

- **Rejected: one webhook per Match.** The Discord channel is a
  property of the group of friends, not of an individual match.
  Per-match would mean re-typing the same URL for every match.
- **Rejected: one global webhook (env var) for all matches.** The
  domain model maps "Group" to "circle of friends", which typically
  has its own Discord. A global URL would funnel notifications from
  unrelated groups into the same channel.

### D2 — Plain-text storage in Postgres, documented

The URL is stored unencrypted in the `groups.discord_webhook_url`
column. Mitigations are listed in the Security considerations
section below.

- **Rejected: column-level encryption with `pgcrypto`.** Adding
  encryption to a single column while every other secret in the DB
  (player tokens are hashed, but bcrypt password hashes are stored
  as-is in `organizers`) is in clear creates an inconsistent threat
  model and gives a false sense of security. Encryption-at-rest is
  the right answer when applied uniformly (full-disk encryption at
  the deploy layer, or pgcrypto on every secret column with a KMS).
  Until SMO adopts one of those holistically, plain-text with strong
  access controls is more honest.
- **Rejected: separate `webhooks` table with FK from `groups`.**
  Overkill for a single 1:1 nullable field.

### D3 — Strict validation in the domain entity

`entities.NewGroup` runs `validateWebhookURL` on the input. Five
rules:

1. Empty input is accepted (group without Discord).
2. URL must be parsable by `net/url.Parse`.
3. Scheme must be exactly `https` — `http://`, `ftp://`, schemeless
   URLs all rejected.
4. No embedded credentials (`User != nil`) — rejects
   `https://attacker:token@discord.com/...` which would route the
   notification to a different host while looking legitimate at a
   glance.
5. Non-empty `Host`.
6. Length ≤ 2048 characters.

The same length cap is enforced as a `CHECK` constraint on the
column (defense in depth: a direct SQL `INSERT` bypassing the
application is still bounded). The other rules are application-only:
the database accepts any short string, the application refuses to
materialize a Group entity with an invalid URL.

- **Rejected: also check the path matches `/api/webhooks/...`.**
  Pinning the URL prefix would break SMO if Discord ever changes
  their webhook URL scheme. The five rules above already prevent the
  most dangerous misconfigurations.
- **Rejected: a typed `WebhookURL` value object.** Single-use
  ceremony for one field. If Slack/Teams/etc. enter the picture
  later, extracting a value object then is straightforward.

### D4 — Masked in HTTP responses

`POST /api/v1/groups` accepts the URL in the request body in clear.
The response, and `GET /api/v1/groups/:id`, never echo the URL. The
DTO exposes `has_webhook bool` only.

A guard test — `TestGroupResponseFromEntity_NeverIncludesWebhookURL_InJSON`
— marshals the response and asserts the URL substring is absent. Any
future change that re-introduces the URL (intentional or accidental
copy-paste) will fail this test.

- **Rejected: hash-suffix display (e.g., last four chars).** A
  partial leak is still a leak. If an organizer needs to verify
  which URL is configured, they can issue a re-create or wait for
  the future PATCH endpoint that will require providing the full
  URL again to update.

### D5 — Cached in Redis alongside the rest of the Group

The `cachedGroup` struct in `infrastructure/cache/redis` includes the
URL. Skipping it would have required the notifier (PR #53b) to
bypass the cache and hit Postgres directly, defeating the cache.

The Redis instance shares the trust boundary with Postgres (same
compose network, same operator credentials). Caching the URL does
not widen the threat surface beyond what Postgres already exposes.

## Security considerations

The webhook URL is a **secret**: anyone in possession of it can post
arbitrary messages to the configured Discord channel. Treating it
properly matters even at SMO's scale.

### Threat model

- **Confidentiality**: the URL must not leak from the application to
  unauthorized parties.
- **Integrity**: the URL must reach the notifier as the organizer
  configured it, not modified in transit.
- **Availability**: notification failures must not block match
  transitions (handled in the notifier ADR).

### Why plain-text storage is acceptable

Adding column-level encryption to this single field would:

- Create an inconsistent posture (other potentially sensitive data
  in `organizers` and elsewhere remains unencrypted).
- Require key management (KMS, env-injected key, …) which SMO does
  not yet have.
- Give a false sense of security: a compromised application server
  reads the encryption key from the same env it reads the database
  password — both are equally exposed.

Plain-text storage, combined with the mitigations below, is a
defensible choice at this scale. A future PR can revisit with
encryption-at-rest **applied uniformly** when SMO adopts a KMS or
moves to a managed Postgres with built-in encryption.

### Mitigations in place

1. **Never logged.** No `slog.String("webhook_url", …)`, no
   `fmt.Sprintf` containing the URL, no error message that embeds
   the URL. Errors that surface from the notifier (PR #53b) wrap
   stdlib `*url.Error` to strip the URL before propagation.
2. **Never echoed in HTTP responses** (D4 + guard test).
3. **Database access is gated by Postgres credentials** which are
   themselves stored in env vars and never logged. The compose
   network exposes Postgres only to the application and an operator
   on `localhost:5433` (test fixture port) — not to the public.
4. **Length cap enforced both in the entity and at the DB level** to
   prevent oversized payloads from being sneaked in.
5. **Strict scheme/credentials checks** prevent classes of misuse
   where an attacker who controls the input could trick SMO into
   posting to a different host.

### Future-PR tracker

When SMO grows past hobby scale, revisit:

- **Column-level encryption** via pgcrypto + KMS-managed key. Right
  scope: separate ADR + dedicated PR that introduces the KMS and
  applies encryption to all secret columns at once (not just this
  one).
- **Audit log** of webhook URL changes (who/when) — once we have a
  multi-organizer model with privilege levels.
- **Webhook URL rotation reminder** — after N months without
  rotation, surface a UI prompt to refresh.

## Consequences

- A schema migration adds the column; mappers and HTTP DTOs are
  updated accordingly.
- The cache layer (PR #48) gains a field, requiring a version bump
  on cached entries. Existing cached entries become invalid until
  TTL expires (5 minutes) or the next write triggers
  invalidation — acceptable transient inconsistency.
- The notifier (PR #53b) can rely on `Group.WebhookURL()` as the
  single source of truth, without additional plumbing.
- Future PRs that touch the Group response DTO must keep the guard
  test green; reviewers should reject any change that adds the URL
  back to the response.
