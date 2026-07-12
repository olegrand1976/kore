# Kore — environnement local Docker Compose
# Usage : make help

SHELL := /bin/bash
COMPOSE_FILE := deploy/docker-compose.yml
ENV_FILE     := .env
COMPOSE      := docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE)

.DEFAULT_GOAL := help

.PHONY: help env up up-infra down migrate seed logs ps restart ready \
        build api test test-integration lint sqlc frontend-dev frontend-install

## Affiche les cibles disponibles
help:
	@echo "Kore — commandes Docker Compose locales"
	@echo ""
	@echo "  make env          Copie .env.example → .env si absent"
	@echo "  make up           Démarre la stack complète (infra + migrate + api + frontend)"
	@echo "  make up-infra     Démarre uniquement db, redis, mailhog, stripe-mock"
	@echo "  make down         Arrête et supprime les conteneurs"
	@echo "  make migrate      Applique les migrations (service one-shot)"
	@echo "  make seed         Seed dev (tenant + admin ADM_admin)"
	@echo "  make logs         Logs API (suivi)"
	@echo "  make ps           État des conteneurs"
	@echo "  make restart      Redémarre api + frontend"
	@echo "  make ready        Vérifie /health et /ready"
	@echo ""
	@echo "  make build        Compile kore-api (local)"
	@echo "  make api          Lance l'API en local (hors Docker)"
	@echo "  make test         Tests unitaires Go"
	@echo "  make lint         golangci-lint"
	@echo "  make frontend-dev Dev Nuxt (hors Docker)"

## Prépare .env depuis .env.example
env:
	@if [ ! -f $(ENV_FILE) ]; then \
		cp .env.example $(ENV_FILE); \
		echo "→ $(ENV_FILE) créé depuis .env.example"; \
	else \
		echo "→ $(ENV_FILE) déjà présent"; \
	fi

## Stack complète : build + détaché (migrate s'exécute avant api via depends_on)
up: env
	$(COMPOSE) up --build -d
	@echo ""
	@echo "Stack démarrée :"
	@echo "  API       http://localhost:8080"
	@echo "  Frontend  http://localhost:3000"
	@echo "  MailHog   http://localhost:8025"
	@echo "  Postgres  localhost:5432"
	@echo "  Redis     localhost:6381"

## Infra seule (db, redis, mailhog, stripe-mock) — utile pour dev Go/Nuxt hors conteneur
up-infra: env
	$(COMPOSE) up -d db redis mailhog stripe-mock
	@echo "Infra prête. Lancez : make migrate && make api"

## Arrête la stack
down:
	$(COMPOSE) down

## Migrations via le runner Go embarqué (conteneur one-shot)
migrate: env
	$(COMPOSE) up -d db
	$(COMPOSE) run --rm --build --no-deps migrate
	@echo "→ migrations appliquées"

## Seed dev idempotent (tenant demo + admin)
seed: env
	$(COMPOSE) run --rm api seed

## Logs API
logs:
	$(COMPOSE) logs -f api

## État des services
ps:
	$(COMPOSE) ps

## Redémarre api et frontend
restart:
	$(COMPOSE) restart api frontend

## Smoke readiness
ready:
	@curl -sf http://localhost:8080/health | grep -q ok && echo "health OK" || (echo "health FAIL"; exit 1)
	@curl -sf http://localhost:8080/ready  | grep -q ready && echo "ready OK"  || (echo "ready FAIL"; exit 1)

# --- Développement local (hors Docker) ---

build:
	go build -o bin/kore-api ./cmd/kore-api

api: env
	go run ./cmd/kore-api

test:
	go test ./...

test-integration:
	docker compose -f deploy/docker-compose.test.yml up -d
	@sleep 3
	DATABASE_URL=postgres://kore:kore@localhost:5433/kore_test?sslmode=disable \
	REDIS_ADDR=localhost:6380 \
	go test -tags=integration ./internal/platform/... ./internal/modules/...
	docker compose -f deploy/docker-compose.test.yml down

lint:
	golangci-lint run ./...

sqlc:
	sqlc generate

frontend-dev:
	cd frontend && npm run dev

frontend-install:
	cd frontend && npm install
