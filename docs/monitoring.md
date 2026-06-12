# Monitoring et observabilité — état actuel

Résumé honnête : **les sondes de santé et le logging structuré sont solides
et entièrement implémentés ; les métriques (Prometheus/Grafana) et
l'alerting ne sont pas encore implémentés.** Ce fichier documente ce qui
existe avec des preuves, puis les manques avec une remédiation concrète.

## Endpoints de santé — implémentés

Fichier : `infrastructure/http/handlers/health_handler.go`. Deux endpoints,
délibérément séparés, avec la justification écrite dans le code
(lignes 37-44) :

| Endpoint | Comportement | Consommateur |
|---|---|---|
| `GET /health/live` | Toujours 200 `{status: "alive", timestamp}` — "as long as the Go runtime can answer, the process is alive" | La sonde liveness d'un futur orchestrateur ; n'échoue jamais pendant les incidents transitoires des backends, pour que le container ne soit pas tué inutilement |
| `GET /health/ready` | `{status, database, cache, timestamp}` ; **503** quand Postgres est injoignable ; **200 + statut "degraded"** quand seul Redis est tombé | HEALTHCHECK Docker (via `cmd/healthcheck`), Dockhand, futur load balancer |

Cela satisfait l'exigence de la grille d'« un endpoint de santé retournant
l'état de l'app et de ses dépendances ». Décisions de conception qui méritent
d'être défendues :

- **La base de données est une dépendance dure, le cache une dépendance
  souple** (lignes 78-101) : une panne de Redis conserve un HTTP 200 parce
  que l'app sert toujours des résultats corrects depuis Postgres (le contrat
  de dépendance souple de l'ADR 0002) ; seul le corps dit la vérité
  (`cache: "down"`, `status: "degraded"`). Une panne de la base bascule en
  503 pour qu'un load balancer puisse drainer l'instance.
- **Sondes bornées** : ping de la base plafonné à 2s, ping du cache à 1s
  (lignes 11-14) — l'endpoint de readiness ne peut jamais rester bloqué.
- **Interfaces côté consommateur** `DBPinger`/`CachePinger` (lignes 21-35) :
  elles gardent le handler exempt d'imports pgx/redis et rendent le scénario
  base-injoignable testable unitairement avec des stubs
  (`health_handler_test.go`).
- **Honnêteté du cache désactivé** : quand `REDIS_URL` n'est pas défini, la
  sonde rapporte `cache: "disabled"` au lieu de prétendre qu'il fonctionne
  (lignes 50-56, câblé dans `cmd/server/main.go:350-358`).
- Les endpoints vivent à la racine, pas sous `/api/v1` — contrat
  d'infrastructure appartenant à l'ops, pas au produit (ADR 0001 ;
  commentaire aux lignes 58-63 du handler).
- Le HEALTHCHECK Docker cible `/health/ready`, pas `/live`, pour que le
  container passe unhealthy en cas de perte de la base
  (`cmd/healthcheck/main.go:1-7`).

## Logging structuré — implémenté

- **JSON dès la première ligne** : `slog.NewJSONHandler` défini par défaut
  dans `cmd/server/main.go:95-96`. Aucun log par concaténation de chaînes
  nulle part ; les événements sont des messages stables + des champs typés,
  ex. la ligne par requête `"request completed"` avec
  `method, path, status, duration_ms, request_id, remote_ip`
  (`infrastructure/http/middlewares/logger.go:44-54`).
- **Les niveaux portent du sens** : le logger de requêtes mappe 5xx -> ERROR,
  tout le reste en INFO (`logger.go:38-42`) ; la dégradation Redis loggue en
  WARN, throttlé à un par minute pour qu'une panne ne puisse pas inonder le
  flux (`ratelimit/middleware.go:21-24,117-127`) ; les jalons de démarrage
  sont en INFO ; le repli de verrouillage sans Redis est un WARN à portée
  opérationnelle (`main.go:307`).
- **ID de corrélation de bout en bout** : le middleware `RequestID`
  (`middlewares/request_id.go`) accepte un `X-Request-ID` entrant UUIDv4
  valide ou en génère un (en signalant la génération via
  `X-Request-ID-Generated`), le stocke dans `context.Context` avec une clé
  typée, et chaque couche en aval — y compris les subscribers d'événements de
  domaine — en hérite via `slog.*Context(ctx, ...)` (ADR 0004, section
  Logging strategy). Le middleware de log est enregistré après RequestID par
  contrat documenté (`logger.go:23-24`).
- **Contrôle du bruit** : les requêtes vers `/health/live` et
  `/health/ready` sont exclues du log de requêtes parce que Docker les
  sollicite toutes les 10 secondes (`logger.go:10-17`) — les logs restent de
  la forensique, pas du bruit.
