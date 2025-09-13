package stage1

import "fmt"

func EnvRootContent(globalPortTraefik int, nodeVersion, hostTraefik string) string {
	return fmt.Sprintf(`# dev, preprod, prod
APP_ENV=dev

NODE_VERSION=%v

HOST_TRAEFIK_APP=%v

PORT=%v
`, nodeVersion, hostTraefik, globalPortTraefik)
}
