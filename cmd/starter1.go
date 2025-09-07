package cmd

import (
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "strings"

    "github.com/spf13/cobra"
)

var (
    starter1Name       string
    starter1Port       int
    starter1NodeVer    string
    starter1HostRule   string
)

// starter1Cmd represents the command to create the Angular SSR starter
var starter1Cmd = &cobra.Command{
    Use:   "starter-1",
    Short: "Crée un projet Angular SSR (starter-1)",
    Long:  "Génère un projet Angular SSR sans backend ni base de données, avec Dockerfile multi-stage et compose (dev/preprod/prod).",
    RunE: func(cmd *cobra.Command, args []string) error {
        project := sanitizeName(starter1Name)
        if project == "" {
            return errors.New("nom de projet invalide")
        }

        fmt.Println("Starter-1: création du projet Angular SSR")
        fmt.Printf("- Dossier: %s\n", project)

        // Defaults/derived
        if starter1HostRule == "" {
            starter1HostRule = fmt.Sprintf("Host(`%s.localhost`)", project)
        }

        // Layout: <project>/{docker,app}
        root := project
        dockerDir := filepath.Join(root, "docker")
        appDir := filepath.Join(root, "app")

        // Create directories (root + docker only; app created by Angular CLI)
        if err := ensureDir(dockerDir); err != nil { return err }

        // Files: Dockerfile (multi-stage)
        dockerfilePath := filepath.Join(dockerDir, "Dockerfile")
        if err := writeFileIfAbsent(dockerfilePath, dockerfileContent(starter1NodeVer)); err != nil { return err }

        // Files: compose variants
        baseCompose := filepath.Join(dockerDir, "compose.yaml")
        preprodCompose := filepath.Join(dockerDir, "compose.preprod.yaml")
        prodCompose := filepath.Join(dockerDir, "compose.prod.yaml")
        compose := composeContent(project, "app")
        if err := writeFileIfAbsent(baseCompose, compose); err != nil { return err }
        if err := writeFileIfAbsent(preprodCompose, compose); err != nil { return err }
        if err := writeFileIfAbsent(prodCompose, compose); err != nil { return err }

        // Launch Angular CLI to create the app (interactive)
        fmt.Println("- Lancement de l'Angular CLI (interactif): ng new app --ssr ...")
        if err := runAngularCreate(root); err != nil {
            return fmt.Errorf("échec de la création Angular: %w", err)
        }

        // After Angular creation, add env files and README inside app
        appEnv := filepath.Join(appDir, ".env")
        appEnvDist := filepath.Join(appDir, ".env.dist")
        if err := writeFileIfAbsent(appEnv, appEnvContent(project, starter1Port, starter1NodeVer, starter1HostRule)); err != nil { return err }
        if err := writeFileIfAbsent(appEnvDist, appEnvContent(project, starter1Port, starter1NodeVer, starter1HostRule)); err != nil { return err }

        readme := filepath.Join(appDir, "README.md")
        if err := writeFileIfAbsent(readme, appReadme(project, "app")); err != nil { return err }

        fmt.Println("- Fichiers générés:")
        fmt.Printf("  %s\n", dockerfilePath)
        fmt.Printf("  %s\n", baseCompose)
        fmt.Printf("  %s\n", preprodCompose)
        fmt.Printf("  %s\n", prodCompose)
        fmt.Printf("  %s\n", appEnv)
        fmt.Printf("  %s\n", appEnvDist)
        fmt.Printf("  %s\n", readme)

        fmt.Println()
        // Preflight Docker/Compose and network guidance
        printDockerHints(project)

        fmt.Println("Next steps:")
        composeCmd := chooseComposeCmd()
        fmt.Printf("- cd %s/docker && APP_ENV=dev %s -f compose.yaml up --build\n", project, composeCmd)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(starter1Cmd)

    // Flags
    starter1Cmd.Flags().StringVarP(&starter1Name, "name", "n", "starter-1-app", "Nom du projet/dossier à créer")
    starter1Cmd.Flags().IntVarP(&starter1Port, "port", "p", 4200, "Port HTTP de l'app Angular")
    starter1Cmd.Flags().StringVar(&starter1NodeVer, "node", "20", "Version majeure de Node.js utilisée dans les images")
    starter1Cmd.Flags().StringVar(&starter1HostRule, "host", "", "Règle Traefik (ex: Host(`app.localhost`))")
}

// sanitizeName converts the provided name into a docker/host friendly slug
func sanitizeName(s string) string {
    s = strings.TrimSpace(strings.ToLower(s))
    s = strings.ReplaceAll(s, " ", "-")
    re := regexp.MustCompile(`[^a-z0-9-_]+`)
    s = re.ReplaceAllString(s, "")
    s = strings.Trim(s, "-")
    return s
}

func ensureDir(path string) error {
    if err := os.MkdirAll(path, 0o755); err != nil {
        return fmt.Errorf("création du dossier %s: %w", path, err)
    }
    return nil
}

func writeFileIfAbsent(path, content string) error {
    if _, err := os.Stat(path); err == nil {
        fmt.Printf("  (skip) %s existe déjà\n", path)
        return nil
    }
    if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
        return fmt.Errorf("écriture du fichier %s: %w", path, err)
    }
    return nil
}

