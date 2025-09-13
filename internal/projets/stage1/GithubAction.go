package stage1

import "fmt"

// GithubActionPreprodContent returns the GitHub Actions workflow (YAML)
// for preprod: run tests and deploy (no versioning).
// The tests run in a Node container and use `npm run test:ci` inside the app directory.
func GithubActionPreprodContent(nameFolderProject string) string {
	return fmt.Sprintf(`name: preprod

on:
  push:
    branches: ["preprod"]
  pull_request:
    branches: [ "preprod" ]
  workflow_dispatch:

concurrency:
  group: preprod-deploy
  cancel-in-progress: true

jobs:
  test:
    name: Run tests
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Create .env files from dist
        run: |
          # Créer les fichiers .env à partir des .env.dist
          find . -name "*.env.dist" -exec sh -c 'cp "$1" "${1%%.dist}"' _ {} \;

      - name: Create Docker networks
        run: |
          # Créer le réseau traefik-nseven requis par docker-compose
          docker network create traefik-nseven || true

      - name: Start services in dev mode
        run: |
          # Lancer le projet en mode dev
          make up
          
          # Copie le fichier environment.dist et creer le fichier environment.ts
          # cp app/src/environments/environment.dist app/src/environments/environment.ts
          
          # Attendre que les services soient prêts
          sleep 30

      - name: Run frontend tests
        run: |
          # Lancer tous les tests frontend
          make tafc

      - name: Check logs on failure
        if: failure()
        run: |
          echo "=== APP Logs ==="
          make lapp

      - name: Cleanup
        if: always()
        run: |
          make down || true

  build_and_push:
    if: github.event_name == 'push'
    needs: test
    name: Build & push image to Github registry
    runs-on: ubuntu-latest
    env:
      IMAGE_BASE: ghcr.io/${{ github.repository }}
      TAG_BASE: preprod
    strategy:
      matrix:
        service:
          - name: app
            context: ./app
            dockerfile: ./docker/app.dockerfile
            target: preprod
          # ajouter ici les différents services si besoin
    steps:
      - name: Compute tag timestamp (UTC)
        run: echo "TAG_TS=$(date -u +%%Y.%%m.%%d-%%H%%M)" >> $GITHUB_ENV

      - name: Checkout & setup
        uses: actions/checkout@v4

      - name: Set up QEMU
        uses: docker/setup-qemu-action@v3

      - name: Set up Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Github registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build & Push ${{ matrix.service.name }}
        uses: docker/build-push-action@v6
        with:
          context: ${{ matrix.service.context }}
          file: ${{ matrix.service.dockerfile }}
          target: ${{ matrix.service.target }}
          push: true
          platforms: linux/amd64
          build-args: |
            NODE_VERSION=22.19.0
          tags: |
            ${{ env.IMAGE_BASE }}/${{ matrix.service.name }}:${{ env.TAG_BASE }}
            ${{ env.IMAGE_BASE }}/${{ matrix.service.name }}:${{ env.TAG_BASE }}-${{ github.sha }}
            ${{ env.IMAGE_BASE }}/${{ matrix.service.name }}:preprod-${{ env.TAG_TS }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy:
    if: github.event_name == 'push'
    needs: build_and_push
    name: Deploy to Ionos Preprod
    runs-on: ubuntu-latest
    env:
      IMAGE_TAG: preprod-${{ github.sha }}
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup SSH key
        run: |
          mkdir -p ~/.ssh
          # Write SSH key with proper formatting
          printf '%%s\n' "${{ secrets.IONOS_SSH_KEY }}" > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa
          chmod 700 ~/.ssh
          
          # Debug: Check key format and size
          echo "SSH key file size:"
          wc -c ~/.ssh/id_rsa
          echo "SSH key first line:"
          head -1 ~/.ssh/id_rsa
          echo "SSH key last line:"
          tail -1 ~/.ssh/id_rsa
          
          # Test SSH key format
          ssh-keygen -l -f ~/.ssh/id_rsa
          
          # Add server to known hosts
          ssh-keyscan -H ${{ secrets.IONOS_HOST }} >> ~/.ssh/known_hosts
          
          # Test SSH connection with verbose output
          ssh -v -o ConnectTimeout=10 -o StrictHostKeyChecking=no ${{ secrets.IONOS_USER }}@${{ secrets.IONOS_HOST }} "echo 'SSH connection successful'"

      - name: Deploy
        run: |
          ssh ${{ secrets.IONOS_USER }}@${{ secrets.IONOS_HOST }} << 'EOF'
            set -e
            
            # Navigate to project directory
            cd ~/projects/test/%[1]v
          
            # Pull latest changes (safe even if already on preprod branch)
            git fetch origin
            git checkout preprod || git checkout -b preprod origin/preprod
            git pull origin preprod

            # Login GHCR (token GitHub avec scope packages:read côté serveur)
            echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u "${{ github.actor }}" --password-stdin
          
            export IMAGE_TAG="${{ env.IMAGE_TAG }}"

            # Pull & up
            make down || true
            make up
          EOF

  cleanup_preprod_images:
    name: Cleanup images (safe, preprod)
    needs: deploy
    runs-on: ubuntu-latest
    env:
      TAG_BASE: preprod
      KEEP_TAG: preprod-${{ github.sha }}
    steps:
      - name: Setup SSH key
        run: |
          mkdir -p ~/.ssh
          # Write SSH key with proper formatting
          printf '%%s\n' "${{ secrets.IONOS_SSH_KEY }}" > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa
          chmod 700 ~/.ssh

          # Debug: Check key format and size
          echo "SSH key file size:"
          wc -c ~/.ssh/id_rsa
          echo "SSH key first line:"
          head -1 ~/.ssh/id_rsa
          echo "SSH key last line:"
          tail -1 ~/.ssh/id_rsa

          # Test SSH key format
          ssh-keygen -l -f ~/.ssh/id_rsa

          # Add server to known hosts
          ssh-keyscan -H ${{ secrets.IONOS_HOST }} >> ~/.ssh/known_hosts

          # Test SSH connection with verbose output
          ssh -v -o ConnectTimeout=10 -o StrictHostKeyChecking=no ${{ secrets.IONOS_USER }}@${{ secrets.IONOS_HOST }} "echo 'SSH connection successful'"
      
      - name: Safe cleanup on server
        run: |
          ssh ${{ secrets.IONOS_USER }}@${{ secrets.IONOS_HOST }} << 'EOF'
            set -e
            REPO="ghcr.io/${{ github.repository }}/app"
            TAG_BASE="${TAG_BASE}"
            KEEP_REF="$REPO:${KEEP_TAG}"

            echo "== Safe cleanup starting =="
            echo "TAG_BASE=$TAG_BASE"
            echo "KEEP_REF=$KEEP_REF"

            # Liste des IDs d'images réellement utilisées par des conteneurs
            IN_USE_IDS=$(docker ps --format '{{.Image}}' | xargs -r docker inspect --format '{{.Image}}' 2>/dev/null | sort -u || true)
            echo "In-use image IDs:"
            echo "$IN_USE_IDS"

            # Prune léger: ne supprime que les couches orphelines
            docker image prune -f || true

            # Parcours des tags correspondants (prod-* ici), en évitant les images en cours d'utilisation et le tag courant
            docker image ls "$REPO" --format '{{.Repository}}:{{.Tag}} {{.ID}}' \
            | awk -v base="$TAG_BASE" '$1 ~ (":" base "-") {print $1, $2}' \
            | while read REF ID; do
                if [ "$REF" = "$KEEP_REF" ]; then
                  echo "Keep current tag: $REF"
                  continue
                fi
                if echo "$IN_USE_IDS" | grep -q "$ID"; then
                  echo "In use, skip: $REF ($ID)"
                  continue
                fi
                echo "Remove unused tag: $REF"
                docker rmi "$REF" || true
              done

            echo "== Safe cleanup done =="
          EOF
`, nameFolderProject)
}

