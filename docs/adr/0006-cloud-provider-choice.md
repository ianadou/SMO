# ADR 0006 — Cloud provider for production hosting

**Status:** Accepted (2026-04-30) — supersedes the prior intent to host on Hetzner Cloud CAX11

## Context

SMO targets a French audience under the `sportpotes.fr` domain (purchase
scheduled 2026-05-01). The 100-organizer target is small but the project
must run reliably as a long-lived portfolio piece, not just for the
3-day school deliverable. The stack is Docker Compose-based locally
(Go backend, Postgres, Redis, Nuxt SPA, Nginx, Prometheus + Grafana)
and Docker images are already built multi-arch (`linux/amd64` +
`linux/arm64`).

Three constraints drive the provider choice:

1. **Finish fast.** The author explicitly prioritizes velocity to free
   up time for parallel projects. Anything that adds 1-2 weeks of
   provisioning friction or forces a stack refactor is disqualifying.
2. **Budget up to €40/month is acceptable.** The decision should not be
   optimized for raw cost when speed and reliability are at stake.
3. **Cloud with a logical justification, not for academic display.**
   The chosen architecture must be defensible on its own merits and
   not chosen to pattern-match a course rubric.

A prior ADR-less decision had pinned Hetzner Cloud CAX11. This ADR
revisits that choice once the three constraints above were made
explicit and finalizes the production target.

## Decision

**Provider:** OVHcloud
**Plan:** B2-7 (2 vCPU dedicated, 7 GB RAM, 50 GB SSD NVMe)
**Region:** Gravelines (`GRA`) or Strasbourg (`SBG`) — to be pinned at
provisioning time, do not silently default
**Budget:** ~€18/month, well within the €40 ceiling
**TLS:** Let's Encrypt via certbot wired into Nginx on the VM
**No CDN/proxy** in front (Cloudflare rejected previously, see hosting
memory)

The same VM hosts the entire Compose stack (app + Postgres + Redis +
Prometheus + Grafana + Nginx) at MVP. Vertical scaling via OVH plan
upgrade is the path forward if 100 organizers grows materially.

## Rationale

### Three reasons OVHcloud wins under the stated constraints

1. **Brand and legal alignment.** A `.fr` domain selling to a French
   audience hosted on a French cloud provider eliminates the Schrems II
   / Cloud Act conversation entirely. Data residency, RGPD posture, and
   billing are all in-jurisdiction. This is a substantive technical
   reason, not a flag-waving one — every cross-border transfer otherwise
   needs a documented basis.
2. **Zero refactor of the deployment model.** The Compose stack we
   already run locally maps 1:1 onto a single VM. Terraform + Ansible
   already target the "provision Linux + run Compose + reverse proxy"
   pattern. No new infrastructure paradigm to learn before shipping.
3. **Vertical headroom within budget.** B2-7 has comfortable margin for
   the 100-organizer target with the full observability sidecars
   running on the same box. If growth pushes that limit, B2-15 (4 vCPU,
   15 GB) at ~€32/mo still fits the €40 ceiling — same provider, same
   API, no migration.

### Why the existing memory's "Hetzner CAX11" was reconsidered

The previous decision was made under an implicit "minimize cost"
constraint that does not actually hold once budget tolerance is set
to €40/month. With cost no longer the binding constraint, brand
alignment and dev/prod mental-model parity become the dominant
factors, both of which point to OVHcloud over a German-based provider.

## Alternatives considered and rejected

### Hetzner Cloud CAX21 (~€10/month)

Best raw price-per-performance in the EU IaaS market, with native ARM
Ampere matching our multi-arch CI pipeline. Rejected because:

- Datacenter is in Germany (Falkenstein/Nuremberg), which weakens
  the FR-product / FR-host narrative for marginal savings
  (€8/month difference) that do not matter under our budget.
- The "ARM matches the multi-arch pipeline" argument cuts both ways:
  multi-arch images run anywhere, so it is not a real differentiator.

