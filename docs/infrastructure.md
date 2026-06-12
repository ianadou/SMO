# Infrastructure — Docker, Compose, Terraform, Ansible

## Lecture commentée du Dockerfile backend (`Dockerfile`)

Deux stages, chaque décision annotée dans le fichier lui-même.

### Stage 1 — builder (`Dockerfile:16-48`)

- Base `golang:1.26.4-bookworm`, **version pinnée** (Renovate la met à jour
  via le groupe de manager `dockerfile`, `renovate.json:26-31`).
- **L'ordre des couches est la stratégie de cache** : `COPY go.mod go.sum` +
  `go mod download` viennent avant `COPY . .` (lignes 20-27), si bien que la
  couche de téléchargement des dépendances est réutilisée tant que les
  fichiers de modules ne changent pas — seules les modifications de source
  invalident la couche de build.
- Drapeaux de build (lignes 29-39) : `CGO_ENABLED=0` produit un binaire
  entièrement statique (obligatoire pour `distroless/static`), `-trimpath`
  retire les chemins du système de fichiers local, `-ldflags="-s -w"` retire
  les tables DWARF/symboles pour réduire la taille.
- Un **second binaire statique** est construit : `cmd/healthcheck`
  (lignes 44-48). Raison : distroless n'a pas de shell, donc
  `HEALTHCHECK CMD curl ...` est impossible — la sonde est un programme Go de
  40 lignes (`cmd/healthcheck/main.go`) qui fait un GET sur `/health/ready`
  et sort avec 0/1. Les deux RUN `go build` ne sont délibérément pas
  fusionnés (hadolint DL3059 ignoré avec justification dans `ci.yml:53-56`)
  pour conserver des couches de cache séparées.

### Stage 2 — runtime (`Dockerfile:57-84`)

- Base `gcr.io/distroless/static-debian12` pinnée sur un **digest SHA256
  immuable** (ligne 57), pas seulement le tag `:nonroot` — builds
  reproductibles entre machines et dans le temps ; le commentaire documente
  la procédure de mise à jour.
- Labels OCI (lignes 60-63) : titre, description, source, licence. Le
  pipeline CD ajoute les labels revision/version via
  `docker/metadata-action` (`cd.yml:72`).
- `USER nonroot:nonroot` (ligne 70) — UID 65532, pas de root dans l'image de
  runtime. Combiné à distroless : pas de shell, pas de gestionnaire de
  paquets, pas de libc.
- `HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=3`
  (lignes 80-81) avec la justification du réglage en commentaire : 10s garde
  l'UI de supervision locale (Dockhand) réactive, le timeout de 3s absorbe la
  latence de démarrage à froid sur un petit VPS, le start-period de 15s
  couvre la connexion Postgres + les migrations goose au démarrage. La sonde
  cible `/health/ready` (pas `/health/live`) à dessein : le container doit
  passer unhealthy quand la base de données est injoignable, pas seulement
  quand le processus meurt (`cmd/healthcheck/main.go:1-7`).

### Taille de l'image

`Dockerfile:9` documente « ~20 MB » (binaire statique strippé + base
distroless static, qui pèse ~2 MB). C'est très en dessous de la cible de
200 MB de la grille.

> **MANQUE :** la taille réelle est affirmée dans un commentaire mais pas
> consignée comme preuve. Exécuter une fois
> `docker build -t smo-app . && docker images smo-app` et coller la sortie
> ici (et dans le README).

### Posture vis-à-vis des secrets

- `.dockerignore` exclut `.env` et `.env.*` (conserve `.env.example`),
  `.git`, les artefacts de build — les secrets ne peuvent pas entrer dans le
  contexte de build.
- Aucun secret n'est intégré au build ; la configuration passe uniquement par
  l'environnement au runtime. Le binaire **refuse de démarrer** sans
  `JWT_SECRET` (`cmd/server/main.go:116-119`) plutôt que de retomber sur une
  valeur par défaut faible.
- Deux gardes pre-commit bloquent tout contenu en forme de secret avant qu'il
  n'atteigne git : `scripts/check-discord-webhook.sh` et le
  `detect-private-key` de `pre-commit-hooks` (`.pre-commit-config.yaml`).

## Dockerfile frontend (`frontend/Dockerfile`)