// GithubActionProdContent returns the GitHub Actions workflow (YAML)
// for production: test, release (semantic-release) with changelog, tags, and a release branch.
// Triggered on push to main (PR merges land on main), and manually via dispatch.
func GithubActionProdContent(nameFolderProject string) string {
	return fmt.Sprintf(`name: prod

on:
  pull_request:
    branches: [ "main" ]
  push:
    branches: [ "main" ]

concurrency:
  group: main-pipeline
  cancel-in-progress: true

jobs:
  test:
    name: Run tests (main)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Create .env files from dist
        run: |
          find . -name "*.env.dist" -exec sh -c 'cp "$1" "${1%%.dist}"' _ {} \;

      - name: Create Docker networks
        run: docker network create traefik-nseven || true

      - name: Start services in dev mode
        run: |
          make up
          sleep 30

      - name: Run frontend tests
        run: make tafc

      - name: Check logs on failure
        if: failure()
        run: |
          echo "=== APP Logs ==="
          make lapp

      - name: Cleanup
        if: always()
        run: make down || true

  release:
    if: github.event_name == 'push'
    needs: test
    name: Semantic release (create tag & GitHub Release)
    runs-on: ubuntu-latest
    outputs:
      published: ${{ steps.semrel.outputs.new_release_published }}
      tag: ${{ steps.semrel.outputs.new_release_git_tag }}
      version: ${{ steps.semrel.outputs.new_release_version }}
    permissions:
      contents: write
      issues: write
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0   # requis par semantic-release

      - uses: actions/setup-node@v4
        with:
          node-version: 22.19.0

      - name: Semantic Release
        id: semrel
        uses: cycjimmy/semantic-release-action@v4
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          dry_run: false
          # installe les plugins/presets nécessaires avant executer semantic-release
          extra_plugins: |
            conventional-changelog-conventionalcommits

      - name: Install semantic-release
        run: npm i -D semantic-release @semantic-release/changelog @semantic-release/git @semantic-release/github conventional-changelog-conventionalcommits

      - name: Run semantic-release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: npx semantic-release

  build_and_push_prod:
    needs: release
    if: needs.release.outputs.published == 'true'
    runs-on: ubuntu-latest
    env:
      IMAGE_BASE: ghcr.io/${{ github.repository }}
      TAG_BASE: prod
      VERSION: ${{ needs.release.outputs.tag }}   # ex: v1.2.3
    strategy:
      matrix:
        service:
          - name: app
            context: ./app
            dockerfile: ./docker/app.dockerfile
            target: prod
            # ajoute d'autres services ici si besoin
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ env.VERSION }}   # on build exactement le code de la release taguée

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - uses: docker/setup-qemu-action@v3
      - uses: docker/setup-buildx-action@v3

      - name: Compute tag timestamp (UTC)
        run: echo "TAG_TS=$(date -u +%%Y.%%m.%%d-%%H%%M)" >> $GITHUB_ENV

      - name: Build & Push ${{ matrix.service.name }}
        uses: docker/build-push-action@v6
        with:
          context: ${{ matrix.service.context }}
          file: ${{ matrix.service.dockerfile }}
          target: ${{ matrix.service.target }}
          push: true
          platforms: linux/amd64
          build-args: |
            NODE_VERSION=22.19.0
          tags: |
            ${{ env.IMAGE_BASE }}/${{ matrix.service.name }}:${{ env.TAG_BASE }}
            ${{ env.IMAGE_BASE }}/${{ matrix.service.name }}:${{ env.TAG_BASE }}-${{ env.VERSION }}
            ${{ env.IMAGE_BASE }}/${{ matrix.service.name }}:${{ env.TAG_BASE }}-${{ github.sha }}
            ${{ env.IMAGE_BASE }}/${{ matrix.service.name }}:${{ env.TAG_BASE }}-${{ env.TAG_TS }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  # 4) Déploiement PROD (ne s’exécute que s’il y a une release)
  deploy_prod:
    needs: [build_and_push_prod, release]
    if: needs.release.outputs.published == 'true'
    runs-on: ubuntu-latest
    env:
      IMAGE_TAG: prod-${{ needs.release.outputs.tag }}  # prod-vX.Y.Z
    steps:
      - name: Setup SSH key
        run: |
          mkdir -p ~/.ssh
          # Write SSH key with proper formatting
          printf '%%s\n' "${{ secrets.IONOS_SSH_KEY }}" > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa
          chmod 700 ~/.ssh

          # Debug: Check key format and size
          echo "SSH key file size:"
          wc -c ~/.ssh/id_rsa
          echo "SSH key first line:"
          head -1 ~/.ssh/id_rsa
          echo "SSH key last line:"
          tail -1 ~/.ssh/id_rsa

          # Test SSH key format
          ssh-keygen -l -f ~/.ssh/id_rsa

          # Add server to known hosts
          ssh-keyscan -H ${{ secrets.IONOS_HOST }} >> ~/.ssh/known_hosts

          # Test SSH connection with verbose output
          ssh -v -o ConnectTimeout=10 -o StrictHostKeyChecking=no ${{ secrets.IONOS_USER }}@${{ secrets.IONOS_HOST }} "echo 'SSH connection successful'"

      - name: Deploy on server
        run: |
          ssh ${{ secrets.IONOS_USER }}@${{ secrets.IONOS_HOST }} << 'EOF'
            set -e
            cd ~/projects/prod/%[1]v

            # pull le code main (si tu gardes des fichiers compose/*.yaml dans le repo)
            git fetch origin
            git checkout main || git checkout -b main origin/main
            git pull origin main

            echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u "${{ github.actor }}" --password-stdin

            export IMAGE_TAG="${{ env.IMAGE_TAG }}"

            make down || true
            make up
          EOF

  cleanup_prod_images:
    name: Cleanup images (safe, prod)
    needs: [deploy_prod, release]
    if: needs.release.outputs.published == 'true'
    runs-on: ubuntu-latest
    env:
      TAG_BASE: prod
      KEEP_TAG: prod-${{ needs.release.outputs.tag || needs.release.outputs.version }}
    steps:
      - name: Setup SSH key
        run: |
          mkdir -p ~/.ssh
          # Write SSH key with proper formatting
          printf '%%s\n' "${{ secrets.IONOS_SSH_KEY }}" > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa
          chmod 700 ~/.ssh

          # Debug: Check key format and size
          echo "SSH key file size:"
          wc -c ~/.ssh/id_rsa
          echo "SSH key first line:"
          head -1 ~/.ssh/id_rsa
          echo "SSH key last line:"
          tail -1 ~/.ssh/id_rsa

          # Test SSH key format
          ssh-keygen -l -f ~/.ssh/id_rsa

          # Add server to known hosts
          ssh-keyscan -H ${{ secrets.IONOS_HOST }} >> ~/.ssh/known_hosts

          # Test SSH connection with verbose output
          ssh -v -o ConnectTimeout=10 -o StrictHostKeyChecking=no ${{ secrets.IONOS_USER }}@${{ secrets.IONOS_HOST }} "echo 'SSH connection successful'"

      - name: Safe cleanup on server
        run: |
          ssh ${{ secrets.IONOS_USER }}@${{ secrets.IONOS_HOST }} << 'EOF'
            set -e
            REPO="ghcr.io/${{ github.repository }}/app"
            TAG_BASE="${TAG_BASE}"
            KEEP_REF="$REPO:${KEEP_TAG}"

            echo "== Safe cleanup starting =="
            echo "TAG_BASE=$TAG_BASE"
            echo "KEEP_REF=$KEEP_REF"

            # Liste des IDs d'images réellement utilisées par des conteneurs
            IN_USE_IDS=$(docker ps --format '{{.Image}}' | xargs -r docker inspect --format '{{.Image}}' 2>/dev/null | sort -u || true)
            echo "In-use image IDs:"
            echo "$IN_USE_IDS"

            # Prune léger: ne supprime que les couches orphelines
            docker image prune -f || true

            # Parcours des tags correspondants (prod-* ici), en évitant les images en cours d'utilisation et le tag courant
            docker image ls "$REPO" --format '{{.Repository}}:{{.Tag}} {{.ID}}' \
            | awk -v base="$TAG_BASE" '$1 ~ (":" base "-") {print $1, $2}' \
            | while read REF ID; do
                if [ "$REF" = "$KEEP_REF" ]; then
                  echo "Keep current tag: $REF"
                  continue
                fi
                if echo "$IN_USE_IDS" | grep -q "$ID"; then
                  echo "In use, skip: $REF ($ID)"
                  continue
                fi
                echo "Remove unused tag: $REF"
                docker rmi "$REF" || true
              done

            echo "== Safe cleanup done =="
          EOF
`, nameFolderProject)
}

