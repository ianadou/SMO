# Contraintes et solutions de contournement — ce qui n'a pas pu être fait, et ce qui le remplace

La règle spéciale de la grille de notation : toute exigence impossible (outil
payant, pas de serveur, pas de domaine) doit être documentée — ce qui était
demandé, pourquoi c'est irréalisable maintenant, ce qui aurait été utilisé,
et le contournement choisi. Ce fichier est cette section. Chaque entrée suit
cette forme en quatre volets.

## 1. Serveur de production — pas encore provisionné

- **Demandé** : une application déployée avec staging automatique et
  déploiements production manuels.
- **Pourquoi irréalisé aujourd'hui** : la VM de production est une ressource
  payante dont l'achat est délibérément séquencé après le feature-complete
  (la stratégie en deux passes documentée du projet : d'abord le périmètre
  complet, puis la passe optimisation/sécurisation/déploiement). Aucun
  serveur n'existe à la date de l'audit.
- **Ce qui SERA utilisé (décidé, pas vague)** : l'ADR 0006
  (`docs/adr/0006-cloud-provider-choice.md`) fixe la cible avec une étude de
  compromis complète : **OVHcloud B2-7** (2 vCPU dédiés, 7 Go de RAM, 50 Go
  NVMe, ~18 EUR/mois dans un plafond de 40 EUR/mois), région Gravelines ou
  Strasbourg. Rejetés avec raisons écrites : Hetzner CAX21 (moins cher mais
  affaiblit l'argument de résidence des données produit-FR/hébergeur-FR pour
  8 EUR/mois), Oracle Cloud Always Free (loterie de capacité + risque
  documenté de récupération de compte pour une pièce de portfolio censée
  vivre 2-3 ans), Scaleway (repli viable, conservé comme plan B), le
  serverless Cloud Run+Neon (force une réécriture du déploiement et casse la
  parité dev/prod de compose), AWS/GCP/Azure (free tiers limités dans le
  temps, coût 2-3x pour des specs équivalentes). L'argument RGPD/Schrems II
  en faveur d'un hébergeur français servant une audience `.fr` est de fond,
  pas du pavoisement.
- **Contournement en place** : toute la forme de production tourne en
  local — `compose.yml` est l'unité de déploiement (la même stack que la VM
  exécutera, ADR 0006 "Zero refactor of the deployment model") ; le CD publie
  déjà des images multi-arch déployables sur `ghcr.io/ianadou/smo-app`
  (`cd.yml`) ; les HEALTHCHECK Docker + Dockhand fournissent la supervision
  que la VM aura. La grille accepte local + documenté + captures d'écran pour
  exactement ce cas.

## 2. Domaine et TLS — domaine acheté, pas encore en service

