package stage2

import "fmt"

func EnvRootContent(hostTraefikFront, hostTraefikApi string) string {
	return fmt.Sprintf(`# dev, preprod, prod à changer en fonction de l'environnement
APP_ENV=dev
# pour traefik dans le compose (à adapter en fonction de votre configuration locale)
HOST_TRAEFIK_FRONT=Host(`+"`%v`"+`)
HOST_TRAEFIK_API=Host(`+"`%v`"+`)
PORT=3000
# access externe database # a supprimer en prod ou preprod
DB_PORT_EX=27017
`, hostTraefikFront, hostTraefikApi)
}
