package starter1

import "fmt"

// GithubActionPreprodStarter1Content returns the GitHub Actions workflow (YAML)
// for preprod: run tests and deploy (no versioning).
// The tests run in a Node container and use `npm run test:ci` inside the app directory.
func GithubActionPreprodStarter1Content(project string) string {
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
    container: node:20
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install dependencies
        working-directory: ./%[1]s
        run: npm ci

      - name: Run tests (CI)
        working-directory: ./%[1]s
        run: npm run test:ci

  deploy:
    name: Deploy to Preprod
    needs: test
    runs-on: ubuntu-latest
    steps:
      - name: Setup SSH key
        run: |
          mkdir -p ~/.ssh
          printf '%s\n' "${{ secrets.PREPROD_SSH_KEY }}" > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa
          ssh-keyscan -H "${{ secrets.PREPROD_HOST }}" >> ~/.ssh/known_hosts

      - name: Deploy via SSH
        run: |
          ssh -o StrictHostKeyChecking=no "${{ secrets.PREPROD_USER }}@${{ secrets.PREPROD_HOST }}" << 'EOF'
            set -e
            cd "${{ secrets.PREPROD_PROJECT_PATH }}"
            git fetch origin
            git checkout preprod || git checkout -b preprod origin/preprod
            git pull --ff-only origin preprod
            APP_ENV=preprod docker compose --env-file .env -f docker/compose.preprod.yaml up -d --build
          EOF
`, project)
}

// GithubActionProdStarter1Content returns the GitHub Actions workflow (YAML)
// for production: test, release (semantic-release) with changelog, tags, and a release branch.
// Triggered on push to main (PR merges land on main), and manually via dispatch.
func GithubActionProdStarter1Content(project string) string {
    return fmt.Sprintf(`name: Release to Prod

on:
  push:
    branches: ["main"]
  workflow_dispatch:

concurrency:
  group: prod-release
  cancel-in-progress: false

jobs:
  test:
    name: Run tests
    runs-on: ubuntu-latest
    container: node:20
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install dependencies
        working-directory: ./%[1]s
        run: npm ci

      - name: Run tests (CI)
        working-directory: ./%[1]s
        run: npm run test:ci

  release:
    name: Semantic Release
    needs: test
    runs-on: ubuntu-latest
    outputs:
      new_release_published: ${{ steps.sr.outputs.new_release_published }}
      new_release_version: ${{ steps.sr.outputs.new_release_version }}
    permissions:
      contents: write
      issues: write
      pull-requests: write
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - id: sr
        name: Semantic Release
        uses: cycjimmy/semantic-release-action@v4
        with:
          branch: main
          tag_format: 'v${version}'
          extra_plugins: |
            @semantic-release/commit-analyzer
            @semantic-release/release-notes-generator
            @semantic-release/changelog
            @semantic-release/git
            @semantic-release/github
            @semantic-release/npm
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Create release branch
        if: steps.sr.outputs.new_release_published == 'true'
        run: |
          VERSION="${{ steps.sr.outputs.new_release_version }}"
          git config user.name "github-actions"
          git config user.email "github-actions@github.com"
          git fetch origin
          git checkout main
          git pull --ff-only origin main
          git branch "release/v${VERSION}"
          git push origin "release/v${VERSION}"

  deploy:
    name: Deploy to Prod
    needs: release
    if: needs.release.outputs.new_release_published == 'true'
    runs-on: ubuntu-latest
    steps:
      - name: Setup SSH key
        run: |
          mkdir -p ~/.ssh
          printf '%s\n' "${{ secrets.PROD_SSH_KEY }}" > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa
          ssh-keyscan -H "${{ secrets.PROD_HOST }}" >> ~/.ssh/known_hosts

      - name: Deploy via SSH
        run: |
          ssh -o StrictHostKeyChecking=no "${{ secrets.PROD_USER }}@${{ secrets.PROD_HOST }}" << 'EOF'
            set -e
            cd "${{ secrets.PROD_PROJECT_PATH }}"
            git fetch origin
            git checkout main || git checkout -b main origin/main
            git pull --ff-only origin main
            APP_ENV=prod docker compose --env-file .env -f docker/compose.prod.yaml up -d --build
          EOF
`, project)
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
` , project)
}