- **Aucune donnée sensible, par ingénierie délibérée** :
  - L'URL du webhook Discord (un secret) n'atteint jamais les erreurs ni les
    logs ; le notifier abandonne les erreurs `net/http` sous-jacentes
    précisément parce qu'elles contiennent l'URL en clair
    (`infrastructure/notifications/discord/notifier.go:96-110`, doc de
    package lignes 7-14, ADR 0003 mitigation 1).
  - Les tokens d'invitation en clair ne sont jamais loggués ni persistés —
    seul le hash SHA-256 est stocké ; le contrat du port indique "callers
    must treat the return value as a secret and never log or persist it"
    (`domain/ports/invitation_token_service.go`).
  - Les erreurs du client Redis sont expurgées avant journalisation pour que
    la configuration de connexion ne puisse pas fuir
    (`ratelimit/middleware.go:128-140`).
  - Les mots de passe n'existent que sous forme de hashs bcrypt
    (`infrastructure/auth/bcrypt/`) ; aucun appel de log ne les touche.

## Prometheus + Grafana — NON implémentés

Preuves de l'absence (énoncées pour que le rendu ne puisse pas être accusé
de rester dans le vague) :

- `compose.yml` ne définit que `postgres`, `redis`, `app`, `frontend` — aucun
  service prometheus ou grafana.
- `go.mod` ne contient aucune dépendance `prometheus/client_golang` ; aucun
  endpoint `/metrics` n'est enregistré dans `cmd/server/main.go`.
- Les points d'ancrage existent à dessein : l'ADR 0004 cite "Prometheus
  counters and histograms on every state transition" comme subscriber
  d'événements planifié (`MetricsSubscriber`), donc l'ajout des métriques ne
  touchera aucun use case.

> **MANQUE (Prometheus/Grafana, ~1 pt) :** chemin de remédiation, de petite
> taille parce que l'architecture l'a anticipé : (1) ajouter
> `prometheus/client_golang`, exposer `/metrics` à côté des endpoints de
> santé (au niveau racine, même justification ops) ; (2) un histogramme en
> middleware HTTP (method/path/status/duration) dans la chaîne existante +
> un `MetricsSubscriber` sur `match.teams_ready` (la conception de
> l'ADR 0004) ; (3) ajouter les services `prom/prometheus` +
> `grafana/grafana` à `compose.yml` avec une config de scrape et un dashboard
> JSON à 5 panneaux (débit de requêtes, latence p95, taux de 5xx, compteur de
> transitions de match, statut de /health/ready) ; (4) des captures d'écran
> pour le rendu — la grille accepte explicitement une stack locale avec
> captures d'écran.

> **MANQUE (cohérence) :** CLAUDE.md et l'ADR 0006 décrivent la stack compose
> comme incluant déjà Prometheus + Grafana. Corriger la formulation ou livrer
> les services ; ne pas laisser les documents de soutenance contredire
> `compose.yml`.

## Alerting — NON implémenté (stratégie à documenter)

Aucune règle d'alerte Grafana, pas d'Alertmanager. La grille accepte au
minimum « une stratégie d'alerting justifiée » ; la stratégie actuelle,
honnêtement, est :

- **Au niveau des containers** : chaque container a un HEALTHCHECK (le
  backend sonde `/health/ready`, le frontend sonde `/`, postgres
  `pg_isready`, redis `redis-cli ping`) ; Dockhand (l'UI de supervision
  locale documentée dans CLAUDE.md) affiche la santé en rouge/vert et l'usage
  des ressources. Un container qui passe unhealthy est visible en ~30s
  (intervalle 10s, 3 retries).
- **Au niveau des logs** : ERROR est réservé à « cassé, attention humaine
  requise » (le logger de requêtes le mappe strictement sur les 5xx), donc
  grepper ERROR dans le flux JSON est un signal qui a du sens, pas du bruit.

> **MANQUE (alerting, ~0,5 pt) :** une fois la stack Prometheus ci-dessus en
> place, ajouter une seule alerte Grafana : *taux de 5xx > 1 % sur
> 5 minutes* (la métrique existe dans l'histogramme HTTP planifié) avec un
> point de contact webhook Discord — en réutilisant exactement le transport
> de notification que le produit livre déjà
> (`infrastructure/notifications/discord/`), ce qui rend le chemin d'alerte
> testable avec la même technique de fake notifier.

## Quoi montrer pendant la soutenance

1. `docker compose up -d`, puis `curl localhost:8081/health/ready` — 200 avec
   `database: ok, cache: ok`.
2. `docker compose stop redis`, re-curl — 200, `status: degraded,
   cache: down` (dépendance souple, ADR 0002).
3. `docker compose stop postgres`, re-curl — 503 ; `docker ps` montre le
   container de l'app basculant à *unhealthy* (le binaire de healthcheck à
   l'œuvre).
4. `docker compose logs app` — des lignes JSON avec `request_id`, le mapping
   des niveaux, et aucun token/webhook/mot de passe nulle part.
