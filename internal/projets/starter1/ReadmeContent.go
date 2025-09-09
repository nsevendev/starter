package starter1

import "fmt"

func ReadmeContent(project string) string {
	return fmt.Sprintf(`# %s
## Utilisation

- tapez "make" pour voir toutes les commandes disponibles

**(vous pouvez supprimer cette partie)**
## Aide  

- pour les tests creer des commandes adequate dans le package.json  
de l'application angular, creer un script "test" et "test:ci"  
elle sont deja relier dans le Makefile  

`, project)
}
