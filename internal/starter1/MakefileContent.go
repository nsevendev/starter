package starter1

import "fmt"

func MakefileContent() string {
	return fmt.Sprintf(`
-include .env

# Redéfinir MAKEFILE_LIST pour qu'il ne contienne que le Makefile
MAKEFILE_LIST := Makefile

ENV_FILE := --env-file .env

# Couleurs
GREEN = \033[0;32m
YELLOW = \033[0;33m
NC = \033[0m # No Color

# Variables
COMPOSE_FILE = $(if $(filter $(APP_ENV),prod),docker/compose.prod.yaml,$(if $(filter $(APP_ENV),preprod),docker/compose.preprod.yaml,docker/compose.yaml))
DOCKER_COMPOSE = docker compose $(ENV_FILE) -f $(COMPOSE_FILE)

.PHONY: help build up upb upl down stop l lapp shapp ps status clean

help: ## Affiche cette aide
	@echo ""
	@echo "Commandes disponibles:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) $(YELLOW)%s$(NC)\n", $$1, $$2}'
	@echo ""

build: ## Construit les images Docker du fichier compose ciblé (APP_ENV=dev|preprod|prod)
	$(DOCKER_COMPOSE) build

build-app: ## Construit uniquement l'image de l'application
	$(DOCKER_COMPOSE) build app

up: ## Lance les services en arrière-plan
	$(DOCKER_COMPOSE) up -d

upb: ## Reconstruit et lance les services en arrière-plan
	$(DOCKER_COMPOSE) up -d --build

upl: ## Lance les services avec les journaux au premier plan
	$(DOCKER_COMPOSE) up

upbl: ## Build et lance les services avec les journaux au premier plan
	$(DOCKER_COMPOSE) up --build

down: ## Arrête et supprime les conteneurs
	$(DOCKER_COMPOSE) down

stop: ## Arrête les conteneurs
	$(DOCKER_COMPOSE) stop

l: ## Affiche les journaux de tous les services
	$(DOCKER_COMPOSE) logs -f

lapp: ## Affiche les journaux de l'application
	$(DOCKER_COMPOSE) logs -f app

shapp: ## Ouvre un shell dans le conteneur de l'application
	$(DOCKER_COMPOSE) exec app sh

ps: ## Liste les conteneurs du projet
	$(DOCKER_COMPOSE) ps

taf: ## Lance tous les tests de l'app (attention modifier le nom du container)
	docker exec -it app_dev_app npm run test

tafc: ## Lance tous les tests de l'app en mode CI (headless) (attention modifier le nom du container)
	docker exec app_dev_app npm run test:ci
`)
}