func dockerfileContent(nodeVer string) string {
    if nodeVer == "" { nodeVer = "20" }
    return fmt.Sprintf(`# syntax=docker/dockerfile:1.7

ARG NODE_VERSION=%s

FROM node:${NODE_VERSION}-alpine AS base
WORKDIR /app
ENV NODE_ENV=development \\
    CI=true \\
    PUPPETEER_SKIP_DOWNLOAD=true

# Dépendances système utiles
RUN apk add --no-cache git bash

# Stage de dev: dépendances et outils
FROM base AS dev
ENV NODE_ENV=development
# Dépendances installées si package.json présent (no-fail pour repo vide)
COPY package*.json ./ 2>/dev/null || true
RUN [ -f package.json ] && npm ci || true
CMD ["npm", "run", "start"]

# Stage build: build SSR
FROM base AS build
ENV NODE_ENV=production
COPY . .
RUN [ -f package.json ] && npm ci && npm run build:ssr || true

# Stage runtime minimal
FROM node:${NODE_VERSION}-alpine AS runtime
ENV NODE_ENV=production \\
    PORT=4200
WORKDIR /app
# Copier uniquement les artefacts nécessaires (chemins communs Angular SSR)
COPY --from=build /app/dist /app/dist
COPY --from=build /app/package*.json /app/
RUN [ -f package.json ] && npm ci --omit=dev || true
EXPOSE 4200
# Entrée par défaut pour Angular SSR (adapter si besoin)
CMD ["sh", "-c", "node dist/server/server.mjs || node dist/server/main.js"]

# Alias pour cibles compose
FROM runtime AS prod
FROM runtime AS preprod
`, nodeVer)
}

func composeContent(project, appDir string) string {
    // Compose file placed in <project>/docker/, paths are relative to that
    return fmt.Sprintf(`name: %s-${APP_ENV}
services:
  %s:
    build:
      target: ${APP_ENV}
      context: ../%s
      dockerfile: ../docker/Dockerfile
      args:
        - NODE_VERSION=${NODE_VERSION}
    container_name: %s_${APP_ENV}_app
    image: %s-app:${APP_ENV}
    labels:
      - "traefik.enable=true"
      - "traefik.docker.network=%s-nseven"
      - "traefik.http.routers.%s.rule=${HOST_TRAEFIK_APP}"
      - "traefik.http.routers.%s.entrypoints=websecure"
      - "traefik.http.routers.%s.tls=true"
      - "traefik.http.routers.%s.tls.certresolver=default"
      - "traefik.http.services.%s.loadbalancer.server.port=${PORT}"
      - "traefik.http.services.%s.loadbalancer.server.scheme=http"
    volumes:
      - ../%s:/app
    env_file:
      - ../%s/.env
    networks:
      - %s-nseven
      - %s

networks:
  %s-nseven:
    external: true
  %s:
    driver: bridge
`, project, project, appDir, project, project, project, project, project, project, project, project, appDir, appDir, project, project, project, project)
}

