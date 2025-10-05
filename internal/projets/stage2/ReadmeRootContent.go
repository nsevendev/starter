package stage2

import "fmt"

func ReadmeContent(nameApp string) string {
	return fmt.Sprintf(`# %v

## Utilisation

- utilisation du projet: taper la commande make

## Contenu du projet

- front.dockerfile => dockerfile pour l'application astro
- api.dockerfile => dockerfile pour l'application astro
- compose.yaml => docker dev
- compose.preprod.yaml => docker preprod
- compose.prod.yaml => docker prod
- front/ => code astro projet front ssr
- api/ => code go projet api
- Makefile => commandes make

## indication CI

- preprod
preparer sur le server dans le dossier ~/preprod/%v avec le contenu suivant
.env
Makefile
docker/compose.preprod.yaml
docker/mongo-init
front/.env
api/.env

- prod
preparer sur le server dans le dossier ~/prod/%v avec le contenu suivant
.env
Makefile
docker/compose.preprod.yaml
docker/mongo-init
front/.env
api/.env

`, nameApp, nameApp, nameApp)
}