- **Demandé** : une URL de production accessible.
- **Statut** : le domaine **sportpotes.fr est déjà acheté** (enregistrement
  OVH le 2026-04-30, durée 3 ans — référencé dans le contexte de l'ADR 0006).
  L'activation DNS/TLS attend la VM (entrée 1).
- **Ce qui SERA utilisé** : Nginx en reverse proxy sur la VM avec Let's
  Encrypt via certbot (décision de l'ADR 0006 : "TLS: Let's Encrypt via
  certbot wired into Nginx" ; explicitement **pas de CDN/Cloudflare** devant
  — rejeté antérieurement, consigné dans l'ADR). L'app est déjà prête pour le
  proxy : les CIDR de proxys de confiance sont configurables
  (`TRUSTED_PROXIES`, `cmd/server/main.go:394-415`) pour que `c.ClientIP()` —
  sur lequel le rate limiter indexe — résolve la vraie IP cliente derrière
  Nginx ; les origines CORS sont pilotées par l'environnement
  (`ALLOWED_ORIGINS`, lignes 417-439).
- **Contournement** : les origines localhost sont le défaut documenté pour le
  dev et le e2e.

## 3. Outillage payant/SaaS — résolu avec des free tiers, aucun compromis nécessaire

| Besoin | Option payante/lourde non retenue | Ce que SMO utilise à la place | Preuve |
|---|---|---|---|
| SaaS d'analyse statique | SonarQube auto-hébergé (exige un serveur) | **SonarCloud gratuit pour les dépôts publics**, analyse pilotée par la CI ingérant la couverture réelle | `ci.yml:287-313`, `sonar-project.properties` |
| Suivi de couverture | Plans payants Codecov | **Mode Codecov tokenless/gratuit pour dépôts publics**, explicitement non bloquant en cas de panne | `ci.yml:246-255` |
| Registry de containers | Registry payant / limites de débit Docker Hub | **ghcr.io** avec le `GITHUB_TOKEN` intégré (aucun identifiant à gérer) | `cd.yml:51-56` |
| Bot de mise à jour des dépendances | — | **Renovate** (gratuit), PR groupées, alertes de sécurité à tout moment | `renovate.json` |
| Scan de secrets | DLP payant | GitGuardian côté serveur (free tier) + **deux gardes pre-commit maison** qui bloquent avant le push | `.pre-commit-config.yaml:52-78`, `scripts/check-discord-webhook.sh` |
| Scan de vulnérabilités | Snyk payant | Trivy + govulncheck + OSV-scanner + pnpm audit, tous gratuits, tous bloquants | `ci.yml:57-162` |

Aucune exigence de la grille n'a été abandonnée pour des raisons de coût dans
cette catégorie.

## 4. Notifications temps réel — consciemment remplacées par des webhooks Discord

- **Demandé** (par la forme du produit, pas par la grille) : notifier les
  joueurs quand les équipes sont prêtes.
- **Pourquoi pas de WebSocket/SSE** : les joueurs sont des porteurs de tokens
  non authentifiés sans connexion persistante ; faire tourner une
  infrastructure de push pour un produit à ~100 organisateurs est
  disproportionné. CLAUDE.md consigne la décision : "Discord webhooks replace
  real-time notifications".
- **Ce qui est utilisé** : l'événement de domaine `MatchTeamsReady` déclenche
  `discord.Subscriber` -> POST webhook (ADR 0003/0004), en best-effort avec
  un timeout dur de 5s pour qu'un Discord lent ne puisse jamais bloquer un
  use case (`cmd/server/main.go:66-70`).

## 5. Chiffrement de secrets par colonne — rejeté avec un modèle de menace documenté

- **Demandé** (implicitement, par les bonnes pratiques de sécurité) :
  chiffrer les secrets stockés.
- **Décision** : l'URL du webhook Discord est stockée en clair dans Postgres.
  L'ADR 0003 D2 porte la justification complète : du pgcrypto sur une seule
  colonne créerait un modèle de menace incohérent (les autres secrets restent
  en clair), exige un KMS que le projet n'a pas, et le serveur d'app lit
  toute clé de chiffrement depuis le même environnement que le mot de passe
  de la base — "equally exposed". Des contrôles compensatoires sont
  implémentés et testés : jamais loggué (expurgation des erreurs du
  notifier), jamais renvoyé via HTTP (booléen `has_webhook` uniquement + un
  test de garde vérifiant que la sous-chaîne de l'URL est absente du JSON
  sérialisé), validation stricte dans le domaine (https uniquement, pas
  d'identifiants embarqués, plafond de longueur aussi imposé par une
  contrainte CHECK en base). Condition de réexamen documentée : chiffrement
  au repos **uniforme** quand un KMS arrivera.

## 6. Ce qui est simulé en local en attendant — la correspondance explicite

| Élément de production | Simulation locale aujourd'hui | Pont déjà construit |
|---|---|---|
| VM OVH B2-7 | machine du développeur + `compose.yml` | définition de stack identique ; argument de parité dev/prod de l'ADR 0006 |
| Nginx + Let's Encrypt | ports directs 8081/3000 | coutures env `TRUSTED_PROXIES` + `ALLOWED_ORIGINS` prêtes |
| Terraform sur `ovh/ovh` | > **MANQUE :** `deploy/terraform/` vide — root local `kreuzwerker/docker` planifié (voir [infrastructure.md](infrastructure.md)) | répartition des providers documentée dans l'ADR 0006 |
| Ansible contre l'inventaire de la VM | > **MANQUE :** `deploy/ansible/` vide | liste des rôles cadrée dans [infrastructure.md](infrastructure.md) |
| Environnement cible du CD | publication GHCR uniquement (`cd.yml`) | tags SHA immuables = mécanisme de rollback (runbook de [ci-cd-pipeline.md](ci-cd-pipeline.md)) |
| Prometheus/Grafana sur la VM | > **MANQUE :** pas encore dans compose | hook de métriques par subscriber d'événements pré-conçu (ADR 0004) |
| UI de monitoring | Dockhand sur `localhost:3009` lisant les HEALTHCHECK des containers | chaque container embarque un healthcheck exactement pour cela |

## 7. Note de coût (pour la question « pourquoi ne pas simplement l'acheter »)

Engagé : sportpotes.fr ~ quelques EUR/an (enregistrement 3 ans, acquis).
Planifié : OVH B2-7 ~18 EUR/mois — sous le plafond auto-imposé de
40 EUR/mois consigné dans l'ADR 0006, laissant de la marge pour la voie de
montée en gamme B2-15 (~32 EUR/mois) sans changer de fournisseur ni de
scripts. Tout le reste de la chaîne d'outils (GitHub Actions, GHCR,
SonarCloud, Codecov, Renovate, Trivy, OSV) est gratuit à l'échelle de ce
projet et avec son statut de dépôt public.

## 8. Contraintes d'honnêteté que le projet s'impose à lui-même

Deux règles auto-imposées façonnent chaque contournement ci-dessus et
méritent d'être énoncées en soutenance parce qu'elles sont visibles dans le
code :

- **La couverture doit être réelle, jamais cosmétique** : les exclusions du
  dénominateur de couverture se limitent au code réellement incouvrable,
  chacune justifiée en ligne (`ci.yml:204-217`, `codecov.yml:29-41`) ; le
  gate suit le chiffre mesuré ("that drop is the truth showing through, not
  a regression" — `ci.yml:230-236`).
- **Les dégradations doivent être bruyantes** : pas de Redis -> WARN
  "per-account lockout unavailable" au démarrage (`main.go:307`) ; la
  readiness rapporte `cache: disabled` au lieu de faire semblant
  (`health_handler.go:50-56`). Rien n'échoue silencieusement pour paraître
  plus vert que la réalité.