// SemanticReleaseConfigContent returns a .releaserc.json tailored for prod
// so semantic-release updates CHANGELOG.md and package files inside the app folder.
func SemanticReleaseConfigContent(project string) string {
	return fmt.Sprintf(`{
  "branches": ["main"],
  "tagFormat": "v${version}",
  "plugins": [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    ["@semantic-release/changelog", {"changelogFile": "CHANGELOG.md"}],
    "@semantic-release/npm",
    ["@semantic-release/git", {
      "assets": [
        "CHANGELOG.md",
        "%[1]s/package.json",
        "%[1]s/package-lock.json"
      ],
      "message": "chore(release): ${nextRelease.version} [skip ci]\n\n${nextRelease.notes}"
    }],
    "@semantic-release/github"
  ]
}
`, project)
}

// GithubActionPRValidationContent returns the GitHub Actions workflow (YAML)
// to validate PRs with lint/build/test in the app directory.
func GithubActionPRValidationContent(project string) string {
	return fmt.Sprintf(`name: PR Validation

on:
  pull_request:
  workflow_dispatch:

concurrency:
  group: pr-validation-${{ github.ref }}
  cancel-in-progress: true

jobs:
  validate:
    runs-on: ubuntu-latest
    container: node:20
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install dependencies
        working-directory: ./%[1]s
        run: npm ci

      - name: Lint (if present)
        working-directory: ./%[1]s
        run: npm run -s lint --if-present

      - name: Build (if present)
        working-directory: ./%[1]s
        run: npm run -s build --if-present

      - name: Test (prefer CI script)
        working-directory: ./%[1]s
        run: |
          npm run -s test:ci --if-present || npm test
`, project)
}