Également multi-stage : builder `node:24.10.0-alpine` (pnpm via corepack,
ordre des couches lockfile-d'abord) -> runtime `node:24.10.0-alpine` qui ne
copie que le bundle Nitro `.output`, s'exécute avec l'utilisateur non-root
`node`, et porte son propre HEALTHCHECK (wget sur `/`, fourni par alpine).
Distroless a été évalué et **rejeté avec une raison écrite** (en-tête du
fichier) : Nuxt a besoin du runtime Node, et node:alpine (~50 MB) garde wget
disponible pour la sonde — un compromis documenté, pas un oubli.

## Stack Compose (`compose.yml`)

Quatre services pour le dev local et la future forme de production
mono-VM (la parité dev/prod est un objectif explicite, ADR 0006) :

| Service | Image / build | Décisions notables |
|---|---|---|
| `postgres` | `postgres:16-alpine` | Volume nommé ; healthcheck via `pg_isready` ; port hôte 5433 pour éviter la collision avec un Postgres local (commentaire, lignes 33-35). |
| `redis` | `redis:7-alpine` | Healthcheck via `redis-cli ping` ; supporte le cache, le rate limiting et le verrouillage de connexion. |
| `app` | construit depuis le `Dockerfile` racine | `depends_on` avec `condition: service_healthy` sur les deux backends — l'app ne démarre que lorsque ses dépendances répondent à leurs healthchecks. L'hôte de `DATABASE_URL` est le nom du service (le commentaire explique le piège 5432-vs-5433). |
| `frontend` | construit depuis `frontend/Dockerfile` | Délibérément NON conditionné par la santé de l'app, avec la raison en commentaire (lignes 90-93) : la SPA parle à l'API depuis le navigateur, donc un hoquet du backend ne doit pas faire tomber la SPA. |

Gestion fail-fast des secrets : `POSTGRES_PASSWORD` et `JWT_SECRET` utilisent
le garde compose `${VAR:?error}` (lignes 30, 64) — **aucun mot de passe par
défaut n'existe nulle part** ; compose refuse de démarrer sans un `.env`
explicite, et le commentaire d'en-tête indique que c'est pour forcer "a
conscious choice instead of a hardcoded password baked into version control".

> **MANQUE :** CLAUDE.md décrit la stack compose comme
> "app + postgres + redis + prometheus + grafana", mais `compose.yml` ne
> contient aucun service prometheus ou grafana. Soit les ajouter (voir
> [monitoring.md](monitoring.md)), soit corriger l'affirmation de CLAUDE.md
> avant la soutenance — un examinateur qui lit les deux fichiers le
> remarquera.

## Terraform — `deploy/terraform/`

**État actuel : le répertoire est vide.** Aucun fichier `.tf` n'existe dans
le dépôt à la date de cet audit.

Ce qui est déjà décidé et documenté (pour que l'implémentation soit cadrée,
pas imaginaire) :

- L'ADR 0006 fixe la cible : OVHcloud B2-7 (2 vCPU, 7 Go, ~18 EUR/mois),
  région GRA ou SBG, et énonce la répartition des providers : "Terraform code
  under `deploy/terraform/` will use the `ovh/ovh` provider for OVH-side
  resource management; the existing `kreuzwerker/docker` provider continues
  to handle VM-internal Docker provisioning".
- La voie de simulation locale (le provider Docker créant l'équivalent
  compose — réseau/volumes/containers — sur la machine du développeur) est le
  repli « pas de cloud » accepté par la grille, à condition d'être justifié —
  justification dans
  [constraints-and-workarounds.md](constraints-and-workarounds.md).

> **MANQUE (le plus gros du projet, /3 points) :** écrire le code Terraform.
> Minimum viable pour la grille : un root `deploy/terraform/local/` utilisant
> `kreuzwerker/docker` pour provisionner le réseau SMO + les containers
> postgres/redis/app (prouve la mécanique provider/ressources/état), avec
> l'idempotence de `terraform plan` démontrée (second apply = aucun
> changement) et l'état local documenté (`terraform.tfstate` gitignoré ;
> l'état distant listé comme suite de production). Un root `ovh/` peut rester
> à l'état de variables + stub de provider jusqu'à l'achat de la VM.

## Ansible — `deploy/ansible/`

**État actuel : le répertoire est vide.** Aucun playbook ni inventaire
n'existe.

Le rôle prévu est documenté (CLAUDE.md : "Ansible for container configuration
and deployment" ; les conséquences de l'ADR 0006 notent que l'inventaire
pinnera le hostname de l'instance OVH et que toute hypothèse cloud-init
héritée de l'époque Hetzner devra être auditée au moment du déploiement).

> **MANQUE :** écrire un playbook avec des rôles idempotents : (1) installer
> Docker + le plugin compose (module apt, idempotent par nature),
> (2) templater `.env` depuis des variables chiffrées avec Ansible Vault,
> (3) copier `compose.yml`, (4) `community.docker.docker_compose_v2` pour
> faire converger la stack (ne rapporte `changed` que lorsque les containers
> changent réellement — idempotence démontrable en soutenance : exécuter deux
> fois, montrer `changed=0` à la seconde exécution).

## Idempotence et état — comment ce sera défendu

- **Terraform** : le fichier d'état est l'enregistrement de ce qui existe ;
  l'idempotence se démontre par un second `terraform apply` produisant un
  plan vide. Backend local pour le rendu scolaire (documenté), backend
  distant (object storage compatible S3 d'OVH) listé comme suite de
  production.
- **Ansible** : des modules (pas des commandes shell) pour que chaque tâche
  converge au lieu de se réexécuter ; la seconde exécution montre
  `changed=0`.
- **Déjà idempotent aujourd'hui, à citer** : les migrations goose
  s'appliquent exactement une fois (versionnées, append-only, exécutées au
  démarrage — `cmd/server/main.go:131-134`), et `docker compose up -d` fait
  lui-même converger la stack locale. Ce sont de vrais mécanismes
  d'idempotence sur lesquels le projet s'appuie déjà quotidiennement.
