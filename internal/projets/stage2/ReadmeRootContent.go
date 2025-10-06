package stage2

import "fmt"

func ReadmeContent(nameApp string) string {
	return fmt.Sprintf(`# %v

## Prérequis

- avant le premier push en preprod ou prod quand il n'y a pas encore de code  
commenter les parties test dans les CI

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
docker/compose.prod.yaml
docker/mongo-init
front/.env
api/.env

`, nameApp, nameApp, nameApp)
}
