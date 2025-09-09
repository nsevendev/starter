package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/nsevendev/starter/internal/starter1"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
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
		project := sanitizeName(starter1Name)
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
			if askYesNo(fmt.Sprintf("  Appliquer le host par défaut => %v ? [o/N]: ", def), true) {
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
			if askYesNo(fmt.Sprintf("  Appliquer la version de node par défaut => %v ? [o/N]: ", starter1NodeVer), true) {
				fmt.Printf("- Version %v appliquée -\n", starter1NodeVer)
			} else {
				return errors.New("commande annulée: aucune version de node défini")
			}
		}

		// pas de port renseigné, propose port par defaut
		if !cmd.Flags().Changed("port") {
			fmt.Println("- Aucun port renseigné -")
			if askYesNo(fmt.Sprintf("  Appliquer le port par défaut => %v ? [o/N]: ", starter1Port), true) {
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
		if askYesNo(fmt.Sprintf("  Est ce que ses valeurs vous conviennent ? [o/N]: "), true) {
			fmt.Printf("------ Initialisation du projet ------\n")
		} else {
			return errors.New("commande annulée: les valeurs définis ne conviennent pas")
		}

		// creation du projet angular
		fmt.Printf("- Lancement Angular CLI dans %s: ng new %s --ssr \n", project, project)
		if err := runAngularCreate(project, wd); err != nil {
			return fmt.Errorf("échec de la création Angular: %w", err)
		}

		fmt.Println("- [OK] Application Angular -")

		// creation du dossier docker
		dockerDir := filepath.Join("docker")
		if err := ensureDir(dockerDir); err != nil {
			return err
		}

		fmt.Println("- [OK] Dossier docker -")

		// creation du dockerfile app
		dockerfilePath := filepath.Join(dockerDir, "Dockerfile")
		if err := writeFileIfAbsent(dockerfilePath, starter1.DockerfileContent(starter1NodeVer)); err != nil {
			return err
		}

		fmt.Println("- [OK] Dockerfile du projet angular -")

		// creation des fichiers compose
		baseCompose := filepath.Join(dockerDir, "compose.yaml")
		preprodCompose := filepath.Join(dockerDir, "compose.preprod.yaml")
		prodCompose := filepath.Join(dockerDir, "compose.prod.yaml")
		if err := writeFileIfAbsent(baseCompose, starter1.ComposeContent(project, cwdName)); err != nil {
			return err
		}
		fmt.Println("- [OK] compose.yaml du projet -")

		if err := writeFileIfAbsent(preprodCompose, starter1.ComposePreprodContent(project, cwdName)); err != nil {
			return err
		}
		fmt.Println("- [OK] compose.preprod.yaml du projet angular -")

		if err := writeFileIfAbsent(prodCompose, starter1.ComposeProdContent(project, cwdName)); err != nil {
			return err
		}
		fmt.Println("- [OK] compose.prod.yaml du projet angular -")

		// creation des env root
		env := filepath.Join(".env")
		envDist := filepath.Join(".env.dist")
		if err := writeFileIfAbsent(env, starter1.EnvRootContent(starter1Port, starter1NodeVer, starter1HostRule)); err != nil {
			return err
		}
		if err := writeFileIfAbsent(envDist, starter1.EnvRootContent(starter1Port, starter1NodeVer, starter1HostRule)); err != nil {
			return err
		}

		fmt.Println("- [OK] .env root -")

		// creation des env app
		appEnv := filepath.Join(appDir, ".env")
		appEnvDist := filepath.Join(appDir, ".env.dist")
		if err := writeFileIfAbsent(appEnv, starter1.EnvAppContent()); err != nil {
			return err
		}
		if err := writeFileIfAbsent(appEnvDist, starter1.EnvAppContent()); err != nil {
			return err
		}

		fmt.Println("- [OK] app/.env -")

		// ceration du readme
		readme := filepath.Join("README.md")
		if err := writeFileAlways(readme, starter1.ReadmeContent(project)); err != nil {
			return err
		}

		fmt.Println("- [OK] README.md -")

		// creation du makefile
		makefile := filepath.Join("Makefile")
		if err := writeFileIfAbsent(makefile, starter1.MakefileContent()); err != nil {
			return err
		}

		fmt.Println("- [OK] Makefile -")

		// creation du makefile
		entrypoint := filepath.Join(appDir, "entrypoint.sh")
		if err := writeFileIfAbsent(entrypoint, starter1.EntrypointShContent()); err != nil {
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

		printDockerHints(project)

		fmt.Println("Next steps:")

		composeCmd := chooseComposeCmd()
		fmt.Printf("- cd %s/docker && APP_ENV=dev %s -f compose.yaml up --build\n", project, composeCmd)
		return nil
	},
}

// sanitizeName converts the provided name into a docker/host friendly slug
func sanitizeName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9-_]+`)
	s = re.ReplaceAllString(s, "")
	s = strings.Trim(s, "-")
	return s
}

func ensureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("création du dossier %s: %w", path, err)
	}
	return nil
}

func writeFileAlways(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("écriture du fichier %s: %w", path, err)
	}
	fmt.Printf("  (écrasé) %s\n", path)
	return nil
}

func writeFileIfAbsent(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  (skip) %s existe déjà\n", path)
		return nil
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("écriture du fichier %s: %w", path, err)
	}
	return nil
}

func askYesNo(prompt string, defaultNo bool) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return !defaultNo
	}
	switch input {
	case "o", "oui", "y", "yes":
		return true
	case "n", "non", "no":
		return false
	default:
		return false
	}
}

func contentReadme(project string) string {
	return fmt.Sprintf(`# %s (Angular SSR Starter)

