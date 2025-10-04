package stage2

import "fmt"

func ReadmeContent(nameApp string) string {
	return fmt.Sprintf(`# %v

- utilisation du projet: taper la commande make

- app.dockerfile => dockerfile pour l'application angular
- compose.yaml => docker dev
- compose.preprod.yaml => docker preprod
- compose.prod.yaml => docker prod

# indication CI
- pour le premier push sur main apres avoir creer le projet  
commenter les lignes de tous les fichiers CI comportant la section "on:"  
ensuite vous pourrez faire votre premier push ajuster le server pour correspondre au projet  
puis décommenter les lignes "on:" et faire un push pour que la CI se lance automatiquement

# erreur éventuelle (node 22.19.0 => angular cli 20.2.1)
(seront potentiellement réparer dans les prochaines versions d'angular)
- Error: NG0401:
Mettre ce code dans le fichier main.server.ts
(probleme de cli angular qui ne met pas le type BootstrapContext dans le main.server.ts)
~~~ts
import { bootstrapApplication, type BootstrapContext } from '@angular/platform-browser';
import { App } from './app/app';
import { config } from './app/app.config.server';

const bootstrap = (context: BootstrapContext) => bootstrapApplication(App, config, context);

export default bootstrap;
~~~

`, nameApp)
}
