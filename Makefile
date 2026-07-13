# Kore — environnement local Docker Compose
# Usage : make help

SHELL := /bin/bash
COMPOSE_FILE := deploy/docker-compose.yml
ENV_FILE     := .env
COMPOSE      := docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE)

KORE_API_PORT      ?= 8081
KORE_FRONTEND_PORT ?= 3001
KORE_DB_PORT       ?= 5434
KORE_REDIS_PORT    ?= 6381

# Surcharge depuis .env si présent
ifneq (,$(wildcard $(ENV_FILE)))
KORE_API_PORT      := $(shell grep -E '^KORE_API_PORT=' $(ENV_FILE) | cut -d= -f2 | tr -d '\r')
KORE_FRONTEND_PORT := $(shell grep -E '^KORE_FRONTEND_PORT=' $(ENV_FILE) | cut -d= -f2 | tr -d '\r')
KORE_DB_PORT       := $(shell grep -E '^KORE_DB_PORT=' $(ENV_FILE) | cut -d= -f2 | tr -d '\r')
KORE_REDIS_PORT    := $(shell grep -E '^KORE_REDIS_PORT=' $(ENV_FILE) | cut -d= -f2 | tr -d '\r')
endif
KORE_API_PORT      ?= 8081
KORE_FRONTEND_PORT ?= 3001
KORE_DB_PORT       ?= 5434
KORE_REDIS_PORT    ?= 6381

.DEFAULT_GOAL := help

.PHONY: help env up up-infra up-front front down migrate seed seed-reset logs ps restart ready smoke \
        build api test test-integration lint sqlc frontend-dev frontend-install \
        gcp-setup gcp-config gcp-deploy gcp-deploy-jobs gcp-postdeploy gcp-smoke gcp-domain gcp-github-deploy

## Affiche les cibles disponibles
help:
	@echo "Kore — commandes Docker Compose locales"
	@echo ""
	@echo "  make env          Copie .env.example → .env si absent"
	@echo "  make up           Démarre la stack complète (infra + migrate + api + frontend)"
	@echo "  make up-infra     Démarre uniquement db, redis, mailhog, stripe-mock"
	@echo "  make up-front     Rebuild et redémarre uniquement le frontend (alias: front)"
	@echo "  make down         Arrête et supprime les conteneurs"
	@echo "  make migrate      Applique les migrations (service one-shot)"
	@echo "  make seed         Seed demo complet (tenant, org, CRA, congés, TMA, budget…)"
	@echo "  make seed-reset   Réinitialise et recharge le jeu de données demo"
	@echo "  make ready        Vérifie /health et /ready"
	@echo "  make smoke        Smoke test API complet"
	@echo "  make logs         Logs API (suivi)"
	@echo "  make ps           État des conteneurs"
	@echo ""
	@echo "  Admin dev : ADM_admin / Admin123!"
	@echo ""
	@echo "  GCP Premedica (premedica-prod-2025) :"
	@echo "  make gcp-setup          Bootstrap DB, secrets, SA, Artifact Registry"
	@echo "  make gcp-config         Publie la clé Gemini (Secret Manager) depuis .env.gemini"
	@echo "  make gcp-github-deploy  Workload Identity Federation GitHub Actions"
	@echo "  make gcp-deploy         Cloud Build → API + frontend"
	@echo "  make gcp-deploy-jobs    Cloud Run Jobs (migrate, seed)"
	@echo "  make gcp-postdeploy     Smoke test après deploy CI"
	@echo "  make gcp-postdeploy-full Première install : migrate + seed + smoke"
	@echo "  make gcp-domain         Domaine custom kore.ll-it-sc.be"
	@echo "  make gcp-smoke          Smoke test services déployés"
	@echo ""
	@echo "  Ports par défaut (modifiables dans .env) :"
	@echo "    API       http://localhost:$(KORE_API_PORT)"
	@echo "    Frontend  http://localhost:$(KORE_FRONTEND_PORT)"

## Prépare .env depuis .env.example
env:
	@if [ ! -f $(ENV_FILE) ]; then \
		cp .env.example $(ENV_FILE); \
		echo "→ $(ENV_FILE) créé depuis .env.example"; \
	else \
		echo "→ $(ENV_FILE) déjà présent (vérifiez KORE_API_PORT/KORE_FRONTEND_PORT si conflits)"; \
	fi

## Stack complète : build + détaché (migrate s'exécute avant api via depends_on)
up: env
	$(COMPOSE) up --build -d
	@echo ""
	@echo "Stack démarrée :"
	@echo "  API       http://localhost:$(KORE_API_PORT)"
	@echo "  Frontend  http://localhost:$(KORE_FRONTEND_PORT)"
	@echo "  MailHog   http://localhost:$${KORE_MAILHOG_UI_PORT:-8025}"
	@echo "  Postgres  localhost:$(KORE_DB_PORT)"
	@echo "  Redis     localhost:$(KORE_REDIS_PORT)"
	@echo "  Admin     ADM_admin / Admin123!"

