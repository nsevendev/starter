package starter1

import "fmt"

func EnvAppContent() string {
	return fmt.Sprintf(`# dev, preprod, prod
APP_ENV=dev
`)
}
