package cmd

import (
	"errors"
	"fmt"
	"github.com/nsevendev/starter/internal/docker"
	"github.com/nsevendev/starter/internal/projets/framework"
	"github.com/nsevendev/starter/internal/projets/stage1"
	"github.com/nsevendev/starter/internal/tools"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	hostTraefik string
	repoGit     string
	allowedHost []string
)

func init() {
    rootCmd.AddCommand(starter1Cmd)

    // Flags
    starter1Cmd.Flags().StringVar(&hostTraefik, "host", "", "host traefik => format: Host(``) (requis) ")
    _ = starter1Cmd.MarkFlagRequired("host")
    starter1Cmd.Flags().StringVar(&repoGit, "repo", "", "npm du repository git (requis) ")
    _ = starter1Cmd.MarkFlagRequired("repo")
    // allowedhost doit être une liste et alimenter la variable allowedHost
    starter1Cmd.Flags().StringSliceVar(&allowedHost, "allowedhost", nil, "allowed host pour angular.json (requis)")
    _ = starter1Cmd.MarkFlagRequired("allowedhost")
}

// starter1Cmd represents the command to create the Angular SSR starter
var starter1Cmd = &cobra.Command{
	Use:   "stage-1-22.19.0",
	Short: "Crée un projet Angular SSR",
	Long: `Génère un projet Angular SSR sans backend ni base de données, avec Dockerfile multi-stage et compose (dev/preprod/prod). 
	Cette a besoin de savoir le nom du repository git`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nameApp := "app"
		nodeVersion := "22.19.0"
		globalPortTraefik := 3000
		appPort := "4000"
		hostTraefik = fmt.Sprintf("Host(`%v`)", hostTraefik)

		// recuperation du nom de dossier courant
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("erreur récupération du dossier courant: %w", err)
		}

		nameFolderProject := filepath.Base(pwd)
		pathFolderApp := filepath.Join(pwd, nameApp)

		fmt.Println("Stage-1: création du projet Angular SSR avec ses données")
		fmt.Printf("- Path du projet: %v\n", pwd)
		fmt.Printf("- Dossier du projet: %v\n", nameFolderProject)
		fmt.Printf("- Path de l'app: %v\n", pathFolderApp)
		fmt.Printf("- Dossier de l'app: %v\n", nameApp)
		fmt.Printf("- Host du traefik: %v\n", hostTraefik)
		fmt.Printf("- Version de node: %v\n", nodeVersion)
		fmt.Printf("- Port pour tout les services traefik: %v\n", globalPortTraefik)
		fmt.Printf("- Port de l'app: %v\n", appPort)

		// validation des données de creation
		if tools.AskYesNo(fmt.Sprintf("  Est ce que ses valeurs vous conviennent ? [o/N]: "), true) {
			fmt.Printf("------ Initialisation du projet ------\n")
		} else {
			return errors.New("commande annulée: les valeurs définis ne conviennent pas")
		}

		// creation du projet angular
		fmt.Println(" - CLI angular va etre executer, vérifier d'avoir angular CLI d'installer avec une version node 22.19.0 - ")
		fmt.Println(" - Si ce n'est pas le cas il y a des risques de conflit avec la version docker qui sera créer - ")
		if tools.AskYesNo(fmt.Sprintf("  Est ce que vous voulez continuer ? [o/N]: "), true) {
			fmt.Printf("- Lancement Angular CLI dans %s: ng new %s --ssr \n", nameApp, nameApp)
		} else {
			return errors.New("commande annulée: les valeurs définis ne conviennent pas")
		}
		if err := framework.RunAngularSsrCreate(nameApp, pwd); err != nil {
			return fmt.Errorf("échec de la création Angular: %w", err)
		}
		fmt.Println("- [OK] Création application Angular -")

		// creation du dossier docker
		dockerDir := filepath.Join("docker")
		{
			if err := tools.EnsureDir(dockerDir); err != nil {
				fmt.Println("- [KO] Création dossier docker -")
			} else {
				fmt.Println("- [OK] Création dossier docker -")
			}
		}

		// creation du dockerfile app
		dockerfilePath := filepath.Join(dockerDir, "app.dockerfile")
		{
			if err := tools.WriteFileIfAbsent(dockerfilePath, stage1.DockerfileContent(nodeVersion)); err != nil {
				fmt.Println("- [KO] création docker/app.dockerfile du projet app -")
			} else {
				fmt.Println("- [OK] création docker/app.dockerfile du projet app -")
			}
		}

		// creation des fichiers compose
		composeBase := filepath.Join(dockerDir, "compose.yaml")
		{
			if err := tools.WriteFileIfAbsent(composeBase, stage1.ComposeContent(nameApp, nameFolderProject)); err != nil {
				fmt.Println("- [KO] création docker/compose.yaml du projet -")
			} else {
				fmt.Println("- [OK] création docker/compose.yaml du projet -")
			}
		}

		composePreprod := filepath.Join(dockerDir, "compose.preprod.yaml")
		{
			if err := tools.WriteFileIfAbsent(composePreprod, stage1.ComposePreprodContent(nameApp, nameFolderProject)); err != nil {
				fmt.Println("- [KO] création docker/compose.preprod.yaml du projet -")
			} else {
				fmt.Println("- [OK] création docker/compose.preprod.yaml du projet -")
			}
		}

		composeProd := filepath.Join(dockerDir, "compose.prod.yaml")
		{
			if err := tools.WriteFileIfAbsent(composeProd, stage1.ComposeProdContent(nameApp, nameFolderProject)); err != nil {
				fmt.Println("- [KO] création docker/compose.prod.yaml du projet angular -")
			} else {
				fmt.Println("- [OK] création docker/compose.prod.yaml du projet angular -")
			}
		}

		// creation des env root
		env := filepath.Join(".env")
		envDist := filepath.Join(".env.dist")
		{
			if err := tools.WriteFileIfAbsent(env, stage1.EnvRootContent(globalPortTraefik, nodeVersion, hostTraefik)); err != nil {
				fmt.Println("- [KO] création .env root -")
			} else {
				fmt.Println("- [OK] création .env root -")
			}

			if err := tools.WriteFileIfAbsent(envDist, stage1.EnvRootContent(globalPortTraefik, nodeVersion, hostTraefik)); err != nil {
				fmt.Println("- [KO] création .env.dist root -")
			} else {
				fmt.Println("- [OK] création .env.dist root -")
			}
		}

		// creation des env app
		appEnv := filepath.Join(pathFolderApp, ".env")
		appEnvDist := filepath.Join(pathFolderApp, ".env.dist")
		{
			if err := tools.WriteFileIfAbsent(appEnv, stage1.EnvAppContent()); err != nil {
				fmt.Println("- [KO] création app/.env -")
			} else {
				fmt.Println("- [OK] création app/.env -")
			}

			if err := tools.WriteFileIfAbsent(appEnvDist, stage1.EnvAppContent()); err != nil {
				fmt.Println("- [KO] création app/.env.dist -")
			} else {
				fmt.Println("- [OK] création app/.env.dist -")
			}
		}

		// ceration du readme
		readme := filepath.Join("README.md")
		{
			if err := tools.WriteFileAlways(readme, stage1.ReadmeContent(nameApp)); err != nil {
				fmt.Println("- [KO] création du README.md -")
			} else {
				fmt.Println("- [OK] création du README.md -")
			}
		}

		// creation du makefile
		makefile := filepath.Join("Makefile")
		{
			if err := tools.WriteFileIfAbsent(makefile, stage1.MakefileContent()); err != nil {
				fmt.Println("- [KO] création du Makefile -")
			} else {
				fmt.Println("- [OK] création du Makefile -")
			}
		}

		// creation du makefile
		entrypoint := filepath.Join(pathFolderApp, "entrypoint.sh")
		{
			if err := tools.WriteFileIfAbsent(entrypoint, stage1.EntrypointShContent()); err != nil {
				fmt.Println("- [KO] création app/entrypoint.sh -")
			} else {
				fmt.Println("- [OK] création app/entrypoint.sh -")
			}
		}

		// creation releaserc
		releaserc := filepath.Join(".releaserc.json")
		{
			if err := tools.WriteFileIfAbsent(releaserc, stage1.ReleasercContent()); err != nil {
				fmt.Println("- [KO] création du .releaserc.json -")
			} else {
				fmt.Println("- [OK] création du .releaserc.json -")
			}
		}

		// creation CI/CD
		ciDir := filepath.Join(".github", "workflows")
		{
			if err := tools.EnsureDir(ciDir); err != nil {
				return err
			} else {
				fmt.Println("- [OK] création du dossier ci .github/workflows -")
			}
		}

		preprodWorkflowPath := filepath.Join(ciDir, "preprod.yml")
		{
			if err := tools.WriteFileIfAbsent(preprodWorkflowPath, stage1.GithubActionPreprodContent(nameFolderProject)); err != nil {
				fmt.Println("- [KO] création .github/workflows/preprod.yml -")
			} else {
				fmt.Println("- [OK] création .github/workflows/preprod.yml -")
			}
		}

		prodWorkflowPath := filepath.Join(ciDir, "prod.yml")
		{
			if err := tools.WriteFileIfAbsent(prodWorkflowPath, stage1.GithubActionProdContent(nameApp)); err != nil {
				fmt.Println("- [KO] création .github/workflows/prod.yml -")
			} else {
				fmt.Println("- [OK] création .github/workflows/prod.yml -")
			}
		}

        // modification angular.json
        {
            if err := stage1.PatchAngularJSON(stage1.PatchOptions{
                // angular.json est dans le dossier de l'app
                AngularJSONPath: filepath.Join(pathFolderApp, "angular.json"),
                ProjectOldName:  "app",
                ProjectNewName:  "app",
                OutputPath:      "dist/app",
                BudgetStyleWarn: "500kB",
                BudgetStyleErr:  "1MB",
				Serve: &stage1.ServeOptions{
					Host:         "0.0.0.0",
					Port:         3000,
					Poll:         2000,
					AllowedHosts: allowedHost,
				},
				DisableAnalytics: true,
			}); err != nil {
				return fmt.Errorf("échec de la modification angular.json: %v", err)
			}
			fmt.Println("- [OK] Modifcation app/angular.json -")
		}

        // modification package.json
        {
            // package.json est dans le dossier de l'app
            if err := stage1.ReplacePackageJSONScripts(filepath.Join(pathFolderApp, "package.json"), map[string]string{
                "ng":            "ng",
                "start":         "ng serve",
                "build":         "ng build --configuration production",
                "build:ssr":     "ng build --configuration production",
                "watch":         "ng build --watch --configuration development",
				"test":          "ng test --browsers=ChromeHeadlessNoSandbox --watch --poll=2000",
				"test:ci":       "ng test --watch=false --browsers=ChromeHeadlessNoSandbox",
				"serve:ssr:app": "node dist/app/server/server.mjs",
			}); err != nil {
				return fmt.Errorf("échec de la modification package.json: %v", err)
			}
			fmt.Println("- [OK] Modifcation app/package.json -")
		}

		// installation tailwindcss
		// npm i -D @tailwindcss/postcss dans le dossier de l'app
		{
			cmd := exec.Command("bash", "-lc", fmt.Sprintf("cd %s && npm install -D @tailwindcss/postcss", pathFolderApp))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("échec installation tailwindcss/postcss: %w", err)
			}
			fmt.Println("- [OK] installation @tailwindcss/postcss -")
		}

		// créer postcss.config.js à la racine de l'app
		postcssConfigPath := filepath.Join(pathFolderApp, "postcss.config.js")
		{
			postcssConfig := `module.exports = {\n  plugins: {\n    '@tailwindcss/postcss': {},\n  },\n};\n`
			if err := os.WriteFile(postcssConfigPath, []byte(postcssConfig), 0o644); err != nil {
				return fmt.Errorf("échec écriture %s: %w", postcssConfigPath, err)
			}
			fmt.Println("- [OK] création app/postcss.config.js -")
		}

		// s'assurer que src/styles.css importe Tailwind
		stylesPath := filepath.Join(pathFolderApp, "src", "styles.css")
		{
			stylesPath = filepath.Join(pathFolderApp, "src", "styles.css")
			content := "@import \"tailwindcss\";\n"
			if err := os.WriteFile(stylesPath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("échec écriture %s: %w", stylesPath, err)
			}
			fmt.Println("- [OK] fichier app/src/styles.css écrasé avec import Tailwind -")
		}

		fmt.Println("- Fichiers générés:")
		fmt.Printf("  %s\n", dockerfilePath)
		fmt.Printf("  %s\n", composeBase)
		fmt.Printf("  %s\n", composePreprod)
		fmt.Printf("  %s\n", composeProd)
		fmt.Printf("  %s\n", env)
		fmt.Printf("  %s\n", envDist)
		fmt.Printf("  %s\n", appEnv)
		fmt.Printf("  %s\n", appEnvDist)
		fmt.Printf("  %s\n", readme)
		fmt.Printf("  %s\n", makefile)
		fmt.Printf("  %s\n", entrypoint)
		fmt.Printf("  %s\n", preprodWorkflowPath)
		fmt.Printf("  %s\n", prodWorkflowPath)
		fmt.Printf("  %s\n", releaserc)
		fmt.Printf("  %s\n", postcssConfigPath)
		fmt.Printf("  %s\n", stylesPath)

		fmt.Println()

		docker.PrintDockerHints(nameApp)

		fmt.Println("- Projet Angular SSR créé avec succès -")
		fmt.Println("- utiliser les commandes make pour commencer à dev ... -")

		return nil
	},
}
