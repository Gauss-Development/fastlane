SHELL := /bin/bash

COMPOSE_FILE ?= docker-compose.yml
DC := docker compose -f $(COMPOSE_FILE)

# Optional service for service-scoped commands, e.g. make logs-svc SVC=api-gateway
SVC ?=
# Optional command for exec/run helpers
CMD ?= sh

SERVICES := api-gateway auth-service user-service post-service search-service notification-service design-service catalog-service

.PHONY: help setup compose up up-d down down-v stop start restart build build-no-cache pull push \
	ps top images config logs logs-f logs-svc shell exec run \
	infra-up infra-down app-up app-down clean prune seed embed embed-dry

help: ## Show available commands
	@echo "Microblog Docker Compose commands"
	@echo ""
	@grep -E '^[a-zA-Z0-9_.-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make up-d"
	@echo "  make logs-f"
	@echo "  make logs-svc SVC=auth-service"
	@echo "  make shell SVC=api-gateway"
	@echo "  make exec SVC=user-service CMD='ls -la /app'"
	@echo "  make compose ARGS='events --since 10m'"

setup: ## Bootstrap local env files from committed templates (idempotent; never clobbers)
	@cp -n .env.example .env 2>/dev/null && echo "created .env" || echo ".env exists — kept"
	@for s in $(SERVICES); do \
		cp -n services/$$s/.env.example services/$$s/.env 2>/dev/null && echo "created services/$$s/.env" || echo "services/$$s/.env exists — kept"; \
	done
	@echo "Done. Fill in secrets (JWT_SECRET, *_PASSWORD, *_API_KEY, ...) in .env before 'make up-d'."

compose: ## Pass through any docker compose args: make compose ARGS='ps'
	$(DC) $(ARGS)

up: ## Start all services in foreground
	$(DC) up

up-d: ## Start all services in detached mode
	$(DC) up -d

down: ## Stop and remove containers/network
	$(DC) down

down-v: ## Stop and remove containers/network/volumes
	$(DC) down -v

stop: ## Stop running services
	$(DC) stop

start: ## Start existing services
	$(DC) start

restart: ## Restart all services
	$(DC) restart

build: ## Build or rebuild services
	$(DC) build

build-no-cache: ## Build services without cache
	$(DC) build --no-cache

pull: ## Pull service images
	$(DC) pull

push: ## Push service images
	$(DC) push

ps: ## List containers
	$(DC) ps

top: ## Display running processes
	$(DC) top

images: ## List images used by compose services
	$(DC) images

config: ## Validate and view resolved compose config
	$(DC) config

logs: ## Show logs for all services
	$(DC) logs

logs-f: ## Follow logs for all services
	$(DC) logs -f --tail=200

logs-svc: ## Follow logs for one service (SVC=service-name)
	@test -n "$(SVC)" || (echo "SVC is required. Example: make logs-svc SVC=api-gateway" && exit 1)
	$(DC) logs -f --tail=200 $(SVC)

shell: ## Open shell in service container (SVC=service-name)
	@test -n "$(SVC)" || (echo "SVC is required. Example: make shell SVC=api-gateway" && exit 1)
	$(DC) exec $(SVC) sh

exec: ## Exec custom command in service container (SVC=..., CMD='...')
	@test -n "$(SVC)" || (echo "SVC is required. Example: make exec SVC=auth-service CMD='ls -la'" && exit 1)
	$(DC) exec $(SVC) $(CMD)

run: ## Run one-off command in new service container (SVC=..., CMD='...')
	@test -n "$(SVC)" || (echo "SVC is required. Example: make run SVC=post-service CMD='go test ./...'" && exit 1)
	$(DC) run --rm $(SVC) $(CMD)

infra-up: ## Start infrastructure only (redis, postgres, rabbitmq, minio, prometheus, grafana)
	$(DC) up -d redis postgres_user postgres_post postgres_notification postgres_design postgres_catalog rabbitmq minio createbuckets prometheus grafana

infra-down: ## Stop infrastructure only
	$(DC) stop redis postgres_user postgres_post postgres_notification postgres_design postgres_catalog rabbitmq minio prometheus grafana

app-up: ## Start app services only (without infra)
	$(DC) up -d auth-service user-service post-service notification-service search-service design-service catalog-service api-gateway

app-down: ## Stop app services only
	$(DC) stop auth-service user-service post-service notification-service search-service design-service catalog-service api-gateway

clean: ## Compose down + remove volumes
	$(DC) down -v --remove-orphans

prune: ## Compose down + remove volumes and local images
	$(DC) down -v --rmi local --remove-orphans

seed: ## Seed post-service Postgres with real suppliers + transceiver SKUs (Fiberlane catalog)
	@test -n "$$POSTGRES_POST_PASSWORD" || { . ./.env >/dev/null 2>&1 || true; }; \
	cd services/post-service && \
	DATABASE_URL="postgres://postgres:$${POSTGRES_POST_PASSWORD}@localhost:$${POSTGRES_POST_HOST_PORT:-15432}/postdb?sslmode=disable" \
	  go run ./cmd/seed -seeds ./seeds

embed: ## Generate vector embeddings for catalog products (Voyage primary, OpenAI fallback)
	@test -n "$$POSTGRES_POST_PASSWORD" || { . ./.env >/dev/null 2>&1 || true; }; \
	cd services/post-service && \
	DATABASE_URL="postgres://postgres:$${POSTGRES_POST_PASSWORD}@localhost:$${POSTGRES_POST_HOST_PORT:-15432}/postdb?sslmode=disable" \
	VOYAGE_API_KEY="$${VOYAGE_API_KEY}" OPENAI_API_KEY="$${OPENAI_API_KEY}" \
	  go run ./cmd/embed

embed-dry: ## Print embedding texts for catalog products without calling any API
	@test -n "$$POSTGRES_POST_PASSWORD" || { . ./.env >/dev/null 2>&1 || true; }; \
	cd services/post-service && \
	DATABASE_URL="postgres://postgres:$${POSTGRES_POST_PASSWORD}@localhost:$${POSTGRES_POST_HOST_PORT:-15432}/postdb?sslmode=disable" \
	  go run ./cmd/embed -dry-run -limit 5

embed-fake: ## Populate embeddings with deterministic hash-based vectors (local dev only — no semantic signal)
	@test -n "$$POSTGRES_POST_PASSWORD" || { . ./.env >/dev/null 2>&1 || true; }; \
	cd services/post-service && \
	DATABASE_URL="postgres://postgres:$${POSTGRES_POST_PASSWORD}@localhost:$${POSTGRES_POST_HOST_PORT:-15432}/postdb?sslmode=disable" \
	  go run ./cmd/embed -fake
