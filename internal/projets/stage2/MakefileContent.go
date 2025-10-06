package stage2

import "fmt"

func MakefileContent(nameFolderProject string) string {
	return fmt.Sprintf(`-include .env

# Redefinir MAKEFILE_LIST pour qu'il ne contienne que le Makefile
MAKEFILE_LIST := Makefile

ENV_FILE := --env-file .env

# Couleurs
GREEN = \033[0;32m
YELLOW = \033[0;33m
NC = \033[0m # No Color

# Variables
COMPOSE_FILE = $(if $(filter $(APP_ENV),prod),docker/compose.prod.yaml,$(if $(filter $(APP_ENV),preprod),docker/compose.preprod.yaml,docker/compose.yaml))
DOCKER_COMPOSE = docker compose $(ENV_FILE) -f $(COMPOSE_FILE)

.PHONY: help build up down logs shell restart clean status ps ta tap tav tavp tf tfv

help: ## Affiche cette aide
	@echo ""
	@echo "Commandes disponibles:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%%-15s$(NC) $(YELLOW)%%s$(NC)\n", $$1, $$2}'
	@echo ""

build: ## build all images
	$(DOCKER_COMPOSE) build

build-front: ## build image front
	$(DOCKER_COMPOSE) build front

build-api: ## build image api
	$(DOCKER_COMPOSE) build api

up: ## start container
	$(DOCKER_COMPOSE) up -d

upb: ## build + start container
	$(DOCKER_COMPOSE) up -d --build

upl: ## start container + logs
	$(DOCKER_COMPOSE) up

upbl: ## build + start container +
	$(DOCKER_COMPOSE) up --build

deploy: ## pull and start containers (for CI/CD)
	$(DOCKER_COMPOSE) pull
	$(DOCKER_COMPOSE) up -d

down: ## stop and delete container
	$(DOCKER_COMPOSE) down

stop: ## stop and delete container
	$(DOCKER_COMPOSE) stop

lfront: ## logs front
	$(DOCKER_COMPOSE) logs -f front

lapi: ## logs api
	$(DOCKER_COMPOSE) logs -f api

ldb: ## logs db
	$(DOCKER_COMPOSE) logs -f db

shfront: ## shell conteneur front
	$(DOCKER_COMPOSE) exec front bash

shapi: ## shell conteneur api
	$(DOCKER_COMPOSE) exec api bash

shdb: ## shell conteneur db mongo
	$(DOCKER_COMPOSE) exec db mongosh

ta: ## Lance tous les tests api
	docker exec -i -e APP_ENV=test %s_dev_api go test ./...

tai: ## Lance tous les tests api d'integration avec logs (fmt-print)
	docker exec -i -e APP_ENV=test %s_dev_api go test -tags=integration ./...

tap: ## Lance les tests api pour un path spécifique (usage: make tap path=monpath)
	docker exec -i -e APP_ENV=test %s_dev_api go test ./$(path)

taip: ## Lance les tests api d'integration pour un path spécifique (usage: make tap path=monpath)
	docker exec -i -e APP_ENV=test %s_dev_api go test -tags=integration ./$(path)

tav: ## Lance tous les tests api en verbose
	docker exec -i -e APP_ENV=test %s_dev_api go test -v ./...

taiv: ## Lance tous les tests api en verbose + integration
	docker exec -i -e APP_ENV=test %s_dev_api go test -tags=integration -v ./...

tavp: ## Lance les tests api en verbose pour un path (usage: make tavp path=monpath)
	docker exec -i -e APP_ENV=test %s_dev_api go test -v ./$(path)

taivp: ## Lance les tests api en verbose + integration pour un path (usage: make tavp path=monpath)
	docker exec -i -e APP_ENV=test %s_dev_api go test -v -tags=integration ./$(path)
`,
		nameFolderProject, // ta
		nameFolderProject, // tai
		nameFolderProject, // tap
		nameFolderProject, // taip
		nameFolderProject, // tav
		nameFolderProject, // taiv
		nameFolderProject, // tavp
		nameFolderProject, // taivp
	)
}
