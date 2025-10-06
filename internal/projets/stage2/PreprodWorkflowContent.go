package stage2

import "fmt"

func PreprodWorkflowContent(nameFolderProject string) string {
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
  #test:
  #  name: Run tests
  #  runs-on: ubuntu-latest
  #
  #  steps:
  #    - name: Checkout code
  #      uses: actions/checkout@v4
  #
  #    - name: Create .env files from dist
  #      run: |
  #        # Créer les fichiers .env à partir des .env.dist
  #        find . -name "*.env.dist" -exec sh -c 'cp "$1" "${1%%.dist}"' _ {} \;
  #
  #    - name: Create Docker networks
  #      run: |
  #        # Créer le réseau traefik-nseven requis par docker-compose
  #        docker network create traefik-nseven || true
  #
  #    - name: Start services in dev mode
  #      run: |
  #        # Lancer le projet en mode dev
  #        make up
  #
  #        # Copie le fichier environment.dist et creer le fichier environment.ts
  #        # cp app/src/environments/environment.dist app/src/environments/environment.ts
  #
  #        # Attendre que les services soient prêts
  #        sleep 30
  #
  #    - name: Run frontend tests
  #      run: |
  #        # Lancer tous les tests frontend
  #        make tafc
  #
  #    - name: Check logs on failure
  #      if: failure()
  #      run: |
  #        echo "=== APP Logs ==="
  #        make lapp
  #
  #    - name: Cleanup
  #      if: always()
  #      run: |
  #        make down || true

  build_and_push:
    if: github.event_name == 'push'
    #needs: test
    name: Build & push image to Github registry
    runs-on: ubuntu-latest
    env:
      IMAGE_BASE: ghcr.io/${{ github.repository }}
      TAG_BASE: preprod
    strategy:
      matrix:
        service:
          - name: front
            context: ./front
            dockerfile: ./docker/front.dockerfile
            target: preprod
          - name: api
            context: ./api
            dockerfile: ./docker/api.dockerfile
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
            cd ~/preprod/%[1]v

            # Login GHCR (token GitHub avec scope packages:read côté serveur)
            echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u "${{ github.actor }}" --password-stdin

            export IMAGE_TAG="${{ env.IMAGE_TAG }}"
			export APP_ENV=preprod

            # Pull & up
            make down || true
            make deploy
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
            REPO="ghcr.io/${{ github.repository }}/front"
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
