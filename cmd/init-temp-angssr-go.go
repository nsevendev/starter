package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/nsevendev/starter/internal/tools"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	initProjectName string
	initVersion     string
	initHostTraefik string
)

var initTempAngssrGo = &cobra.Command{
	Use:   "init-temp-angssr-go",
	Short: "initialise un projet angular ssr avec go, mongo",
	Long:  `Initialise un projet angular ssr avec go, mongo, redis, docker, r2, mailer, etc...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// validation des flags
		if initProjectName == "" {
			return fmt.Errorf("le flag --name est requis")
		}
		if initVersion == "" {
			return fmt.Errorf("le flag --version est requis (ex: --version=v1.0.0)")
		}

		// chemin du projet (dossier courant + nom du projet)
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("erreur lors de la récupération du dossier courant: %w", err)
		}
		projectPath := filepath.Join(currentDir, initProjectName)

		// check si le dossier existe deja
		if _, err := os.Stat(projectPath); err == nil {
			return fmt.Errorf("le dossier %s existe déjà", initProjectName)
		}

		// clone du repo avec la version specifiée
		fmt.Printf("Clonage du template (version %s)...\n", initVersion)
		cloneCmd := exec.Command("git", "clone", "--branch", initVersion, "--depth", "1",
			"https://github.com/nsevendev/temp-angssr-go.git", initProjectName)
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		if err := cloneCmd.Run(); err != nil {
			return fmt.Errorf("erreur lors du clonage: %w", err)
		}

		// delete old .git
		fmt.Println("Suppression du .git...")
		gitPath := filepath.Join(projectPath, ".git")
		if err := os.RemoveAll(gitPath); err != nil {
			return fmt.Errorf("erreur lors de la suppression du .git: %w", err)
		}

		// Application des modifications automatiques
		fmt.Println("\nConfiguration du projet...")
		if err := applyTemplateModifications(projectPath); err != nil {
			return fmt.Errorf("erreur lors de la configuration: %w", err)
		}

		fmt.Printf("\n✓ Projet %s créé et configuré avec succès !\n", initProjectName)
		return nil
	},
}

func init() {
	initTempAngssrGo.Flags().StringVar(&initProjectName, "name", "", "nom du projet (requis)")
	initTempAngssrGo.Flags().StringVar(&initVersion, "version", "", "version du template (ex: v1.0.0) (requis)")
	initTempAngssrGo.Flags().StringVar(&initHostTraefik, "hostTraefik", "", "host pour Traefik (ex: myproject.local)")

	rootCmd.AddCommand(initTempAngssrGo)
}

// applyTemplateModifications applique toutes les modifications du template selon les flags fournis
func applyTemplateModifications(projectPath string) error {
	// 1. Modification de app/angular.json (ligne 72 - allowedHosts)
	// allowedHosts = hostTraefik (si fourni)
	if initHostTraefik != "" {
		if err := modifyAngularJson(projectPath, initHostTraefik); err != nil {
			return err
		}
	}

	// 2. Modification de .env.dist (ligne 5 host traefik, ligne 32 nom réseau)
	if initHostTraefik != "" {
		if err := modifyRootEnvDist(projectPath); err != nil {
			return err
		}
	}

	// 3. Modification de .github/workflows/preprod.yml
	// deployFolder = projectName
	if err := modifyPreprodWorkflow(projectPath, initProjectName); err != nil {
		return err
	}

	// 4. Modification de .github/workflows/prod.yml
	// deployFolder = projectName
	if err := modifyProdWorkflow(projectPath, initProjectName); err != nil {
		return err
	}

	// 5. Modification de docker/mongo-init/init-volume-db.js
	// dbName = projectName
	if err := modifyMongoInit(projectPath, initProjectName); err != nil {
		return err
	}

	// 6. Modification de docker/compose.yaml
	// dbName = projectName
	if err := modifyComposeYaml(projectPath, initProjectName); err != nil {
		return err
	}

	// 7. Modification de docker/compose.preprod.yaml
	// dbName = projectName
	if err := modifyComposePreprod(projectPath, initProjectName); err != nil {
		return err
	}

	// 8. Modification de api/.env.dist
	// dbName = projectName, allowedHosts = hostTraefik
	if err := modifyApiEnvDist(projectPath, initProjectName, initHostTraefik); err != nil {
		return err
	}

	// 9. Modification du Makefile (nom du container)
	if err := modifyMakefile(projectPath); err != nil {
		return err
	}

	// 10. Copie de app/.env.dist vers app/.env
	if err := copyAppEnv(projectPath); err != nil {
		return err
	}

	return nil
}

// modifyAngularJson modifie app/angular.json ligne 72 pour allowedHosts
func modifyAngularJson(projectPath, allowedHost string) error {
	fmt.Println("  Modification de app/angular.json...")
	filePath := filepath.Join(projectPath, "app", "angular.json")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("lecture de angular.json: %w", err)
	}

	var angularConfig map[string]interface{}
	if err := json.Unmarshal(content, &angularConfig); err != nil {
		return fmt.Errorf("parsing de angular.json: %w", err)
	}

	// Navigation dans la structure JSON pour trouver allowedHosts
	projects, ok := angularConfig["projects"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("structure projects non trouvée dans angular.json")
	}

	// Trouver le premier projet (généralement le nom du projet)
	for projectKey, projectVal := range projects {
		projectData, ok := projectVal.(map[string]interface{})
		if !ok {
			continue
		}

		architect, ok := projectData["architect"].(map[string]interface{})
		if !ok {
			continue
		}

		serve, ok := architect["serve"].(map[string]interface{})
		if !ok {
			continue
		}

		options, ok := serve["options"].(map[string]interface{})
		if !ok {
			continue
		}

		// Mise à jour de allowedHosts avec un slice contenant le host
		options["allowedHosts"] = []string{allowedHost}
		fmt.Printf("    ✓ allowedHosts configuré pour le projet '%s': [%s]\n", projectKey, allowedHost)
		break
	}

	// Réécriture du fichier JSON
	newContent, err := json.MarshalIndent(angularConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("serialization de angular.json: %w", err)
	}

	if err := os.WriteFile(filePath, newContent, 0o644); err != nil {
		return fmt.Errorf("écriture de angular.json: %w", err)
	}

	return nil
}

// modifyRootEnvDist modifie .env.dist et crée .env à la racine
func modifyRootEnvDist(projectPath string) error {
	fmt.Println("  Modification de .env.dist et création de .env...")
	filePathDist := filepath.Join(projectPath, ".env.dist")
	filePathEnv := filepath.Join(projectPath, ".env")

	// Modifier .env.dist
	// Ligne 5: TRAEFIK_HOST=myhost -> TRAEFIK_HOST=<hostTraefik>
	if err := tools.ReplaceInFile(filePathDist, "TRAEFIK_HOST=myhost", "TRAEFIK_HOST="+initHostTraefik); err != nil {
		return err
	}

	// Ligne 6: HOST_TRAEFIK_APP=Host(`test.local`) -> HOST_TRAEFIK_APP=Host(`<hostTraefik>.local`)
	if err := tools.ReplaceInFile(filePathDist, "HOST_TRAEFIK_APP=Host(`test.local`)", "HOST_TRAEFIK_APP=Host(`"+initHostTraefik+".local`)"); err != nil {
		return err
	}

	// Ligne 7: HOST_TRAEFIK_API=Host(`test-api.local`) -> HOST_TRAEFIK_API=Host(`<hostTraefik>-api.local`)
	if err := tools.ReplaceInFile(filePathDist, "HOST_TRAEFIK_API=Host(`test-api.local`)", "HOST_TRAEFIK_API=Host(`"+initHostTraefik+"-api.local`)"); err != nil {
		return err
	}

	// Ligne 32: NAME_APP=monapp -> NAME_APP=<projectName>
	if err := tools.ReplaceInFile(filePathDist, "NAME_APP=monapp", "NAME_APP="+initProjectName); err != nil {
		return err
	}

	// Copier .env.dist vers .env
	content, err := os.ReadFile(filePathDist)
	if err != nil {
		return fmt.Errorf("lecture de .env.dist: %w", err)
	}

	if err := os.WriteFile(filePathEnv, content, 0o644); err != nil {
		return fmt.Errorf("création de .env: %w", err)
	}

	fmt.Println("    ✓ .env.dist et .env configurés")
	return nil
}

// modifyPreprodWorkflow modifie .github/workflows/preprod.yml
func modifyPreprodWorkflow(projectPath, deployFolder string) error {
	fmt.Println("  Modification de .github/workflows/preprod.yml...")
	filePath := filepath.Join(projectPath, ".github", "workflows", "preprod.yml")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("lecture de preprod.yml: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Décommenter les lignes 1 à 8 (index 0 à 7)
	for i := 0; i < 8 && i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "#") {
			lines[i] = strings.TrimPrefix(lines[i], "#")
		}
	}

	contentStr := strings.Join(lines, "\n")

	// Décommenter les autres lignes commentées
	contentStr = strings.ReplaceAll(contentStr, "#     - name:", "    - name:")
	contentStr = strings.ReplaceAll(contentStr, "#       uses:", "      uses:")
	contentStr = strings.ReplaceAll(contentStr, "#       with:", "      with:")

	// Ligne 155: changer "myfolder" par le deployFolder
	contentStr = strings.ReplaceAll(contentStr, "myfolder", deployFolder)

	if err := os.WriteFile(filePath, []byte(contentStr), 0o644); err != nil {
		return fmt.Errorf("écriture de preprod.yml: %w", err)
	}

	fmt.Printf("    ✓ preprod.yml configuré avec le dossier: %s\n", deployFolder)
	return nil
}

// modifyProdWorkflow modifie .github/workflows/prod.yml
func modifyProdWorkflow(projectPath, deployFolder string) error {
	fmt.Println("  Modification de .github/workflows/prod.yml...")
	filePath := filepath.Join(projectPath, ".github", "workflows", "prod.yml")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("lecture de prod.yml: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Décommenter les lignes 1 à 7 (index 0 à 6)
	for i := 0; i < 7 && i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "#") {
			lines[i] = strings.TrimPrefix(lines[i], "#")
		}
	}

	contentStr := strings.Join(lines, "\n")

	// Décommenter les autres lignes commentées
	contentStr = strings.ReplaceAll(contentStr, "#     - name:", "    - name:")
	contentStr = strings.ReplaceAll(contentStr, "#       uses:", "      uses:")
	contentStr = strings.ReplaceAll(contentStr, "#       with:", "      with:")

	// Ligne 174: changer "myfolder" par le deployFolder
	contentStr = strings.ReplaceAll(contentStr, "myfolder", deployFolder)

	if err := os.WriteFile(filePath, []byte(contentStr), 0o644); err != nil {
		return fmt.Errorf("écriture de prod.yml: %w", err)
	}

	fmt.Printf("    ✓ prod.yml configuré avec le dossier: %s\n", deployFolder)
	return nil
}

// modifyMongoInit modifie docker/mongo-init/init-volume-db.js
func modifyMongoInit(projectPath, projectName string) error {
	fmt.Println("  Modification de docker/mongo-init/init-volume-db.js...")
	filePath := filepath.Join(projectPath, "docker", "mongo-init", "init-volume-db.js")

	// Lignes 17-20: remplacer "myapp" par le nom du projet
	// Ligne 17: myapp_prod
	if err := tools.ReplaceInFile(filePath, "myapp_prod", projectName+"_prod"); err != nil {
		return err
	}

	// Ligne 18: myapp_preprod
	if err := tools.ReplaceInFile(filePath, "myapp_preprod", projectName+"_preprod"); err != nil {
		return err
	}

	// Ligne 19: myapp_dev
	if err := tools.ReplaceInFile(filePath, "myapp_dev", projectName+"_dev"); err != nil {
		return err
	}

	// Ligne 20: myapp_test
	if err := tools.ReplaceInFile(filePath, "myapp_test", projectName+"_test"); err != nil {
		return err
	}

	return nil
}

// modifyComposeYaml modifie docker/compose.yaml
func modifyComposeYaml(projectPath, projectName string) error {
	fmt.Println("  Modification de docker/compose.yaml...")
	filePath := filepath.Join(projectPath, "docker", "compose.yaml")

	// Ligne 85: temp-angssr-go
	if err := tools.ReplaceInFile(filePath, "temp-angssr-go", projectName); err != nil {
		return err
	}

	// Ligne 89: temp-angssr-go_dev_db (garder _dev_db)
	if err := tools.ReplaceInFile(filePath, "temp-angssr-go_dev_db", projectName+"_dev_db"); err != nil {
		return err
	}

	// Ligne 90: temp-angssr-go_dev_redis_data (garder _dev_redis_data)
	if err := tools.ReplaceInFile(filePath, "temp-angssr-go_dev_redis_data", projectName+"_dev_redis_data"); err != nil {
		return err
	}

	return nil
}

// modifyComposePreprod modifie docker/compose.preprod.yaml
func modifyComposePreprod(projectPath, projectName string) error {
	fmt.Println("  Modification de docker/compose.preprod.yaml...")
	filePath := filepath.Join(projectPath, "docker", "compose.preprod.yaml")

	// Ligne 66: temp-angssr-go (même que ligne 85 de compose.yaml)
	if err := tools.ReplaceInFile(filePath, "temp-angssr-go", projectName); err != nil {
		return err
	}

	// Ligne 70: temp-angssr-go_dev_db (même que ligne 89 de compose.yaml)
	if err := tools.ReplaceInFile(filePath, "temp-angssr-go_dev_db", projectName+"_dev_db"); err != nil {
		return err
	}

	// Ligne 71: temp-angssr-go_dev_redis_data (même que ligne 90 de compose.yaml)
	if err := tools.ReplaceInFile(filePath, "temp-angssr-go_dev_redis_data", projectName+"_dev_redis_data"); err != nil {
		return err
	}

	return nil
}

// modifyApiEnvDist modifie api/.env.dist et crée api/.env
func modifyApiEnvDist(projectPath, projectName, hostTraefik string) error {
	fmt.Println("  Modification de api/.env.dist et création de api/.env...")
	filePathDist := filepath.Join(projectPath, "api", ".env.dist")
	filePathEnv := filepath.Join(projectPath, "api", ".env")

	// Ligne 4: DB_NAME=myapp_dev -> DB_NAME=<projectName>_dev
	if err := tools.ReplaceInFile(filePathDist, "DB_NAME=myapp_dev", "DB_NAME="+projectName+"_dev"); err != nil {
		return err
	}

	// Ligne 8: HOST_TRAEFIK_API=Host(`test-api.local`) -> HOST_TRAEFIK_API=Host(`<hostTraefik>-api.local`)
	if hostTraefik != "" {
		if err := tools.ReplaceInFile(filePathDist, "HOST_TRAEFIK_API=Host(`test-api.local`)", "HOST_TRAEFIK_API=Host(`"+hostTraefik+"-api.local`)"); err != nil {
			return err
		}
	}

	// Lignes 34, 35, 36: http://test.local -> http://<projectName>.local
	if err := tools.ReplaceInFile(filePathDist, "http://test.local", "http://"+hostTraefik+".local"); err != nil {
		return err
	}

	// Copier .env.dist vers .env
	content, err := os.ReadFile(filePathDist)
	if err != nil {
		return fmt.Errorf("lecture de api/.env.dist: %w", err)
	}

	if err := os.WriteFile(filePathEnv, content, 0o644); err != nil {
		return fmt.Errorf("création de api/.env: %w", err)
	}

	fmt.Println("    ✓ api/.env.dist et api/.env configurés")
	return nil
}

// modifyMakefile modifie le Makefile pour le nom du container
func modifyMakefile(projectPath string) error {
	fmt.Println("  Modification du Makefile...")
	filePath := filepath.Join(projectPath, "Makefile")

	// Lignes 108, 111, 114, 117, 120, 123, 126, 129: temp-angssr-go_dev_api
	if err := tools.ReplaceInFile(filePath, "temp-angssr-go_dev_api", initProjectName+"_dev_api"); err != nil {
		return err
	}

	fmt.Printf("    ✓ Makefile configuré avec le container: %s_dev_api\n", initProjectName)
	return nil
}

// copyAppEnv copie app/.env.dist vers app/.env
func copyAppEnv(projectPath string) error {
	fmt.Println("  Création de app/.env...")
	filePathDist := filepath.Join(projectPath, "app", ".env.dist")
	filePathEnv := filepath.Join(projectPath, "app", ".env")

	// Copier .env.dist vers .env
	content, err := os.ReadFile(filePathDist)
	if err != nil {
		return fmt.Errorf("lecture de app/.env.dist: %w", err)
	}

	if err := os.WriteFile(filePathEnv, content, 0o644); err != nil {
		return fmt.Errorf("création de app/.env: %w", err)
	}

	fmt.Println("    ✓ app/.env créé")
	return nil
}
