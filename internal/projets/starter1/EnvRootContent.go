package starter1

import "fmt"

func EnvRootContent(port int, nodeVer, hostRule string) string {
	return fmt.Sprintf(`# dev, preprod, prod
APP_ENV=dev

# version de Node.js utilisée pour le projet (dockerfile)
NODE_VERSION=%v

HOST_TRAEFIK_APP=%v
PORT=%v
`, nodeVer, hostRule, port)
}
