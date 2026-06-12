# Design patterns dans SMO — inventaire et justification

Chaque pattern ci-dessous est **réellement présent dans le code** (chemins de
fichiers vérifiés), résout un problème que ce projet rencontre vraiment, et
est relié à un test qui prouve le bénéfice. Aucun pattern n'a été ajouté pour
décorer : la règle d'architecture 4 de CLAUDE.md interdit les interfaces (et
donc la plupart des patterns) sans l'une des trois justifications —
testabilité, implémentations multiples, ou frontière de couche. Chaque entrée
précise quelle justification s'applique.

---

## 1. Repository

**Où**
- Ports (contrats) : `domain/ports/group_repository.go`,
  `match_repository.go`, `player_repository.go`, `invitation_repository.go`,
  `vote_repository.go`, `organizer_repository.go`
- Adaptateurs de production : `infrastructure/persistence/postgres/repositories/`
  (ex. `group_repository.go`, qui enveloppe les requêtes générées par sqlc)
- Implémentation en mémoire : `infrastructure/persistence/inmemory/group_repository.go`
- Fakes de test : un par package de use case, ex.
  `application/usecases/group/fake_repository_test.go`,
  `application/usecases/invitation/fake_repository_test.go`,
  `application/usecases/vote/fake_match_repository_test.go`

**Le problème résolu ici.** Les use cases doivent orchestrer la persistance
sans rien connaître de pgx, sqlc ou SQL. Sans le port, chaque test de use
case exigerait un Postgres démarré, et la couche unitaire de 70-80 % de la
pyramide de tests serait physiquement impossible (la suite unitaire doit
s'exécuter en moins d'une seconde d'après CLAUDE.md).

**Pourquoi ce pattern et pas une alternative.** L'alternative — des use cases
appelant directement les `Queries` générées par sqlc — a été rejetée parce
qu'elle couple la couche application aux structs générés (qui portent des
types pgx) et impose des containers aux tests unitaires. Une seconde
alternative, mocker l'interface `Querier` générée (sqlc en émet une, voir
`sqlc.yaml` `emit_interface: true`), a été rejetée parce qu'elle teste la
plomberie SQL au lieu de l'orchestration métier ; SMO teste plutôt
l'adaptateur contre un vrai Postgres via testcontainers.

**En quoi il facilite les tests (le lien explicite Repository -> fake InMemory).**
`fakeGroupRepository` dans `application/usecases/group/fake_repository_test.go`
est un struct d'environ 70 lignes adossé à une map, qui implémente le port
complet, protégé par mutex pour que les tests tournent avec `-race` et
`t.Parallel()`. Chaque test unitaire de use case (`create_group_test.go`,
`rename_group_test.go`, ...) s'exécute contre lui sans aucune E/S. Le même
port est aussi implémenté par l'adaptateur Postgres, vérifié séparément par
`infrastructure/persistence/postgres/repositories/
invitation_repository_integration_test.go` contre un vrai container — le
contrat est donc exercé aux deux niveaux de la pyramide.

**Note SOLID.** C'est l'inversion de dépendance rendue structurelle : la
politique de haut niveau (`application/`) dépend d'une abstraction déclarée
dans `domain/ports/`, et le détail de bas niveau (`infrastructure/`)
l'implémente. Le câblage n'a lieu que dans `cmd/server/main.go` (composition
root).

---

## 2. Strategy

**Où**
- Contrat : `domain/strategies/strategy.go` (interface `AssignmentStrategy`)
- Implémentations : `domain/strategies/random.go` (`RandomAssignmentStrategy`),
  `domain/strategies/ranking_based.go` (`RankingBasedStrategy`, snake draft +
  « le meilleur joueur rejoint le vainqueur précédent »),
  `domain/strategies/manual.go`
- Consommateur : `application/usecases/match/generate_teams.go`

**Le problème résolu ici.** L'affectation des équipes est le comportement
variable au cœur du produit : aléatoire pour le premier match d'un groupe,
fondée sur le classement dès qu'un historique de votes existe, manuelle quand
l'organisateur déplace les joueurs. Trois algorithmes réels derrière un seul
contrat — c'est la justification 2 de la règle 4 de CLAUDE.md (plusieurs
implémentations réelles), le cas d'école où Strategy mérite son interface.