## Infra seule — utile pour dev Go/Nuxt hors conteneur
up-infra: env
	$(COMPOSE) up -d db redis mailhog stripe-mock
	@echo "Infra prête. DATABASE_URL=postgres://kore:kore@localhost:$(KORE_DB_PORT)/kore?sslmode=disable"
	@echo "Puis : make migrate && make seed && HTTP_ADDR=:8081 make api"

## Rebuild et redémarre uniquement le service frontend (sans toucher api/infra)
up-front front: env
	$(COMPOSE) up --build --force-recreate -d --no-deps frontend
	@echo ""
	@echo "Frontend regénéré : http://localhost:$(KORE_FRONTEND_PORT)"

## Arrête la stack
down:
	$(COMPOSE) down

## Migrations via le runner Go embarqué (conteneur one-shot)
migrate: env
	$(COMPOSE) up -d db
	$(COMPOSE) run --rm --build --no-deps migrate
	@echo "→ migrations appliquées"

## Seed demo idempotent (tous modules)
seed: env
	$(COMPOSE) up -d db redis
	$(COMPOSE) run --rm --build --no-deps api seed
	@echo "→ seed appliqué — voir internal/seed/constants.go pour les comptes"

## Réinitialise et recharge le jeu de données demo complet
seed-reset: env
	$(COMPOSE) up -d db redis
	$(COMPOSE) run --rm --build --no-deps api seed reset
	@echo "→ seed reset appliqué — voir internal/seed/constants.go pour les comptes"

## Logs API
logs:
	$(COMPOSE) logs -f api

## État des services
ps:
	$(COMPOSE) ps -a

## Redémarre api et frontend
restart:
	$(COMPOSE) restart api frontend

## Smoke readiness
ready:
	@curl -sf "http://localhost:$(KORE_API_PORT)/health" | grep -q ok && echo "health OK" || (echo "health FAIL"; exit 1)
	@curl -sf "http://localhost:$(KORE_API_PORT)/ready"  | grep -q ready && echo "ready OK"  || (echo "ready FAIL"; exit 1)

## Smoke test complet (login + routes publiques)
smoke:
	@bash scripts/smoke-test.sh

# --- Développement local (hors Docker) ---

build:
	go build -o bin/kore-api ./cmd/kore-api

api: env
	go run ./cmd/kore-api

test:
	go test ./...

## Tests d'intégration via testcontainers (Docker requis, Postgres éphémère auto-géré)
test-integration:
	go test -tags=integration ./internal/platform/... ./internal/modules/...

lint:
	golangci-lint run ./...

sqlc:
	sqlc generate

frontend-dev:
	cd frontend && NUXT_PUBLIC_API_BASE=http://localhost:$(KORE_API_PORT) npm run dev

frontend-install:
	cd frontend && npm install

GCP_PROJECT_ID ?= premedica-prod-2025

## Bootstrap GCP (DB, secrets, SA) — une fois
gcp-setup:
	chmod +x infra/gcp/*.sh infra/gcp/lib/*.sh
	bash infra/gcp/setup-gcp.sh

## Applique la clé API Gemini (Secret Manager) depuis .env.gemini
gcp-config:
	chmod +x infra/gcp/*.sh infra/gcp/lib/*.sh
	bash infra/gcp/apply-config.sh

## WIF GitHub Actions → GCP
gcp-github-deploy:
	chmod +x infra/gcp/*.sh infra/gcp/lib/*.sh
	bash infra/gcp/setup-github-deploy.sh

## Déploiement Cloud Build (API + frontend + migrate)
gcp-deploy:
	gcloud builds submit --config=infra/gcp/cloudbuild.yaml --project=$(GCP_PROJECT_ID)

## Cloud Run Jobs uniquement
gcp-deploy-jobs:
	gcloud builds submit --config=infra/gcp/cloudbuild-jobs.yaml --project=$(GCP_PROJECT_ID)

## Postdeploy : jobs migrate/seed + smoke
gcp-postdeploy:
	chmod +x infra/gcp/*.sh infra/gcp/lib/*.sh
	bash infra/gcp/postdeploy.sh

## Postdeploy complet (première install : migrate + seed + smoke)
gcp-postdeploy-full:
	chmod +x infra/gcp/*.sh infra/gcp/lib/*.sh
	bash infra/gcp/postdeploy.sh --migrate --seed

## Domaine custom kore.ll-it-sc.be
gcp-domain:
	chmod +x infra/gcp/*.sh infra/gcp/lib/*.sh
	bash infra/gcp/setup-custom-domain.sh

## Smoke test GCP
gcp-smoke:
	chmod +x infra/gcp/*.sh infra/gcp/lib/*.sh
	bash infra/gcp/smoke-test.sh

## SSO Google : applique client_id/secret (.env.oidc) local + GCP
setup-oidc-google:
	chmod +x scripts/setup-google-oidc.sh infra/gcp/apply-oidc-config.sh
	bash scripts/setup-google-oidc.sh
