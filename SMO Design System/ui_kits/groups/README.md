# SMO Groups — UI Kit

Page d'accueil de l'organisateur connecté : liste des groupes qu'il administre.

## Files

| File | What |
|---|---|
| `index.html` | Live screen on top, 4-state gallery below (multi · single · empty · loading) |
| `groups.css` | Component styles — uses tokens from `../../colors_and_type.css` only |
| `Icons.jsx` | Lucide-shaped icons (`circle-user`, `plus`, `users-round`, `chevron-right`, `map-pin`, `calendar`) |
| `AvatarCluster.jsx` | Stacked initials, max 4 + "+N" overflow pill |
| `GroupCard.jsx` | Full-width card: name + chevron, avatar cluster + count, next match, last result |
| `GroupsParts.jsx` | Header, page title, FAB, empty state, skeleton, sort helper |

## Sort rule

Groups with a planned `nextMatch` come first (soonest by `sortKey`), then groups with none — locked in `sortGroups()`.

## Production parity

Same approach as the login kit: tokens → `tailwind.config.ts`, JSX → Vue SFCs, swap our local `Icons.jsx` for `lucide-vue-next`. The Wordmark is shared with the login kit (`ui_kits/login/Wordmark.jsx`).

## Out of scope (stubbed)

- Tap on a card → group detail page
- Tap on FAB / empty-state CTA → "create group" modal
- Tap on profile icon → profile menu / logout