**Pourquoi ce pattern et pas une alternative.** Une fonction unique avec un
switch sur enum a été rejetée : chaque algorithme a des besoins de
construction différents (`RandomAssignmentStrategy` exige un `*rand.Rand`
injecté, `RankingBasedStrategy` n'a besoin de rien) et des cas limites
différents, et le switch grossirait jusqu'à devenir exactement la
« substantial if chain in a use case » que la règle d'architecture 3
interdit.

**En quoi il facilite les tests.** Le déterminisme est inscrit dans le
contrat : `strategy.go` impose que l'aléa soit injecté, jamais global. Les
tests passent un `rand.New(rand.NewPCG(...))` à graine fixe, si bien que
`random_test.go` vérifie des compositions d'équipes exactes et que
`ranking_based_test.go` vérifie la sortie exacte du snake draft
(`TestRankingBasedStrategy_Assign_DistributesUsingSnakeDraft` déroule
l'appariement attendu en commentaire, puis le vérifie). Aucune instabilité
(flakiness) possible.

---

## 3. Observer (événements de domaine / publish-subscribe)

**Où**
- Événements : `domain/events/match_events.go` (`MatchTeamsReady`,
  `MatchTeamsReadyEventName = "match.teams_ready"`)
- Ports : `domain/ports/event_publisher.go` (`EventPublisher`,
  `EventSubscriber`)
- Adaptateur publisher : `infrastructure/events/inmemory/publisher.go`
  (synchrone, dispatch indexé par nom, isolation des erreurs par subscriber)
- Subscribers : `infrastructure/events/inmemory/logging_subscriber.go`
  (audit), `infrastructure/notifications/discord/subscriber.go`
  (notification webhook)
- Enregistrement : `cmd/server/main.go:236-242`
- Decision record : `docs/adr/0004-domain-events-pattern.md`

**Le problème résolu ici.** Quand un match passe à `teams_ready`, plusieurs
réactions sans lien entre elles doivent se déclencher : écrire une ligne
d'audit, poster sur Discord et (à terme) incrémenter des compteurs Prometheus
et alimenter la logique de verrouillage de compte. La conception naïve —
injecter un port `Notifier`, puis un port `MetricsRecorder`, puis un port
`AuditLogger` dans `MarkTeamsReadyUseCase` — ajoute un paramètre de
constructeur et un fake de test par préoccupation. L'ADR 0004 consigne que la
PR Discord (#55) proposait initialement la conception à port direct et a été
retravaillée vers les événements.

**Pourquoi ce pattern et pas une alternative.** L'ADR 0004 documente quatre
alternatives rejetées avec leurs raisons : port Notifier direct (ne passe pas
à l'échelle quand les préoccupations se multiplient), factory de notifiers
(YAGNI), publisher asynchrone à base de goroutines (inutile à cette échelle,
coût en ordonnancement et en débogage), et bus externe (totalement
surdimensionné pour un processus unique). Le dispatch se fait par **nom**
d'événement plutôt que par type Go ; l'ADR contient un tableau de compromis
explicite et la mitigation du seul risque réel (une constante d'événement
renommée n'est pas vérifiée par le compilateur).

**En quoi il facilite les tests.** Les tests de use case n'ont besoin que
d'un seul fake publisher adossé à une slice
(`application/usecases/match/fake_publisher_test.go`) au lieu d'un fake par
préoccupation. Le publisher lui-même est testé unitairement dans
`publisher_test.go` (routage, isolation des erreurs), et le subscriber
Discord est testé avec un fake `Notifier` dans `discord/subscriber_test.go` —
aucun serveur HTTP nécessaire grâce au pattern 5 ci-dessous.

---

## 4. Chain of Responsibility (chaîne de middlewares HTTP)

**Où** — assemblée dans `cmd/server/main.go:337-345`, chaque maillon dans
`infrastructure/http/middlewares/` :

1. `RequestID()` (`request_id.go`) — attribue/propage `X-Request-ID`
2. `SLogLogger(...)` (`logger.go`) — une ligne de log structuré par requête
3. `CORS(...)` (`cors.go`) — liste d'origines autorisées
4. `ratelimit.New(...).Middleware()` (`ratelimit/middleware.go`) — limites
   par route à fenêtre fixe, adossées à Redis
5. `gin.Recovery()` — le plus interne, pour que le logger au-dessus consigne
   aussi les panics
6. `JWTAuth(jwtSigner)` (`jwt_auth.go`) — appliqué uniquement au groupe de
   routes `protected` (`main.go:372-374`)

**Le problème résolu ici.** Les préoccupations HTTP transverses (corrélation,
logs, protection contre les abus, auth) doivent s'appliquer uniformément sans
que chaque handler les réimplémente. Chaque maillon décide indépendamment de
passer la main (`c.Next()`) ou d'interrompre (`c.AbortWithStatusJSON`), ce
qui est exactement le contrat de la chaîne.

**Pourquoi cet ordre (l'ordre est porteur de sens, documenté dans
`main.go:330-336`).** RequestID doit s'exécuter en premier pour que le logger
puisse lire l'ID depuis le contexte (`logger.go:24` énonce explicitement la
dépendance). Le rate limiter est placé après CORS pour que les requêtes
preflight ne soient pas limitées, et Recovery est le plus interne pour qu'un
panic produise quand même un 500 journalisé. JWTAuth s'applique par groupe de
routes, pas globalement, parce que les routes joueur authentifiées par token
(accepter une invitation, voter — ADR 0009) sont des mutations publiques par
conception.

**En quoi il facilite les tests.** Chaque maillon est testé isolément avec
`httptest` (`request_id_test.go`, `logger_test.go`, `jwt_auth_test.go`,
`ratelimit/middleware_test.go`) ; le rate limiter dispose en plus d'un test
d'intégration avec un container Redis (`middleware_integration_test.go`). Les
handlers sont testables sans aucun middleware parce que l'ID de requête et
l'ID d'organisateur circulent via des accesseurs `context.Context`
(`RequestIDFromContext`, `OrganizerIDFromContext`) qui se dégradent en
valeurs vides hors de la chaîne.

---

## 5. Adapter

**Où** — chaque technologie externe est encapsulée derrière un port du
domaine (justification 3 de la règle 4 de CLAUDE.md : frontière de couche) :

| Port (domaine) | Adaptateur (infrastructure) | Technologie externe masquée |
|---|---|---|
| `ports.InvitationTokenService` | `infrastructure/token/service.go` | crypto/rand + SHA-256 |
| `ports.PasswordHasher` | `infrastructure/auth/bcrypt/hasher.go` | bcrypt |
| `ports.JWTSigner` | `infrastructure/auth/jwt/signer.go` | golang-jwt |
| `ports.Clock` | `infrastructure/clock/` | `time.Now()` |
| `ports.IDGenerator` | `infrastructure/idgen/` | uuid |
| `ports.LoginAttemptTracker` | `infrastructure/auth/loginlockout/redis_tracker.go` (+ `noop_tracker.go`) | Redis |
| `discord.Notifier` | `discord.HTTPNotifier` (`infrastructure/notifications/discord/notifier.go`) | API HTTP des webhooks Discord |
| `handlers.DBPinger` / `CachePinger` | `*pgxpool.Pool` nativement / `cacheredis.NewPinger` | pgx / go-redis |

**Le problème résolu ici.** Deux problèmes distincts. (a) Déterminisme : les
adaptateurs `Clock` et `IDGenerator` rendent chaque timestamp et chaque ID
injectables, si bien que les tests de use case utilisent
`fake_clock_test.go` / `fake_id_generator_test.go` et vérifient des valeurs
exactes. (b) Confinement des fuites externes : `DiscordNotifier` existe
précisément pour empêcher l'URL du webhook (un secret) de s'échapper —
`notifier.go:96-110` abandonne délibérément les erreurs `net/http`
sous-jacentes parce qu'elles contiennent l'URL en clair. Ce comportement de
sécurité vit à un seul endroit parce que l'adaptateur est le seul code à
toucher l'API brute.

**Pourquoi ce pattern et pas une alternative.** Appeler
`time.Now()`/`bcrypt`/`http.Post` en ligne ferait moins de code, mais
CLAUDE.md interdit purement et simplement `time.Now()` dans le code
domaine/application (règles de test : "Time must be injected through a Clock
port"), et le contrat de hachage des tokens
(`domain/ports/invitation_token_service.go`) est une vraie règle de domaine
(seul le hash est persisté) qui mérite une couture nommée. Noter ce qui n'a
*pas* été encapsulé : Gin et pgx ne sont pas derrière des abstractions maison
à l'intérieur de leur propre couche — les adaptateurs n'existent qu'à la
frontière du domaine, conformément au « no interface without justification »
de la règle 4.

**En quoi il facilite les tests.** `application/usecases/invitation/
fake_token_service_test.go` renvoie des tokens prévisibles, si bien que
`create_invitation_test.go` vérifie le hash exact persisté ;
`discord/subscriber_test.go` injecte un fake `Notifier` qui capture les
payloads ; les interfaces côté consommateur `DBPinger`/`CachePinger` du
handler de santé (`health_handler.go:21-35`) permettent à
`health_handler_test.go` de simuler une base injoignable sans aucune base de
données.

---

## 6. Decorator (repositories en cache-aside)

**Où**
- Décorateurs : `infrastructure/cache/redis/group_repository.go`
  (`CachingGroupRepository`), `player_repository.go`
  (`CachingPlayerRepository`)
- Enveloppement conditionnel : `infrastructure/cache/redis/wiring.go`
  (`WrapGroupRepository` / `WrapPlayerRepository` renvoient le repository
  interne inchangé quand Redis est désactivé)
- Appliqué à la composition root : `cmd/server/main.go:219-221`
- Decision record : `docs/adr/0002-cache-aside-with-redis.md`

**Le problème résolu ici.** `GroupRepository.FindByID` et
`PlayerRepository.ListByGroup` sont les chemins de lecture les plus
sollicités (chaque page groupe/match appelle les deux, et le frontend
interroge en boucle). Le cache devait être ajouté **sans toucher un seul use
case** — les décorateurs implémentent les mêmes ports du domaine que ce
qu'ils enveloppent, donc les appelants ne voient aucune différence. Les
lectures tentent Redis d'abord et retombent sur la base en cas de miss ou
d'erreur ; les écritures délèguent d'abord, puis invalident (l'ordre
base-d'abord est délibéré : un échec de suppression dans le cache laisse
Postgres faire autorité — ADR 0002).

**Pourquoi ce pattern et pas une alternative.** L'ADR 0002 rejette
explicitement deux alternatives : un port `Cache` injecté dans les use cases
(couple l'orchestration au cache, impose un fake de cache dans chaque test de
use case) et un cache en middleware HTTP (trop grossier — il ne peut pas
savoir qu'une écriture sur le groupe X doit invalider la lecture du groupe
X). Il consigne aussi la décision de dépendance souple : les erreurs Redis se
dégradent vers Postgres avec un WARN, jamais un 5xx.

**En quoi il facilite les tests (et prouve le bénéfice).** La décoration
étant transparente, tous les tests de use case existants ont continué de
passer sans modification à l'arrivée du cache — cette absence de churn est le
bénéfice mesurable. Le décorateur lui-même est testé en intégration contre un
vrai Redis :
`player_repository_integration_test.go::TestCachingPlayerRepository_UpdateRanking_InvalidatesGroupListCache`
prouve que la classe de bugs d'obsolescence du cache est couverte, et
`group_repository_integration_test.go::TestCachingGroupRepository_FallsThroughToInner_WhenRedisGoesDown`
prouve la dégradation gracieuse. Le passage direct quand le client est nil
est testé unitairement dans `wiring_test.go`.

---

## Mentions honorables (réelles, mais non revendiquées comme patterns vedettes)

- **Composition Root / DI manuelle** — `cmd/server/main.go:215-392` est
  l'unique endroit où les implémentations concrètes se rencontrent (CLAUDE.md
  impose l'absence de framework de DI).
- **DTO + Mapper** — `infrastructure/http/dto/` et
  `infrastructure/persistence/postgres/mappers/` tiennent les tags JSON/SQL à
  l'écart des entités du domaine (le compromis de pureté sqlc, voir
  [architecture.md](architecture.md)).
- **Null Object** — `loginlockout.NewNoopTracker()` remplace le tracker Redis
  quand Redis est absent (`main.go:301-308`), avec un WARN explicite pour que
  la dégradation soit visible.
- **Capability token** — les invitations comme capabilities pour le RSVP *et*
  le vote (ADR 0009) — une décision de conception sécurité plutôt qu'un
  pattern du GoF, mais celle qui mérite le plus d'être expliquée en
  soutenance.

> **MANQUE :** aucun pour ce critère — les six patterns existent avec leurs
> tests. Le seul suivi est de garder cet inventaire à jour si les use cases
> `sharelink` (branche en cours) introduisent de nouvelles coutures.
