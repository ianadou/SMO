# SMO Group Detail — UI Kit

Page de gestion d'un groupe : 3 onglets (Vue d'ensemble · Joueurs · Matchs).

## Files
- `index.html` — live screen (Vue d'ensemble) + galerie des onglets et état vide
- `Icons.jsx` — icônes Lucide-style locales
- `Overview.jsx` — Header, Tabs, Avatar, RankBar, OverviewTab (next match + stats + top 3)
- `Players.jsx` — formulaire ajout + liste joueurs triée par ranking décroissant
- `Matches.jsx` — bouton créer + cartes match avec badges/résultats + état vide
- `group_detail.css` — styles, tokens uniquement

## Notes
- Aucune nouvelle couleur — badge "En cours" utilise `--warn` (jaune signal) comme prévu dans le DS.
- Stats grid passe en colonne unique sous 420px.
- Le formulaire d'ajout joueur passe en stack vertical sous 420px.
