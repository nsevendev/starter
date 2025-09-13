package stage1

import "fmt"

func MakefileContent() string {
	return fmt.Sprintf(`
-include .env

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
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) $(YELLOW)%s$(NC)\n", $$1, $$2}'
	@echo ""

b: ## Construit toutes les images Docker
	$(DOCKER_COMPOSE) build

bapp: ## Construit uniquement l'image de l'application frontend
	$(DOCKER_COMPOSE) build app

up: ## Lance tous les services
	$(DOCKER_COMPOSE) up -d

upb: ## rebuild l'image et lance tous les services
	$(DOCKER_COMPOSE) up -d --build

upl: ## Lance tous les services avec les logs
	$(DOCKER_COMPOSE) up

upbl: ## Lance tous les services avec les logs
	$(DOCKER_COMPOSE) up --build

down: ## Arrête tous les services et supprime les containers
	$(DOCKER_COMPOSE) down

stop: ## Arrête tous les services
	$(DOCKER_COMPOSE) stop

l: ## Affiche les logs de tous les services
	$(DOCKER_COMPOSE) logs -f

lapp: ## Affiche les logs de l'application frontend
	$(DOCKER_COMPOSE) logs -f app

shapp: ## Ouvre un shell dans le conteneur de l'application frontend
	$(DOCKER_COMPOSE) exec app bash

taf: ## Lance tous les tests de l'app
	$(DOCKER_COMPOSE) exec app npm run test

taff: ## Lance le test du fichier donné de l'app (usage: make taff file=pathdufichier)
	$(DOCKER_COMPOSE) exec app npm run test -- --include $(file)

tafc: ## Lance tous les tests de l'app en mode CI (headless)
	$(DOCKER_COMPOSE) exec app npm run test:ci

`)
}
