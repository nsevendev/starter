package stage2

import "fmt"

func EnvFrontContent() string {
	return fmt.Sprintf(`# modifier selon environement, dev, preprod, prod)
APP_ENV=dev
`)
}
