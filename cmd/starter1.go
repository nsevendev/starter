package cmd

import (
	"errors"
	"fmt"
	"github.com/nsevendev/starter/internal/docker"
	"github.com/nsevendev/starter/internal/projets/starter1"
	"github.com/nsevendev/starter/internal/tools"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	starter1Name     string
	starter1Port     int
	starter1NodeVer  string
	starter1HostRule string
)

func init() {
	rootCmd.AddCommand(starter1Cmd)

	// Flags
	starter1Cmd.Flags().StringVarP(&starter1Name, "name", "n", "app", "Nom du projet (app par défaut)")
	starter1Cmd.Flags().IntVarP(&starter1Port, "port", "p", 4200, "Port HTTP de l'app Angular")
	starter1Cmd.Flags().StringVar(&starter1NodeVer, "node", "22.16.0", "Version majeure de Node.js utilisée dans les images")
	starter1Cmd.Flags().StringVar(&starter1HostRule, "host", "", "Règle Traefik (par defaut: Host(`app.localhost`))")
}

// starter1Cmd represents the command to create the Angular SSR starter
var starter1Cmd = &cobra.Command{
	Use:   "starter-1",
	Short: "Crée un projet Angular SSR",
	Long:  "Génère un projet Angular SSR sans backend ni base de données, avec Dockerfile multi-stage et compose (dev/preprod/prod). Ne fournissez pas de nom de projet, tout est paramètré par défaut avec le nom app",
	RunE: func(cmd *cobra.Command, args []string) error {
		project := tools.SanitizeName(starter1Name)
		if project == "" {
			return errors.New("nom de projet invalide")
		}

		// recuperation du nom de dossier courant
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("erreur récupération du dossier courant: %w", err)
		}
		cwdName := filepath.Base(wd)

		appDir := filepath.Join(wd, project)

		// pas de host renseigné, propose host par defaut
		if !cmd.Flags().Changed("host") {
			def := fmt.Sprintf("Host(`%v.localhost`)", project)
			fmt.Println("- Aucun host renseigné -")
			if tools.AskYesNo(fmt.Sprintf("  Appliquer le host par défaut => %v ? [o/N]: ", def), true) {
				starter1HostRule = def
				fmt.Printf("- Host %v appliquée -\n", starter1HostRule)
			} else {
				return errors.New("commande annulée: aucun host défini")
			}
		} else {
			fmt.Printf("- Application du host: %v -\n", starter1HostRule)
			def := fmt.Sprintf("Host(`%v`)", starter1HostRule)
			starter1HostRule = def
		}

		// pas de version node renseignée, propose version par defaut
		if !cmd.Flags().Changed("node") {
			fmt.Println("- Aucune version node renseigné -")
			if tools.AskYesNo(fmt.Sprintf("  Appliquer la version de node par défaut => %v ? [o/N]: ", starter1NodeVer), true) {
				fmt.Printf("- Version %v appliquée -\n", starter1NodeVer)
			} else {
				return errors.New("commande annulée: aucune version de node défini")
			}
		}

		// pas de port renseigné, propose port par defaut
		if !cmd.Flags().Changed("port") {
			fmt.Println("- Aucun port renseigné -")
			if tools.AskYesNo(fmt.Sprintf("  Appliquer le port par défaut => %v ? [o/N]: ", starter1Port), true) {
				fmt.Printf("- Port %v appliquée -\n", starter1Port)
			} else {
				return errors.New("commande annulée: aucun port défini")
			}
		}

		fmt.Println("Starter-1: création du projet Angular SSR")
		fmt.Printf("- Dossier du projet: %v\n", cwdName)
		fmt.Printf("- Dossier de l'application: %v\n", project)
		fmt.Printf("- Host Traefik: %v\n", starter1HostRule)
		fmt.Printf("- Version de node demandé: %v\n", starter1NodeVer)
		fmt.Printf("- Port de l'app: %v\n", starter1Port)

		// validation des données de creation
		if tools.AskYesNo(fmt.Sprintf("  Est ce que ses valeurs vous conviennent ? [o/N]: "), true) {
			fmt.Printf("------ Initialisation du projet ------\n")
		} else {
			return errors.New("commande annulée: les valeurs définis ne conviennent pas")
		}

		// creation du projet angular
		fmt.Printf("- Lancement Angular CLI dans %s: ng new %s --ssr \n", project, project)
		if err := tools.RunAngularCreate(project, wd); err != nil {
			return fmt.Errorf("échec de la création Angular: %w", err)
		}

		fmt.Println("- [OK] Application Angular -")

		// creation du dossier docker
		dockerDir := filepath.Join("docker")
		if err := tools.EnsureDir(dockerDir); err != nil {
			return err
		}

		fmt.Println("- [OK] Dossier docker -")

		// creation du dockerfile app
		dockerfilePath := filepath.Join(dockerDir, "Dockerfile")
		if err := tools.WriteFileIfAbsent(dockerfilePath, starter1.DockerfileContent(starter1NodeVer)); err != nil {
			return err
		}

		fmt.Println("- [OK] docker/Dockerfile du projet angular -")

		// creation des fichiers compose
		baseCompose := filepath.Join(dockerDir, "compose.yaml")
		preprodCompose := filepath.Join(dockerDir, "compose.preprod.yaml")
		prodCompose := filepath.Join(dockerDir, "compose.prod.yaml")
		if err := tools.WriteFileIfAbsent(baseCompose, starter1.ComposeContent(project, cwdName)); err != nil {
			return err
		}
		fmt.Println("- [OK] docker/compose.yaml du projet -")

		if err := tools.WriteFileIfAbsent(preprodCompose, starter1.ComposePreprodContent(project, cwdName)); err != nil {
			return err
		}
		fmt.Println("- [OK] docker/compose.preprod.yaml du projet angular -")

		if err := tools.WriteFileIfAbsent(prodCompose, starter1.ComposeProdContent(project, cwdName)); err != nil {
			return err
		}
		fmt.Println("- [OK] docker/compose.prod.yaml du projet angular -")

		// creation des env root
		env := filepath.Join(".env")
		envDist := filepath.Join(".env.dist")
		if err := tools.WriteFileIfAbsent(env, starter1.EnvRootContent(starter1Port, starter1NodeVer, starter1HostRule)); err != nil {
			return err
		}
		if err := tools.WriteFileIfAbsent(envDist, starter1.EnvRootContent(starter1Port, starter1NodeVer, starter1HostRule)); err != nil {
			return err
		}

		fmt.Println("- [OK] .env root -")

		// creation des env app
		appEnv := filepath.Join(appDir, ".env")
		appEnvDist := filepath.Join(appDir, ".env.dist")
		if err := tools.WriteFileIfAbsent(appEnv, starter1.EnvAppContent()); err != nil {
			return err
		}
		if err := tools.WriteFileIfAbsent(appEnvDist, starter1.EnvAppContent()); err != nil {
			return err
		}

		fmt.Println("- [OK] app/.env -")

		// ceration du readme
		readme := filepath.Join("README.md")
		if err := tools.WriteFileAlways(readme, starter1.ReadmeContent(project)); err != nil {
			return err
		}

		fmt.Println("- [OK] README.md -")

		// creation du makefile
		makefile := filepath.Join("Makefile")
		if err := tools.WriteFileIfAbsent(makefile, starter1.MakefileContent()); err != nil {
			return err
		}

		fmt.Println("- [OK] Makefile -")

		// creation du makefile
		entrypoint := filepath.Join(appDir, "entrypoint.sh")
		if err := tools.WriteFileIfAbsent(entrypoint, starter1.EntrypointShContent()); err != nil {
			return err
		}

		fmt.Println("- [OK] app/entrypoint.sh -")

		fmt.Println("- Fichiers générés:")
		fmt.Printf("  %s\n", dockerfilePath)
		fmt.Printf("  %s\n", baseCompose)
		fmt.Printf("  %s\n", preprodCompose)
		fmt.Printf("  %s\n", prodCompose)
		fmt.Printf("  %s\n", env)
		fmt.Printf("  %s\n", envDist)
		fmt.Printf("  %s\n", appEnv)
		fmt.Printf("  %s\n", appEnvDist)
		fmt.Printf("  %s\n", readme)
		fmt.Printf("  %s\n", makefile)
		fmt.Printf("  %s\n", entrypoint)

		fmt.Println()

		docker.PrintDockerHints(project)

		fmt.Println("Next steps:")

		composeCmd := docker.ChooseComposeCmd()
		fmt.Printf("- cd %s/docker && APP_ENV=dev %s -f compose.yaml up --build\n", project, composeCmd)
		return nil
	},
}
