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
	initProjectName  string
	initVersion      string
	initAllowedHosts []string
	initHostTraefik  string
	initDbName       string
	initDeployFolder string
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
	initTempAngssrGo.Flags().StringSliceVar(&initAllowedHosts, "allowedHosts", []string{}, "hosts autorisés pour Angular et CORS (ex: --allowedHosts=test.local,api.local)")
	initTempAngssrGo.Flags().StringVar(&initHostTraefik, "hostTraefik", "", "host pour Traefik (ex: myproject.local)")
	initTempAngssrGo.Flags().StringVar(&initDbName, "dbName", "", "nom de la base de données MongoDB")
	initTempAngssrGo.Flags().StringVar(&initDeployFolder, "deployFolder", "", "nom du dossier de déploiement sur le serveur")

	rootCmd.AddCommand(initTempAngssrGo)
}

// applyTemplateModifications applique toutes les modifications du template selon les flags fournis
func applyTemplateModifications(projectPath string) error {
	// 1. Modification de app/angular.json (ligne 72 - allowedHosts)
	if len(initAllowedHosts) > 0 {
		if err := modifyAngularJson(projectPath); err != nil {
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
	if initDeployFolder != "" {
		if err := modifyPreprodWorkflow(projectPath); err != nil {
			return err
		}
	}

	// 4. Modification de .github/workflows/prod.yml
	if initDeployFolder != "" {
		if err := modifyProdWorkflow(projectPath); err != nil {
			return err
		}
	}

	// 5. Modification de docker/mongo-init/init-volume-db.js
	if initDbName != "" {
		if err := modifyMongoInit(projectPath); err != nil {
			return err
		}
	}

	// 6. Modification de docker/compose.yaml
	if initDbName != "" {
		if err := modifyComposeYaml(projectPath); err != nil {
			return err
		}
	}

	// 7. Modification de docker/compose.preprod.yaml
	if initDbName != "" {
		if err := modifyComposePreprod(projectPath); err != nil {
			return err
		}
	}

	// 8. Modification de api/.env.dist
	if initDbName != "" || initHostTraefik != "" || len(initAllowedHosts) > 0 {
		if err := modifyApiEnvDist(projectPath); err != nil {
			return err
		}
	}

	// 9. Modification du Makefile (nom du container)
	if initProjectName != "" {
		if err := modifyMakefile(projectPath); err != nil {
			return err
		}
	}

	return nil
}

// modifyAngularJson modifie app/angular.json ligne 72 pour allowedHosts
func modifyAngularJson(projectPath string) error {
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

		// Mise à jour de allowedHosts
		options["allowedHosts"] = initAllowedHosts
		fmt.Printf("    ✓ allowedHosts configuré pour le projet '%s': %v\n", projectKey, initAllowedHosts)
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

// modifyRootEnvDist modifie .env.dist à la racine
func modifyRootEnvDist(projectPath string) error {
	fmt.Println("  Modification de .env.dist...")
	filePath := filepath.Join(projectPath, ".env.dist")

	// Ligne 5: TRAEFIK_HOST
	if err := tools.ReplaceInFile(filePath, "myhost", initHostTraefik); err != nil {
		return err
	}

	// Ligne 32: NAME_NETWORK (utilise le nom du projet)
	if err := tools.ReplaceInFile(filePath, "NAME_APP", initProjectName); err != nil {
		return err
	}

	return nil
}

// modifyPreprodWorkflow modifie .github/workflows/preprod.yml
func modifyPreprodWorkflow(projectPath string) error {
	fmt.Println("  Modification de .github/workflows/preprod.yml...")
	filePath := filepath.Join(projectPath, ".github", "workflows", "preprod.yml")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("lecture de preprod.yml: %w", err)
	}

	contentStr := string(content)

	// Décommenter les lignes commentées (enlever les # en début de ligne)
	contentStr = strings.ReplaceAll(contentStr, "#     - name:", "    - name:")
	contentStr = strings.ReplaceAll(contentStr, "#       uses:", "      uses:")
	contentStr = strings.ReplaceAll(contentStr, "#       with:", "      with:")

	// Ligne 155: changer "myfolder" par le deployFolder
	contentStr = strings.ReplaceAll(contentStr, "myfolder", initDeployFolder)

	if err := os.WriteFile(filePath, []byte(contentStr), 0o644); err != nil {
		return fmt.Errorf("écriture de preprod.yml: %w", err)
	}

	fmt.Printf("    ✓ preprod.yml configuré avec le dossier: %s\n", initDeployFolder)
	return nil
}

// modifyProdWorkflow modifie .github/workflows/prod.yml
func modifyProdWorkflow(projectPath string) error {
	fmt.Println("  Modification de .github/workflows/prod.yml...")
	filePath := filepath.Join(projectPath, ".github", "workflows", "prod.yml")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("lecture de prod.yml: %w", err)
	}

	contentStr := string(content)

	// Décommenter les lignes commentées
	contentStr = strings.ReplaceAll(contentStr, "#     - name:", "    - name:")
	contentStr = strings.ReplaceAll(contentStr, "#       uses:", "      uses:")
	contentStr = strings.ReplaceAll(contentStr, "#       with:", "      with:")

	// Remplacer myfolder par deployFolder
	contentStr = strings.ReplaceAll(contentStr, "myfolder", initDeployFolder)

	if err := os.WriteFile(filePath, []byte(contentStr), 0o644); err != nil {
		return fmt.Errorf("écriture de prod.yml: %w", err)
	}

	fmt.Printf("    ✓ prod.yml configuré avec le dossier: %s\n", initDeployFolder)
	return nil
}

// modifyMongoInit modifie docker/mongo-init/init-volume-db.js
func modifyMongoInit(projectPath string) error {
	fmt.Println("  Modification de docker/mongo-init/init-volume-db.js...")
	filePath := filepath.Join(projectPath, "docker", "mongo-init", "init-volume-db.js")

	// Remplacer le nom de BDD par défaut (supposons "mydb" ou "testdb")
	if err := tools.ReplaceInFile(filePath, "testdb", initDbName); err != nil {
		// Si testdb n'existe pas, essayer mydb
		if err := tools.ReplaceInFile(filePath, "mydb", initDbName); err != nil {
			return err
		}
	}

	return nil
}

// modifyComposeYaml modifie docker/compose.yaml
func modifyComposeYaml(projectPath string) error {
	fmt.Println("  Modification de docker/compose.yaml...")
	filePath := filepath.Join(projectPath, "docker", "compose.yaml")

	// Lignes 85, 89, 90 concernent probablement le nom de la BDD
	if err := tools.ReplaceInFile(filePath, "testdb", initDbName); err != nil {
		if err := tools.ReplaceInFile(filePath, "mydb", initDbName); err != nil {
			return err
		}
	}

	return nil
}

// modifyComposePreprod modifie docker/compose.preprod.yaml
func modifyComposePreprod(projectPath string) error {
	fmt.Println("  Modification de docker/compose.preprod.yaml...")
	filePath := filepath.Join(projectPath, "docker", "compose.preprod.yaml")

	// Lignes 66, 70, 71 concernent le nom de la BDD
	if err := tools.ReplaceInFile(filePath, "testdb", initDbName); err != nil {
		if err := tools.ReplaceInFile(filePath, "mydb", initDbName); err != nil {
			return err
		}
	}

	return nil
}

// modifyApiEnvDist modifie api/.env.dist
func modifyApiEnvDist(projectPath string) error {
	fmt.Println("  Modification de api/.env.dist...")
	filePath := filepath.Join(projectPath, "api", ".env.dist")

	// Ligne 4: nom de la BDD
	if initDbName != "" {
		if err := tools.ReplaceInFile(filePath, "testdb", initDbName); err != nil {
			if err := tools.ReplaceInFile(filePath, "mydb", initDbName); err != nil {
				return err
			}
		}
	}

	// Ligne 8: host (doit correspondre à .env de la racine)
	if initHostTraefik != "" {
		if err := tools.ReplaceInFile(filePath, "myhost", initHostTraefik); err != nil {
			return err
		}
	}

	// Lignes 34, 35, 36: hosts pour CORS
	if len(initAllowedHosts) > 0 {
		// Créer une liste de hosts séparés par des virgules
		hostsString := strings.Join(initAllowedHosts, ",")
		if err := tools.ReplaceInFile(filePath, "test.local", hostsString); err != nil {
			return err
		}
	}

	return nil
}

// modifyMakefile modifie le Makefile pour le nom du container
func modifyMakefile(projectPath string) error {
	fmt.Println("  Modification du Makefile...")
	filePath := filepath.Join(projectPath, "Makefile")

	// Remplacer le nom du container générique par celui du projet
	// Le format est généralement: project-api-1 ou project_api_1
	containerName := initProjectName + "-api-1"

	if err := tools.ReplaceInFile(filePath, "myproject-api-1", containerName); err != nil {
		// Essayer avec underscore
		containerNameUnderscore := initProjectName + "_api_1"
		if err := tools.ReplaceInFile(filePath, "myproject_api_1", containerNameUnderscore); err != nil {
			return err
		}
	}

	return nil
}
