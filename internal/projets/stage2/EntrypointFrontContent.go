package stage2

import "fmt"

func EntrypointFrontContent() string {
	return fmt.Sprintf(`#!/bin/sh
set -e

echo "Démarrage du conteneur Astro..."

# Vérifier si node_modules existe
if [ ! -d "node_modules" ]; then
    echo "node_modules absent → installation avec pnpm..."

    if [ -f "pnpm-lock.yaml" ]; then
        echo "pnpm-lock.yaml trouvé → pnpm install --frozen-lockfile"
        pnpm install --frozen-lockfile
    else
        echo "pnpm-lock.yaml absent → pnpm install"
        pnpm install
    fi

    echo "Installation terminée"
else
    echo "node_modules déjà présent"
fi

exec "$@"
`)
}
