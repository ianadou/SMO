#!/usr/bin/env bash
#
# Pre-commit guard: refuse to commit a Google Fonts CDN URL inside `frontend/`.
#
# Loading https://fonts.googleapis.com/... or https://fonts.gstatic.com/...
# at runtime transfers the visitor's IP to Google. CNIL (and German
# jurisprudence) classify this as an unconsented data transfer for a
# `.fr` site like sportpotes.fr — non-compliant.
#
# Production fonts are self-hosted via @nuxt/fonts, which downloads the
# files at build time and serves them from the app origin (see ADR 0005).
# This hook prevents an accidental <link rel="stylesheet" href="...">
# or @import that would bypass the self-hosting setup.
#
# Scoped to `frontend/` via the pre-commit `files:` filter — the static
# design system mocks under `SMO Design System/` are intentionally
# allowed to use the CDN form and are out of scope.
#
# Exits 1 if any staged file matches; 0 otherwise.
set -euo pipefail

PATTERN='fonts\.(googleapis|gstatic)\.com'

if [ "$#" -eq 0 ]; then
  exit 0
fi

if matches=$(grep -nHE "$PATTERN" "$@" 2>/dev/null); then
  echo "ERROR: Google Fonts CDN URL detected under frontend/:" >&2
  echo "$matches" >&2
  echo >&2
  echo "Production fonts are self-hosted via @nuxt/fonts (see ADR 0005)." >&2
  echo "Loading from fonts.googleapis.com / fonts.gstatic.com leaks visitor IPs to Google (RGPD)." >&2
  exit 1
fi

exit 0
