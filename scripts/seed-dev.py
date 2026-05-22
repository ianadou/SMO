#!/usr/bin/env python3
"""Seed the local dev DB with realistic data for visual review.

Bypasses rate limits where possible by reusing one organizer session
and chaining API calls. Idempotent on re-run only via a wipe — run
`docker compose down -v && docker compose up -d` first if you want
a clean state.

Usage:
    pnpm exec ./scripts/seed-dev.py
    # or
    python3 scripts/seed-dev.py
"""

from __future__ import annotations

import json
import subprocess
import sys
import time
import urllib.error
import urllib.request
from datetime import datetime, timedelta, timezone

API_BASE = "http://localhost:8081/api/v1"

ORGANIZER = {
    "email": "demo@smo.local",
    "password": "DemoPass1234!",
    "display_name": "Demo Organizer",
}

ROSTER = [
    "Alex L.",
    "Inès R.",
    "Théo B.",
    "Marc R.",
    "Paul S.",
    "Issa K.",
    "Yann N.",
    "Cédric D.",
    "Sami F.",
    "Greg P.",
    "Lola D.",
    "Marin N.",
]


def http(method: str, path: str, body=None, token: str | None = None):
    url = f"{API_BASE}{path}"
    data = json.dumps(body).encode() if body is not None else None
    headers = {"Accept": "application/json"}
    if body is not None:
        headers["Content-Type"] = "application/json"
    if token:
        headers["Authorization"] = f"Bearer {token}"
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            raw = resp.read().decode()
            return resp.status, json.loads(raw) if raw else None
    except urllib.error.HTTPError as e:
        raw = e.read().decode()
        try:
            return e.code, json.loads(raw)
        except json.JSONDecodeError:
            return e.code, raw


def login_or_register() -> str:
    status, body = http("POST", "/auth/login", {
        "email": ORGANIZER["email"],
        "password": ORGANIZER["password"],
    })
    if status == 200 and isinstance(body, dict) and body.get("token"):
        return body["token"]
    print(f"[seed] login failed ({status}), trying register…", file=sys.stderr)
    status, body = http("POST", "/auth/register", ORGANIZER)
    if status not in (200, 201):
        sys.exit(f"register failed: {status} {body}")
    status, body = http("POST", "/auth/login", {
        "email": ORGANIZER["email"],
        "password": ORGANIZER["password"],
    })
    if status != 200:
        sys.exit(f"login after register failed: {status} {body}")
    return body["token"]


def create_group(token: str, name: str) -> str:
    status, body = http("POST", "/groups", {"name": name}, token)
    if status not in (200, 201):
        sys.exit(f"create group failed: {status} {body}")
    return body["id"]


def create_players(token: str, group_id: str, names: list[str]) -> list[dict]:
    out = []
    for name in names:
        status, body = http("POST", "/players", {
            "group_id": group_id, "name": name,
        }, token)
        if status not in (200, 201):
            sys.exit(f"create player {name!r} failed: {status} {body}")
        out.append(body)
    return out


def create_match(token: str, group_id: str, title: str, venue: str, when: datetime) -> dict:
    status, body = http("POST", "/matches", {
        "group_id": group_id,
        "title": title,
        "venue": venue,
        "scheduled_at": when.isoformat().replace("+00:00", "Z"),
    }, token)
    if status not in (200, 201):
        sys.exit(f"create match {title!r} failed: {status} {body}")
    return body


def flush_rate_limit() -> None:
    subprocess.run(["docker", "exec", "smo-redis", "redis-cli", "FLUSHDB"],
                   check=False, capture_output=True)


def invite_and_confirm(token: str, match_id: str, player_id: str) -> None:
    status, body = http("POST", "/invitations", {
        "match_id": match_id, "player_id": player_id,
    }, token)
    if status not in (200, 201):
        sys.exit(f"create invitation failed: {status} {body}")
    plain = body["plain_token"]
    for attempt in range(6):
        status, body = http("POST", "/invitations/respond", {
            "token": plain, "answer": "yes",
        })
        if status in (200, 201):
            return
        if status == 429:
            time.sleep(1.0)
            continue
        sys.exit(f"respond invitation failed: {status} {body}")
    sys.exit(f"respond invitation rate-limited too long")