Ce dossier contient le code de l'application Angular.

Étapes recommandées:

1. Initialiser l'application Angular:
   ng new %s --ssr --directory .
   # ou (sans Angular CLI global)
   npx @angular/cli@latest new %s --ssr --directory .

2. (Optionnel) Ajuster la config SSR si besoin

3. Lancer en dev via Docker Compose:
   cd ../docker && APP_ENV=dev docker compose -f compose.yaml up --build

Notes:
- Les variables sont définies dans .env/.env.dist
- Le Dockerfile inclut les stages: base, dev, build, runtime (+ alias prod/preprod)
- Les fichiers compose: compose.yaml, compose.preprod.yaml, compose.prod.yaml
`, project, project, project)
}

func runAngularCreate(projectName, workdir string) error {
	// Pré-vérification outils
	hasNg := hasCommand("ng")
	hasNpx := hasCommand("npx")

	if !hasNg {
		fmt.Println("[INFO] Angular CLI 'ng' introuvable.")
		if hasNpx {
			fmt.Println("[INFO] Utilisation de 'npx @angular/cli@latest new <name> --ssr --directory .' en fallback.")
		} else {
			return errors.New("ni 'ng' ni 'npx' disponibles. Installez Node.js et Angular CLI: npm install -g @angular/cli")
		}
	}

	// Utilise Angular CLI directement si présent
	cmd := exec.Command("ng", "new", projectName, "--ssr", "--skip-git")
	cmd.Dir = workdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		// Fallback to npx Angular CLI si ng non présent
		if !hasNpx {
			return fmt.Errorf("échec 'ng new'. Et 'npx' n'est pas disponible. Installez Angular CLI: npm install -g @angular/cli (ou installez npx)")
		}
		fallback := exec.Command("npx", "-y", "@angular/cli@latest", "new", projectName, "--ssr", "--skip-git")
		fallback.Dir = workdir
		fallback.Stdout = os.Stdout
		fallback.Stderr = os.Stderr
		fallback.Stdin = os.Stdin
		if err2 := fallback.Run(); err2 != nil {
			return fmt.Errorf("'ng' et 'npx' ont échoué: %v / %v. Guide: installer Node.js >= 18 et Angular CLI: npm install -g @angular/cli", err, err2)
		}
	}
	return nil
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func hasDocker() bool {
	if !hasCommand("docker") {
		fmt.Println("[warn] Docker introuvable. Installez Docker Desktop / Docker Engine.")
		return false
	}
	fmt.Println("[ok] Docker est présent.")
	return true
}

func hasDockerCompose() (bool, bool) {
	// pas de docker
	if !hasDocker() {
		return false, false
	}
	// essaie docker compose
	cmd := exec.Command("docker", "compose", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		for _, line := range lines {
			fmt.Printf("[ok] %s\n", line)
		}
		return true, false
	}
	// essaie ancien commande docker compose
	if hasCommand("docker-compose") {
		cmd2 := exec.Command("docker-compose", "--version")
		var out2 bytes.Buffer
		cmd2.Stdout = &out2
		cmd2.Stderr = &out2
		err2 := cmd2.Run()
		if err2 == nil {
			lines := strings.Split(strings.TrimSpace(out2.String()), "\n")
			for _, line := range lines {
				fmt.Printf("[ok] %s\n", line)
			}
			fmt.Println("[warn] Vous avez une ancienne version de docker-compose")
			return false, true
		}
		fmt.Println("[warn] Vous avez une ancienne version de docker-compose")
	}
	fmt.Println("[warn] 'docker compose' ou 'docker-compose' introuvable.")
	return false, false
}

func chooseComposeCmd() string {
	hasSub, hasBin := hasDockerCompose()
	switch {
	case hasSub:
		return "docker compose"
	case hasBin:
		return "docker-compose"
	default:
		return "docker compose" // default hint
	}
}

func dockerNetworkExists(name string) bool {
	if !hasDocker() {
		return false
	}
	cmd := exec.Command("docker", "network", "inspect", name)
	return cmd.Run() == nil
}

// printDockerHints check docker et le reseau externe
func printDockerHints(project string) {
	network := fmt.Sprintf("%s-nseven", project)
	hasSub, hasBin := hasDockerCompose()

	if !hasDocker() {
		fmt.Println("[warn] Docker introuvable. Installez Docker Desktop / Docker Engine.")
	}
	if !hasSub && !hasBin {
		fmt.Println("[warn] 'docker compose' ou 'docker-compose' introuvable. Installez le plugin Compose ou utilisez Docker Desktop récent.")
	}
	if !dockerNetworkExists(network) {
		fmt.Printf("[INFO] Réseau externe '%s' absent. Créez-le avant de lancer: docker network create %s\n", network, network)
		if askYesNo(fmt.Sprintf("  Voulez vous creer le reseau %v => %v ? [o/N]: ", network), true) {
			cmd := exec.Command("docker", "network", "create", network)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("[ERROR] Échec de la création du réseau '%s': %v\n", network, err)
			} else {
				fmt.Printf("[OK] Réseau '%s' créé avec succès.\n", network)
			}
		} else {
			fmt.Printf("[INFO] Réseau '%s' non créé. Pensez à l'initialiser plus tard:\n  docker network create %s\n", network, network)
		}
	}
}
