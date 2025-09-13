package stage1

import "fmt"

func EntrypointShContent() string {
	return fmt.Sprintf(`
#!/bin/sh
set -e

if [ ! -d "node_modules" ] && [ -f "package-lock.json" ]; then
    echo "node_modules absent mais package-lock.json présent → installation avec npm ci"
    npm ci
elif [ ! -d "node_modules" ] && [ ! -f "package-lock.json" ]; then
    echo "node_modules absent et aucun package-lock.json → installation avec npm install"
    npm install
elif [ -d "node_modules" ] && [ ! -f "package-lock.json" ]; then
    echo "node_modules présent mais package-lock.json manquant → erreur"
    exit 1
else
    echo "node_modules et package-lock.json présents → rien à faire"
fi

exec "$@"
`)
}