def open_match(token: str, match_id: str) -> None:
    status, body = http("POST", f"/matches/{match_id}/open", {}, token)
    if status not in (200, 201):
        sys.exit(f"open match failed: {status} {body}")


def generate_teams(token: str, match_id: str, strategy: str = "random") -> None:
    status, body = http("POST", f"/matches/{match_id}/teams/generate", {
        "strategy": strategy,
    }, token)
    if status not in (200, 201):
        sys.exit(f"generate teams failed: {status} {body}")


def mark_ready(token: str, match_id: str) -> None:
    status, body = http("POST", f"/matches/{match_id}/teams-ready", {}, token)
    if status not in (200, 201):
        sys.exit(f"teams-ready failed: {status} {body}")


def start_match(token: str, match_id: str) -> None:
    status, body = http("POST", f"/matches/{match_id}/start", {}, token)
    if status not in (200, 201):
        sys.exit(f"start match failed: {status} {body}")


def complete_match(token: str, match_id: str, score_a: int, score_b: int) -> None:
    status, body = http("POST", f"/matches/{match_id}/complete", {
        "score_a": score_a, "score_b": score_b,
    }, token)
    if status not in (200, 201):
        sys.exit(f"complete match failed: {status} {body}")


def main() -> None:
    token = login_or_register()
    print(f"[seed] organizer logged in as {ORGANIZER['email']}")

    group_id = create_group(token, "Foot du jeudi")
    print(f"[seed] group created: {group_id}")

    players = create_players(token, group_id, ROSTER)
    print(f"[seed] {len(players)} players created")

    now = datetime.now(timezone.utc).replace(microsecond=0)
    thursday = (now + timedelta(days=(3 - now.weekday()) % 7 or 7)).replace(
        hour=19, minute=30, second=0,
    )

    # Match A — draft, no participants
    match_a = create_match(token, group_id, "matche A (brouillon)",
                           "Salle Pierre Mendès", thursday + timedelta(days=21))
    print(f"[seed] draft match: {match_a['id']}")

    # Match B — open with auto-generated teams (only 10 confirmed)
    match_b = create_match(token, group_id, "matche B (composition)",
                           "Salle Pierre Mendès", thursday + timedelta(days=14))
    flush_rate_limit()
    for p in players[:10]:
        invite_and_confirm(token, match_b["id"], p["id"])
    open_match(token, match_b["id"])
    generate_teams(token, match_b["id"], "random")
    print(f"[seed] composition match: {match_b['id']}")

    # Match C — completed with score (équipe rouge wins)
    match_c = create_match(token, group_id, "matche C (terminé)",
                           "Salle Pierre Mendès", thursday + timedelta(days=7))
    flush_rate_limit()
    for p in players[:10]:
        invite_and_confirm(token, match_c["id"], p["id"])
    open_match(token, match_c["id"])
    generate_teams(token, match_c["id"], "random")
    mark_ready(token, match_c["id"])
    start_match(token, match_c["id"])
    complete_match(token, match_c["id"], 5, 3)
    print(f"[seed] completed match: {match_c['id']}")

    print("\n[seed] visual URLs:")
    print(f"  groups list   : http://localhost:3000/groups")
    print(f"  group detail  : http://localhost:3000/groups/{group_id}")
    print(f"  match draft   : http://localhost:3000/matches/{match_a['id']}")
    print(f"  match compo   : http://localhost:3000/matches/{match_b['id']}")
    print(f"  match finished: http://localhost:3000/matches/{match_c['id']}")
    print(f"\n[seed] credentials: {ORGANIZER['email']} / {ORGANIZER['password']}")


if __name__ == "__main__":
    main()