func appEnvContent(project string, port int, nodeVer, hostRule string) string {
    if port == 0 { port = 4200 }
    if nodeVer == "" { nodeVer = "20" }
    if hostRule == "" { hostRule = fmt.Sprintf("Host('%s.localhost')", project) }
    return fmt.Sprintf(`# Environnement de l'application Angular SSR
APP_ENV=dev
NODE_VERSION=%s
PORT=%d
# Exemple: Host('%s.localhost') ou HostRegexp('%s.{domain}')
HOST_TRAEFIK_APP=%s
`, nodeVer, port, project, project, hostRule)
}

func appReadme(project, appDir string) string {
    return fmt.Sprintf(`# %s (Angular SSR Starter)

Ce dossier contient le code de l'application Angular.

Étapes recommandées:

1. Initialiser l'application Angular:
   npm create @angular@latest %s -- --ssr

2. (Optionnel) Ajuster la config SSR si besoin

3. Lancer en dev via Docker Compose:
   cd ../docker && APP_ENV=dev docker compose -f compose.yaml up --build

Notes:
- Les variables sont définies dans .env/.env.dist
- Le Dockerfile inclut les stages: base, dev, build, runtime (+ alias prod/preprod)
- Les fichiers compose: compose.yaml, compose.preprod.yaml, compose.prod.yaml
`, project, appDir)
}

func runAngularCreate(root string) error {
    // Pré-vérification outils
    hasNg := hasCommand("ng")
    hasNpx := hasCommand("npx")

    if !hasNg {
        fmt.Println("[info] Angular CLI 'ng' introuvable.")
        if hasNpx {
            fmt.Println("[info] Utilisation de 'npx @angular/cli@latest new app --ssr' en fallback.")
        } else {
            return errors.New("ni 'ng' ni 'npx' disponibles. Installez Node.js et Angular CLI: npm install -g @angular/cli")
        }
    }

    // Utilise Angular CLI directement si présent
    cmd := exec.Command("ng", "new", "app", "--ssr")
    cmd.Dir = root
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin

    if err := cmd.Run(); err != nil {
        // Fallback to npx Angular CLI si ng non présent
        if !hasNpx {
            return fmt.Errorf("échec 'ng new'. Et 'npx' n'est pas disponible. Installez Angular CLI: npm install -g @angular/cli (ou installez npx)")
        }
        fallback := exec.Command("npx", "-y", "@angular/cli@latest", "new", "app", "--ssr")
        fallback.Dir = root
        fallback.Stdout = os.Stdout
        fallback.Stderr = os.Stderr
        fallback.Stdin = os.Stdin
        if err2 := fallback.Run(); err2 != nil {
            return fmt.Errorf("'ng' et 'npx' ont échoué: %v / %v. Guide: installer Node.js >= 18 et Angular CLI: npm install -g @angular/cli", err, err2)
        }
    }
    return nil
}

func hasCommand(name string) bool {
    _, err := exec.LookPath(name)
    return err == nil
}

func hasDocker() bool {
    return hasCommand("docker")
}

func hasDockerCompose() (bool, bool) {
    // Returns (hasDockerComposeSubcommand, hasDockerComposeBinary)
    if !hasDocker() { return false, false }
    // Try docker compose
    cmd := exec.Command("docker", "compose", "version")
    if err := cmd.Run(); err == nil {
        return true, false
    }
    // Legacy docker-compose binary
    if hasCommand("docker-compose") {
        return false, true
    }
    return false, false
}

func chooseComposeCmd() string {
    hasSub, hasBin := hasDockerCompose()
    switch {
    case hasSub:
        return "docker compose"
    case hasBin:
        return "docker-compose"
    default:
        return "docker compose" // default hint
    }
}

func dockerNetworkExists(name string) bool {
    if !hasDocker() { return false }
    cmd := exec.Command("docker", "network", "inspect", name)
    return cmd.Run() == nil
}

func printDockerHints(project string) {
    network := fmt.Sprintf("%s-nseven", project)
    hasSub, hasBin := hasDockerCompose()

    if !hasDocker() {
        fmt.Println("[warn] Docker introuvable. Installez Docker Desktop / Docker Engine.")
    }
    if !hasSub && !hasBin {
        fmt.Println("[warn] 'docker compose' ou 'docker-compose' introuvable. Installez le plugin Compose ou utilisez Docker Desktop récent.")
    }
    if !dockerNetworkExists(network) {
        fmt.Printf("[info] Réseau externe '%s' absent. Créez-le avant de lancer: docker network create %s\n", network, network)
    }
}
