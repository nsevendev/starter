package stage1

import "fmt"

func DockerfileContent(nodeVer string) string {
    return fmt.Sprintf(`# -------------------------------------------------------------------
# Étape commune : base Node
# -------------------------------------------------------------------
ARG NODE_VERSION=%v
FROM node:${NODE_VERSION}-slim AS base
RUN apt-get update \
 && apt-get install -y --no-install-recommends bash \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
RUN npm install -g npm
RUN npm install -g @angular/cli

# -------------------------------------------------------------------
# Étape 1 : développement (montage du code, aucune copie au build)
# -------------------------------------------------------------------
FROM base AS dev
ENV NODE_ENV=development
RUN apt-get update && apt-get install -y \
    wget ca-certificates fonts-liberation libasound2 libatk1.0-0 libatk-bridge2.0-0 \
    libc6 libcairo2 libcups2 libdbus-1-3 libdrm2 libexpat1 libgbm1 libglib2.0-0 \
    libgtk-3-0 libnspr4 libnss3 libpango-1.0-0 libx11-6 libx11-xcb1 libxcb1 \
    libxcomposite1 libxcursor1 libxdamage1 libxext6 libxfixes3 libxi6 libxrandr2 \
    libxrender1 libxss1 libxtst6 xdg-utils chromium \
 && rm -rf /var/lib/apt/lists/*

ENV CHROME_BIN=/usr/bin/chromium
COPY package*.json ./
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["npm", "start"]

# -------------------------------------------------------------------
# Étape 2 : construction (préproduction et production)
# -------------------------------------------------------------------
FROM base AS build
# Important: Tailwind (via @tailwindcss/postcss) est une devDependency
# On force l'environnement en développement pour garantir son installation
ENV NODE_ENV=development
COPY package*.json ./
RUN npm ci
COPY . .
RUN npx ng build --configuration production

# -------------------------------------------------------------------
# Étape 3 : exécution (préproduction et production) — image identique
# -------------------------------------------------------------------
FROM node:${NODE_VERSION}-slim AS runtime
ENV NODE_ENV=production
WORKDIR /app
COPY --from=build /app/dist/app/server ./server
COPY --from=build /app/dist/app/browser ./browser
EXPOSE 4000
# Lancement du serveur de rendu côté serveur
CMD ["node", "server/server.mjs"]

# -------------------------------------------------------------------
# les stages prod et preprod utilisent la même image runtime
# -------------------------------------------------------------------
FROM runtime AS prod
FROM runtime AS preprod
`, nodeVer)
}