Hetzner remains the obvious fallback if OVHcloud experiences extended
unavailability — the deployment scripts are portable.

### Oracle Cloud Infrastructure ARM Always Free

Tempting on paper: 4 OCPU + 24 GB RAM + 200 GB storage at zero cost,
forever. Rejected on three grounds:

- **Capacity gamble.** ARM A1 instances in EU regions are routinely
  "Out of capacity" for the free tier. Reports of 1-3 weeks of retry
  loops to get an instance allocated are common. This directly
  contradicts the "finish fast" constraint.
- **Account fragility.** Oracle has a documented history of reclaiming
  free-tier resources for inactivity, anti-fraud false positives, or
  payment-method anomalies. For a portfolio piece intended to stay
  live 2-3 years, this is a non-trivial risk.
- **Stack mismatch.** Even if capacity were instant, Oracle Cloud's
  console UX and IAM model would slow delivery compared to a
  straightforward IaaS VM.

A "start on Oracle, migrate later" plan was specifically considered
and rejected: it would force the same provisioning + DNS + TLS +
monitoring setup work to be done twice, save €18/month for the few
months before migration, and break the brand alignment story during
phase 1. Net negative on every constraint that matters.

### Scaleway PRO2-XS (~€16/month)

Also French, also RGPD-aligned, with a developer experience often
considered more modern than OVH's. Rejected because:

- For a portfolio + academic context, OVHcloud carries more
  institutional weight in France — a marginal but real signal.
- No technical advantage that would justify giving up that signal.

Scaleway remains a viable fallback if OVH becomes problematic.

### Serverless: Cloud Run + Neon + Cloudflare Pages

Architecturally appealing for a bursty workload like SMO (idle 95%
of the time): scale-to-zero compute, serverless Postgres, static SPA
on a global CDN, all within free tiers. Rejected because:

- **Refactor cost.** The current stack is Compose-based with explicit
  Postgres + Redis + Prometheus + Grafana sidecars. Moving to Cloud
  Run forces a rewrite of deployment, breaks the dev/prod parity
  Compose currently provides, and fragments observability across
  three providers.
- **Cold starts.** Cloud Run + Neon idle resume adds 500ms-2s on the
  first request after inactivity. Acceptable for the product but
  contradicts the "optimized" goal stated by the user.
- **Speed of delivery.** Learning the multi-provider operational model
  costs 1-2 weeks before any feature work resumes. Disqualifying
  under the "finish fast" constraint.

Worth revisiting in a future ADR if SMO traffic profile changes
materially or if the project is rebuilt around event-driven workloads.

### AWS / GCP / Azure direct

Rejected without deep evaluation: free tiers are time-limited or
geographically restrictive (12-month AWS, US-only GCP e2-micro free),
and the long-term cost for an equivalent-spec VM exceeds the OVHcloud
B2-7 by 2-3× without any technical advantage relevant to SMO's profile.

## Consequences

- The hosting memory (`project_hosting_target.md`) is updated to
  reflect OVHcloud B2-7 as the new target.
- Terraform code under `deploy/terraform/` will use the
  `ovh/ovh` provider for OVH-side resource management; the existing
  `kreuzwerker/docker` provider continues to handle VM-internal
  Docker provisioning unchanged.
- Ansible inventory pins the OVH instance hostname; playbooks that
  may have been drafted assuming Hetzner-style cloud-init must be
  audited before first deploy. Open as a follow-up task at deploy
  time.
- The Dockerfile multi-arch build is preserved — both `linux/amd64`
  (B2-7 default) and `linux/arm64` (potential future migration to
  Hetzner ARM or OVH Public Cloud ARM) remain supported.
- TLS strategy is unchanged: Let's Encrypt + Nginx on the VM, no CDN
  layer.
- Backup plan if OVHcloud becomes unavailable: Scaleway PRO2-XS in
  Paris (`PAR1`) — same FR sovereignty story, different provider,
  Terraform/Ansible portable.
