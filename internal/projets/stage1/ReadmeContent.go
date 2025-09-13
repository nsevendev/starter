package stage1

import "fmt"

func ReadmeContent(nameApp string) string {
	return fmt.Sprintf(`# %v

- utilisation du projet taper la commande make

- app.dockerfile => dockerfile pour l'application angular
- compose.yaml => docker dev
- compose.preprod.yaml => docker preprod
- compose.prod.yaml => docker prod
`, nameApp)
}
