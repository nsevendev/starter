package stage2

import "fmt"

func FrontDockerfileContent() string {
	return fmt.Sprintf(`# ---------- Base ----------
FROM node:22.19.0-slim AS base
RUN corepack enable && corepack prepare pnpm@latest --activate
RUN apt-get update && apt-get install -y bash && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# ---------- Dev ----------
FROM base AS dev
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["pnpm", "dev"]

# ---------- Build ----------
FROM base AS build
# Install TOUTES les dépendances pour le build (dev + prod)
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

# ---------- Dependencies production ----------
FROM base AS prod-deps
RUN pnpm install --frozen-lockfile

# ---------- Runtime ----------
FROM node:22.19.0-alpine AS runtime-base
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app

# Copier dist depuis build et node_modules depuis prod-deps
COPY --from=build /app/dist ./dist
COPY --from=prod-deps /app/node_modules ./node_modules
COPY --from=build /app/package.json ./

RUN addgroup -g 1001 -S nodejs && \
    adduser -S astro -u 1001 -G nodejs && \
    chown -R astro:nodejs /app

USER astro

ENV HOST=0.0.0.0
ENV PORT=3000
ENV NODE_ENV=production

EXPOSE 3000

CMD ["node", "./dist/server/entry.mjs"]

# ---------- Environnements spécifiques ----------
FROM runtime-base AS prod
ENV NODE_ENV=production

FROM runtime-base AS preprod
ENV NODE_ENV=preprod
`)
}