/*

	return fmt.Sprintf(`name: CI/CD Preprod

on:
  push:
    branches: ["preprod"]
  workflow_dispatch:

concurrency:
  group: preprod-deploy
  cancel-in-progress: true

jobs:
  test:
    name: Run tests
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Create .env files from dist
        run: |
          # Créer les fichiers .env à partir des .env.dist
          find . -name "*.env.dist" -exec sh -c 'cp "$1" "${1%%.dist}"' _ {} \;

      - name: Create Docker networks
        run: |
          # Créer le réseau %v-nseven requis par docker-compose
          docker network create %v-nseven || true

      - name: Start services in dev mode
        run: |
          # Lancer le projet en mode dev
          make up

          # Copie le fichier environment.dist et creer le fichier environment.ts
          # cp app/src/environments/environment.dist app/src/environments/environment.ts

          # Attendre que les services soient prêts
          sleep 30

      - name: Run frontend tests
        run: |
          # Lancer tous les tests frontend
          make tafc

      - name: Check logs on failure
        if: failure()
        run: |
          echo "=== APP Logs ==="
          make lapp

      - name: Cleanup
        if: always()
        run: |
          make down || true

  deploy:
    needs: test
    name: Deploy to Ionos Preprod
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup SSH key
        run: |
          mkdir -p ~/.ssh
          # Write SSH key with proper formatting
          printf '%%s\n' "${{ secrets.IONOS_SSH_KEY }}" > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa
          chmod 700 ~/.ssh

          # Debug: Check key format and size
          echo "SSH key file size:"
          wc -c ~/.ssh/id_rsa
          echo "SSH key first line:"
          head -1 ~/.ssh/id_rsa
          echo "SSH key last line:"
          tail -1 ~/.ssh/id_rsa

          # Test SSH key format
          ssh-keygen -l -f ~/.ssh/id_rsa

          # Add server to known hosts
          ssh-keyscan -H ${{ secrets.IONOS_HOST }} >> ~/.ssh/known_hosts

          # Test SSH connection with verbose output
          ssh -v -o ConnectTimeout=10 -o StrictHostKeyChecking=no ${{ secrets.IONOS_USER }}@${{ secrets.IONOS_HOST }} "echo 'SSH connection successful'"

      - name: Deploy to server
        run: |
          ssh ${{ secrets.IONOS_USER }}@${{ secrets.IONOS_HOST }} << 'EOF'
            set -e

            # Navigate to project directory
            cd ~/projects/test/%v

            # Pull latest changes (safe even if already on preprod branch)
            git fetch origin
            git checkout preprod || git checkout -b preprod origin/preprod
            git pull origin preprod

         #[ -f app/src/environments/environment.ts ] || cp app/src/environments/environment.dist app/src/environments/environment.ts

            # Deploy with Docker
            make down || true
            make upb

            # Verify deployment
            sleep 10
            docker compose --env-file .env -f docker/compose.preprod.yaml logs --tail=20 app

          EOF

`, appDir, project, appDir)

*/
