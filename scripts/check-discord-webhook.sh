#!/usr/bin/env bash
#
# Pre-commit guard: refuse to commit a real-looking Discord webhook URL.
#
# A real Discord webhook URL has the shape
#   https://discord.com/api/webhooks/<id_15+_digits>/<token_50+_chars>
#
# Test fixtures in this repo deliberately use short fake tokens
# (e.g. "test-token", "abcdef-XYZ"), so the regex below requires a
# 15+ digit ID AND a 50+ char token portion before flagging — this
# matches Discord's real format (snowflake ID + base64-ish token of
# 60-80 chars) without breaking legitimate fixtures.
#
# GitGuardian also catches this server-side, but only after the push
# has been received. This hook stops the secret BEFORE the first
# `git push`, which is the only point at which a leak can still be
# fully prevented.
#
# Exits 1 if any staged file matches; 0 otherwise. Filenames are
# passed as arguments by pre-commit.
set -euo pipefail

PATTERN='discord\.com/api/webhooks/[0-9]{15,}/[A-Za-z0-9_-]{50,}'

if [ "$#" -eq 0 ]; then
  exit 0
fi

if matches=$(grep -nHE "$PATTERN" "$@" 2>/dev/null); then
  echo "ERROR: real-looking Discord webhook URL detected:" >&2
  echo "$matches" >&2
  echo >&2
  echo "Use a fake token in tests (short, non-base64 like 'test-token')." >&2
  echo "Never commit a real webhook URL — the token authenticates posts to your channel." >&2
  exit 1
fi

exit 0
