package cmd

import (
	"fmt"
	"github.com/nsevendev/starter/internal/projets/framework"
	"github.com/nsevendev/starter/internal/projets/stage2"
	"github.com/nsevendev/starter/internal/tools"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	err               error
	hostTraefikFront  string
	hostTraefikApi    string
	pathFolderProject string
	nameFolderProject string
	pathFolderFront   string
	nodeVersion       = "22.19.0"
	goVersion         = "1.24.4"
	mongoVersion      = "7.0"
	nameServiceFront  = "front"
	nameServiceApi    = "api"
	portLinkTraefik   = 3000
)

var starter2 = &cobra.Command{
	Use:   "stage2",
	Short: "Astro ssr + api go + mongodb => node version 22.19.0, go version 1.24.4, mongo version 7.0",
	Long: `Création d'application legere, site vitrine, site dynamique, blog, e-commerce, portfolio, 
			ne convient pas pour les applications complexes.
			répondez au question pour la creation du projet Astro avec les réponses suivantes:
				- type de projet => choisissez "basic"
				- installation des dépendances => choisissez "no"
				- init git => choisissez "no"
				Ne suivez pas les instructions d'astro pour l'installation des dépendances.
			`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pathFolderProject, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("erreur récupération du path du dossier courant: %v", err)
		}
		nameFolderProject = filepath.Base(pathFolderProject)
		pathFolderFront = filepath.Join(pathFolderProject, nameServiceFront)

		if err = validateUserForStart(); err != nil {
			return err
		}

		if err = createAndSetFront(); err != nil {
			return err
		}

		if err = installDependenciesFront(); err != nil {
			return err
		}

		if err = createAndSetApi(); err != nil {
			return err
		}

		if err = createAndSetDocker(); err != nil {
			return err
		}

		if err = createCiCd(); err != nil {
			return err
		}

		fmt.Println("------ Initialisation du projet terminé ------")

		return nil
	},
}

func createAndSetDocker() error {
	fmt.Println("------ Création des fichiers Docker ------")

	pathDockerDir := filepath.Join(pathFolderProject, "docker")

	// Créer le dossier docker
	{
		if err := tools.EnsureDir(pathDockerDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier docker: %v", err)
		} else {
			fmt.Println("- [OK] création du dossier docker -")
		}
	}

	// Créer front.dockerfile
	{
		pathFrontDockerfile := filepath.Join(pathDockerDir, "front.dockerfile")
		if err := tools.WriteFileIfAbsent(pathFrontDockerfile, stage2.FrontDockerfileContent()); err != nil {
			return fmt.Errorf("- [KO] création docker/front.dockerfile: %v", err)
		} else {
			fmt.Println("- [OK] création docker/front.dockerfile -")
		}
	}

	// Créer api.dockerfile
	{
		pathApiDockerfile := filepath.Join(pathDockerDir, "api.dockerfile")
		if err := tools.WriteFileIfAbsent(pathApiDockerfile, stage2.ApiDockerfileContent()); err != nil {
			return fmt.Errorf("- [KO] création docker/api.dockerfile: %v", err)
		} else {
			fmt.Println("- [OK] création docker/api.dockerfile -")
		}
	}

	// Créer compose.yaml
	{
		pathComposeYaml := filepath.Join(pathDockerDir, "compose.yaml")
		if err := tools.WriteFileIfAbsent(pathComposeYaml, stage2.ComposeYamlContent(nameFolderProject)); err != nil {
			return fmt.Errorf("- [KO] création docker/compose.yaml: %v", err)
		} else {
			fmt.Println("- [OK] création docker/compose.yaml -")
		}
	}

	// Créer compose.preprod.yaml
	{
		pathComposePreprodYaml := filepath.Join(pathDockerDir, "compose.preprod.yaml")
		if err := tools.WriteFileIfAbsent(pathComposePreprodYaml, stage2.ComposePreprodYamlContent(nameFolderProject)); err != nil {
			return fmt.Errorf("- [KO] création docker/compose.preprod.yaml: %v", err)
		} else {
			fmt.Println("- [OK] création docker/compose.preprod.yaml -")
		}
	}

	// Créer compose.prod.yaml
	{
		pathComposeProdYaml := filepath.Join(pathDockerDir, "compose.prod.yaml")
		if err := tools.WriteFileIfAbsent(pathComposeProdYaml, stage2.ComposeProdYamlContent(nameFolderProject)); err != nil {
			return fmt.Errorf("- [KO] création docker/compose.prod.yaml: %v", err)
		} else {
			fmt.Println("- [OK] création docker/compose.prod.yaml -")
		}
	}

	// Créer mongo-init/init-volume-db.js
	{
		pathMongoInitDir := filepath.Join(pathDockerDir, "mongo-init")
		pathMongoInitJs := filepath.Join(pathMongoInitDir, "init-volume-db.js")
		if err := tools.EnsureDir(pathMongoInitDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier mongo-init: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathMongoInitJs, stage2.MongoInitContent(nameFolderProject)); err != nil {
			return fmt.Errorf("- [KO] création docker/mongo-init/init-volume-db.js: %v", err)
		} else {
			fmt.Println("- [OK] création docker/mongo-init/init-volume-db.js -")
		}
	}

	// Créer Makefile à la racine
	{
		pathMakefile := filepath.Join(pathFolderProject, "Makefile")
		if err := tools.WriteFileIfAbsent(pathMakefile, stage2.MakefileContent(nameFolderProject)); err != nil {
			return fmt.Errorf("- [KO] création Makefile: %v", err)
		} else {
			fmt.Println("- [OK] création Makefile -")
		}
	}

	// Créer .env à la racine
	{
		pathEnvRoot := filepath.Join(pathFolderProject, ".env")
		if err := tools.WriteFileIfAbsent(pathEnvRoot, stage2.EnvRootContent(hostTraefikFront, hostTraefikApi)); err != nil {
			return fmt.Errorf("- [KO] création .env à la racine: %v", err)
		} else {
			fmt.Println("- [OK] création .env à la racine -")
		}
	}

	// Créer .env.dist à la racine
	{
		pathEnvDistRoot := filepath.Join(pathFolderProject, ".env.dist")
		if err := tools.WriteFileIfAbsent(pathEnvDistRoot, stage2.EnvRootContent(hostTraefikFront, hostTraefikApi)); err != nil {
			return fmt.Errorf("- [KO] création .env.dist à la racine: %v", err)
		} else {
			fmt.Println("- [OK] création .env.dist à la racine -")
		}
	}

	// Créer README à la racine
	{
		path := filepath.Join(pathFolderProject, "README.md")
		if err := tools.WriteFileIfAbsent(path, stage2.ReadmeContent(nameFolderProject)); err != nil {
			return fmt.Errorf("- [KO] création du README à la racine: %v", err)
		} else {
			fmt.Println("- [OK] création du README à la racine -")
		}
	}

	// Créer .gitignore à la racine
	{
		pathGitignoreRoot := filepath.Join(pathFolderProject, ".gitignore")
		if err := tools.WriteFileIfAbsent(pathGitignoreRoot, stage2.GitignoreRootContent()); err != nil {
			return fmt.Errorf("- [KO] création .gitignore à la racine: %v", err)
		} else {
			fmt.Println("- [OK] création .gitignore à la racine -")
		}
	}

	return nil
}

func createCiCd() error {
	fmt.Println("------ Création des workflows CI/CD ------")

	pathGithubDir := filepath.Join(pathFolderProject, ".github", "workflows")

	// Créer le dossier .github/workflows
	{
		if err := tools.EnsureDir(pathGithubDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier .github/workflows: %v", err)
		} else {
			fmt.Println("- [OK] création du dossier .github/workflows -")
		}
	}

	// Créer preprod.yml
	{
		pathPreprodYml := filepath.Join(pathGithubDir, "preprod.yml")
		if err := tools.WriteFileIfAbsent(pathPreprodYml, stage2.PreprodWorkflowContent(nameFolderProject)); err != nil {
			return fmt.Errorf("- [KO] création .github/workflows/preprod.yml: %v", err)
		} else {
			fmt.Println("- [OK] création .github/workflows/preprod.yml -")
		}
	}

	// Créer prod.yml
	{
		pathProdYml := filepath.Join(pathGithubDir, "prod.yml")
		if err := tools.WriteFileIfAbsent(pathProdYml, stage2.ProdWorkflowContent(nameFolderProject)); err != nil {
			return fmt.Errorf("- [KO] création .github/workflows/prod.yml: %v", err)
		} else {
			fmt.Println("- [OK] création .github/workflows/prod.yml -")
		}
	}

	return nil
}

func createAndSetApi() error {
	// Vérifier Go
	goCmd := exec.Command("go", "version")
	goOutput, err := goCmd.Output()
	if err != nil {
		return fmt.Errorf("Go n'est pas installé sur cette machine")
	}
	installedGoVersion := strings.TrimSpace(strings.TrimPrefix(string(goOutput), "go version go"))
	// Extraire uniquement la version (ex: "1.24.4 darwin/arm64" -> "1.24.4")
	installedGoVersion = strings.Fields(installedGoVersion)[0]

	goOk := tools.CompareVersion(installedGoVersion, goVersion)
	if !goOk {
		return fmt.Errorf(" Go installé: %s (requis: >= %s)", installedGoVersion, goVersion)
	}

	fmt.Printf("✓ Go %s (requis: >= %s)\n", installedGoVersion, goVersion)
	fmt.Println("------ Création de l'api ------")

	pathFolderApi := filepath.Join(pathFolderProject, nameServiceApi)
	moduleName := fmt.Sprintf("%s/%s", nameFolderProject, nameServiceApi)

	// Créer le dossier api
	{
		if err := tools.EnsureDir(pathFolderApi); err != nil {
			return fmt.Errorf("- [KO] création du dossier api: %v", err)
		} else {
			fmt.Println("- [OK] création du dossier api -")
		}
	}

	// Initialiser le module Go
	{
		cmd := exec.Command("go", "mod", "init", moduleName)
		cmd.Dir = pathFolderApi
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("- [KO] initialisation du module Go: %w", err)
		} else {
			fmt.Println("- [OK] initialisation du module Go -")
		}
	}

	// Installer les dépendances
	{
		dependencies := []string{
			"github.com/gin-contrib/cors@v1.7.6",
			"github.com/gin-gonic/gin",
			"github.com/nsevenpack/env@v1.0.2",
			"github.com/nsevenpack/ginresponse@v1.2.3",
			"github.com/nsevenpack/logger/v2@v2.2.0",
			"github.com/swaggo/swag",
			"go.mongodb.org/mongo-driver",
			"github.com/swaggo/gin-swagger",
			"github.com/swaggo/files",
		}

		for _, dep := range dependencies {
			cmd := exec.Command("go", "get", dep)
			cmd.Dir = pathFolderApi
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("- [KO] installation de %s: %w", dep, err)
			}
		}
		fmt.Println("- [OK] installation des dépendances -")
	}

	// Créer .air.toml
	{
		pathAirToml := filepath.Join(pathFolderApi, ".air.toml")
		if err := tools.WriteFileIfAbsent(pathAirToml, stage2.AirTomlContent()); err != nil {
			return fmt.Errorf("- [KO] création .air.toml: %v", err)
		} else {
			fmt.Println("- [OK] création .air.toml -")
		}
	}

	// Créer .env et .env.dist
	{
		pathEnv := filepath.Join(pathFolderApi, ".env")
		pathEnvDist := filepath.Join(pathFolderApi, ".env.dist")
		if err := tools.WriteFileIfAbsent(pathEnv, stage2.EnvApiContent(nameFolderProject, hostTraefikApi)); err != nil {
			return fmt.Errorf("- [KO] création .env: %v", err)
		} else {
			fmt.Println("- [OK] création .env -")
		}
		if err := tools.WriteFileIfAbsent(pathEnvDist, stage2.EnvApiContent(nameFolderProject, hostTraefikApi)); err != nil {
			return fmt.Errorf("- [KO] création .env.dist: %v", err)
		} else {
			fmt.Println("- [OK] création .env.dist -")
		}
	}

	// Créer .gitignore
	{
		pathGitignore := filepath.Join(pathFolderApi, ".gitignore")
		if err := tools.WriteFileIfAbsent(pathGitignore, stage2.GitignoreApiContent()); err != nil {
			return fmt.Errorf("- [KO] création .gitignore: %v", err)
		} else {
			fmt.Println("- [OK] création .gitignore -")
		}
	}

	// Créer tmp/.gitkeep
	{
		pathTmpDir := filepath.Join(pathFolderApi, "tmp")
		pathGitkeep := filepath.Join(pathTmpDir, ".gitkeep")
		if err := tools.EnsureDir(pathTmpDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier tmp: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathGitkeep, ""); err != nil {
			return fmt.Errorf("- [KO] création tmp/.gitkeep: %v", err)
		} else {
			fmt.Println("- [OK] création tmp/.gitkeep -")
		}
	}

	// Créer tmp/air/.gitkeep
	{
		pathTmpDir := filepath.Join(pathFolderApi, "tmp", "air")
		pathGitkeep := filepath.Join(pathTmpDir, ".gitkeep")
		if err := tools.EnsureDir(pathTmpDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier tmp: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathGitkeep, ""); err != nil {
			return fmt.Errorf("- [KO] création tmp/air/.gitkeep: %v", err)
		} else {
			fmt.Println("- [OK] création tmp/air/.gitkeep -")
		}
	}

	// Créer tmp/air/api/.gitkeep
	{
		pathTmpDir := filepath.Join(pathFolderApi, "tmp", "air", "api")
		pathGitkeep := filepath.Join(pathTmpDir, ".gitkeep")
		if err := tools.EnsureDir(pathTmpDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier tmp: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathGitkeep, ""); err != nil {
			return fmt.Errorf("- [KO] création tmp/air/api/.gitkeep: %v", err)
		} else {
			fmt.Println("- [OK] création tmp/air/api/.gitkeep -")
		}
	}

	// Créer cmd/api/main.go
	{
		pathCmdApiDir := filepath.Join(pathFolderApi, "cmd", "api")
		pathMainGo := filepath.Join(pathCmdApiDir, "main.go")
		if err := tools.EnsureDir(pathCmdApiDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier cmd/api: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathMainGo, stage2.MainGoContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création cmd/api/main.go: %v", err)
		} else {
			fmt.Println("- [OK] création cmd/api/main.go -")
		}
	}

	// Créer infrastructure adapters
	{
		// GinAdapter
		pathGinAdapterDir := filepath.Join(pathFolderApi, "internal", "infrastructure", "adapter", "ginadapter")
		pathGinAdapter := filepath.Join(pathGinAdapterDir, "GinAdapter.go")
		if err := tools.EnsureDir(pathGinAdapterDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier ginadapter: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathGinAdapter, stage2.GinAdapterContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création GinAdapter.go: %v", err)
		} else {
			fmt.Println("- [OK] création GinAdapter.go -")
		}

		// LoggerAdapter
		pathLoggerAdapterDir := filepath.Join(pathFolderApi, "internal", "infrastructure", "adapter", "loggeradapter")
		pathLoggerAdapter := filepath.Join(pathLoggerAdapterDir, "LoggerAdapter.go")
		if err := tools.EnsureDir(pathLoggerAdapterDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier loggeradapter: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathLoggerAdapter, stage2.LoggerAdapterContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création LoggerAdapter.go: %v", err)
		} else {
			fmt.Println("- [OK] création LoggerAdapter.go -")
		}

		// MongoAdapter
		pathMongoAdapterDir := filepath.Join(pathFolderApi, "internal", "infrastructure", "adapter", "mongoadapter")
		pathMongoAdapter := filepath.Join(pathMongoAdapterDir, "MongoAdapter.go")
		if err := tools.EnsureDir(pathMongoAdapterDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier mongoadapter: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathMongoAdapter, stage2.MongoAdapterContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création MongoAdapter.go: %v", err)
		} else {
			fmt.Println("- [OK] création MongoAdapter.go -")
		}
	}

	// Créer application gateways
	{
		// HttpGateway
		pathHttpGatewayDir := filepath.Join(pathFolderApi, "internal", "application", "gateway", "httpgateway")
		pathHttpGateway := filepath.Join(pathHttpGatewayDir, "HttpGateway.go")
		if err := tools.EnsureDir(pathHttpGatewayDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier httpgateway: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathHttpGateway, stage2.HttpGatewayContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création HttpGateway.go: %v", err)
		} else {
			fmt.Println("- [OK] création HttpGateway.go -")
		}

		// LogGateway
		pathLogGatewayDir := filepath.Join(pathFolderApi, "internal", "application", "gateway", "loggateway")
		pathLogGateway := filepath.Join(pathLogGatewayDir, "LoggerGateway.go")
		if err := tools.EnsureDir(pathLogGatewayDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier loggateway: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathLogGateway, stage2.LogGatewayContent()); err != nil {
			return fmt.Errorf("- [KO] création LoggerGateway.go: %v", err)
		} else {
			fmt.Println("- [OK] création LoggerGateway.go -")
		}

		// DbGateway
		pathDbGatewayDir := filepath.Join(pathFolderApi, "internal", "application", "gateway", "dbgateway")
		pathDbGateway := filepath.Join(pathDbGatewayDir, "DbGateway.go")
		if err := tools.EnsureDir(pathDbGatewayDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier dbgateway: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathDbGateway, stage2.DbGatewayContent()); err != nil {
			return fmt.Errorf("- [KO] création DbGateway.go: %v", err)
		} else {
			fmt.Println("- [OK] création DbGateway.go -")
		}
	}

	// Créer use cases
	{
		// NsevenUseCase
		pathNsevenUseCaseDir := filepath.Join(pathFolderApi, "internal", "application", "usecase", "nsevenusecase")
		pathNsevenUseCase := filepath.Join(pathNsevenUseCaseDir, "NsevenUseCase.go")
		if err := tools.EnsureDir(pathNsevenUseCaseDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier nsevenusecase: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathNsevenUseCase, stage2.NsevenUseCaseContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création NsevenUseCase.go: %v", err)
		} else {
			fmt.Println("- [OK] création NsevenUseCase.go -")
		}
	}

	// Créer controllers
	{
		// TestController
		pathTestControllerDir := filepath.Join(pathFolderApi, "internal", "application", "controller", "testcontroller")
		pathTestController := filepath.Join(pathTestControllerDir, "Controller.go")
		pathTestSayHello := filepath.Join(pathTestControllerDir, "SayHello.go")
		if err := tools.EnsureDir(pathTestControllerDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier testcontroller: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathTestController, stage2.TestControllerContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création testcontroller/Controller.go: %v", err)
		} else {
			fmt.Println("- [OK] création testcontroller/Controller.go -")
		}
		if err := tools.WriteFileIfAbsent(pathTestSayHello, stage2.TestSayHelloContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création testcontroller/SayHello.go: %v", err)
		} else {
			fmt.Println("- [OK] création testcontroller/SayHello.go -")
		}

		// NsevenController
		pathNsevenControllerDir := filepath.Join(pathFolderApi, "internal", "application", "controller", "nsevencontroller")
		pathNsevenController := filepath.Join(pathNsevenControllerDir, "Controller.go")
		pathNsevenCreate := filepath.Join(pathNsevenControllerDir, "CreateNseven.go")
		pathNsevenGetAll := filepath.Join(pathNsevenControllerDir, "GetAllNseven.go")
		if err := tools.EnsureDir(pathNsevenControllerDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier nsevencontroller: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathNsevenController, stage2.NsevenControllerContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création nsevencontroller/Controller.go: %v", err)
		} else {
			fmt.Println("- [OK] création nsevencontroller/Controller.go -")
		}
		if err := tools.WriteFileIfAbsent(pathNsevenCreate, stage2.NsevenCreateContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création nsevencontroller/CreateNseven.go: %v", err)
		} else {
			fmt.Println("- [OK] création nsevencontroller/CreateNseven.go -")
		}
		if err := tools.WriteFileIfAbsent(pathNsevenGetAll, stage2.NsevenGetAllContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création nsevencontroller/GetAllNseven.go: %v", err)
		} else {
			fmt.Println("- [OK] création nsevencontroller/GetAllNseven.go -")
		}
	}

	// Créer domain
	{
		pathNsevenDomainDir := filepath.Join(pathFolderApi, "internal", "domain", "nseven")
		pathNsevenEntity := filepath.Join(pathNsevenDomainDir, "Nseven.go")
		pathNsevenRepoInterface := filepath.Join(pathNsevenDomainDir, "NsevenRepositoryInterface.go")
		if err := tools.EnsureDir(pathNsevenDomainDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier domain/nseven: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathNsevenEntity, stage2.NsevenEntityContent()); err != nil {
			return fmt.Errorf("- [KO] création Nseven.go: %v", err)
		} else {
			fmt.Println("- [OK] création Nseven.go -")
		}
		if err := tools.WriteFileIfAbsent(pathNsevenRepoInterface, stage2.NsevenRepositoryInterfaceContent()); err != nil {
			return fmt.Errorf("- [KO] création NsevenRepositoryInterface.go: %v", err)
		} else {
			fmt.Println("- [OK] création NsevenRepositoryInterface.go -")
		}
	}

	// Créer repository
	{
		pathNsevenRepositoryDir := filepath.Join(pathFolderApi, "internal", "infrastructure", "repository", "nsevenrepository")
		pathNsevenRepository := filepath.Join(pathNsevenRepositoryDir, "MongoNsevenRepository.go")
		if err := tools.EnsureDir(pathNsevenRepositoryDir); err != nil {
			return fmt.Errorf("- [KO] création du dossier nsevenrepository: %v", err)
		}
		if err := tools.WriteFileIfAbsent(pathNsevenRepository, stage2.NsevenMongoRepositoryContent(moduleName)); err != nil {
			return fmt.Errorf("- [KO] création MongoNsevenRepository.go: %v", err)
		} else {
			fmt.Println("- [OK] création MongoNsevenRepository.go -")
		}
	}

	// Exécuter go mod tidy
	{
		fmt.Println("------ Nettoyage des dépendances Go ------")
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = pathFolderApi
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("- [KO] exécution de go mod tidy: %w", err)
		} else {
			fmt.Println("- [OK] go mod tidy exécuté -")
		}
	}

	return nil
}

func installDependenciesFront() error {
	fmt.Println("------ Installation des dépendances front ------")

	{
		cmd := exec.Command("pnpm", "add",
			"@astrojs/node", "astro", "@tailwindcss/vite", "tailwindcss")
		cmd.Dir = pathFolderFront
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("- [KO] erreur lors de l'installation des dépendances: %w", err)
		}
	}

	{
		cmd := exec.Command("pnpm", "add", "-D",
			"@astrojs/check", "typescript")
		cmd.Dir = pathFolderFront
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("- [KO] erreur lors de l'installation des dépendances de dev: %w", err)
		}
	}

	fmt.Println("- [OK] Dépendances installées")

	fmt.Println("------ Suppression node_modules pour utiliser celui créé par docker ------")

	{
		cmd := exec.Command("rm", "-rf", "node_modules")
		cmd.Dir = pathFolderFront
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("- [KO] impossible de supprimer les node_modules : %w", err)
		} else {
			fmt.Println("- [OK] node_modules supprimé -")
		}
	}

	return nil
}

func createAndSetFront() error {
	// installation projet front astro
	{
		if err := framework.RunAstroSsrCreate(nameServiceFront, pathFolderProject); err != nil {
			return fmt.Errorf("- [KO] échec création projet Astro SSR: %v", err)
		} else {
			fmt.Println("- [OK] création du projet front -")
		}
	}

	fmt.Println("------ Personnalisation du projet front ------")

	// modification du astro.config.mjs
	{
		allowedHost := []string{".local"}
		pathAstroConfig := filepath.Join(pathFolderFront, "astro.config.mjs")
		if err := tools.WriteFileAlways(pathAstroConfig, stage2.AstroConfigContent(portLinkTraefik, allowedHost)); err != nil {
			return fmt.Errorf("- [KO] écriture front/astro.config.mjs: %v", err)
		} else {
			fmt.Println("- [OK] modification front/astro.config.mjs -")
		}
	}

	// modification du package.json
	{
		pathPackageJson := filepath.Join(pathFolderFront, "package.json")
		if err = stage2.ReplacePackageJsonScripts(pathPackageJson, stage2.PackageJsonScriptContent()); err != nil {
			return fmt.Errorf("- [KO] écriture front/package-lock.json: %v", err)
		} else {
			fmt.Println("- [OK] modification front/package.json -")
		}
	}

	// creation du entrypoint
	{
		pathEntrypoint := filepath.Join(pathFolderFront, "entrypoint.sh")
		if err = tools.WriteFileIfAbsent(pathEntrypoint, stage2.EntrypointFrontContent()); err != nil {
			return fmt.Errorf("- [KO] écriture front/entrypoint.sh: %v", err)
		} else {
			fmt.Println("- [OK] création front/entrypoint.sh -")
		}
	}

	// creation du styles/global.css
	{
		pathStylesDir := filepath.Join(pathFolderFront, "src", "styles")
		pathGlobalcss := filepath.Join(pathStylesDir, "global.css")

		// Créer le dossier styles s'il n'existe pas
		if err := os.MkdirAll(pathStylesDir, 0755); err != nil {
			return fmt.Errorf("- [KO] création du dossier styles: %v", err)
		}

		if err = tools.WriteFileIfAbsent(pathGlobalcss, "@import \"tailwindcss\";"); err != nil {
			return fmt.Errorf("- [KO] écriture front/src/styles/global.css: %v", err)
		} else {
			fmt.Println("- [OK] création front/src/styles/global.css -")
		}
	}

	// creation des env front
	{
		appEnv := filepath.Join(pathFolderFront, ".env")
		appEnvDist := filepath.Join(pathFolderFront, ".env.dist")
		if err = tools.WriteFileIfAbsent(appEnv, stage2.EnvFrontContent()); err != nil {
			return fmt.Errorf("- [KO] création front/.env : %v", err)
		} else {
			fmt.Println("- [OK] création front/.env -")
		}

		if err = tools.WriteFileIfAbsent(appEnvDist, stage2.EnvFrontContent()); err != nil {
			return fmt.Errorf("- [KO] création front/.env.dist : %v", err)
		} else {
			fmt.Println("- [OK] création front/.env.dist -")
		}
	}

	return nil
}

func validateUserForStart() error {
	fmt.Println("Stage-2: création du projet avec ses données")
	fmt.Printf("- Path du projet: %v\n", pathFolderProject)
	fmt.Printf("- Dossier du projet: %v\n", nameFolderProject)
	fmt.Printf("- Path du front: %v\n", pathFolderFront)
	fmt.Printf("- Dossier du front: %v\n", nameServiceFront)
	fmt.Printf("- Host du traefik front: %v\n", hostTraefikFront)
	fmt.Printf("- Host du traefik Api: %v\n", hostTraefikApi)
	fmt.Printf("- Version de node: %v\n", nodeVersion)
	fmt.Printf("- Port pour tout les services traefik: %v\n", portLinkTraefik)

	// validation des données de creation
	if tools.AskYesNo(fmt.Sprintf(" Est ce que ses valeurs vous conviennent ? [o/N]: "), true) {
		fmt.Printf("------ Vérification des prérequis ------\n")

		// Vérifier Node.js
		nodeCmd := exec.Command("node", "--version")
		nodeOutput, err := nodeCmd.Output()
		if err != nil {
			return fmt.Errorf(" Node.js n'est pas installé sur cette machine")
		}
		installedNodeVersion := strings.TrimSpace(strings.TrimPrefix(string(nodeOutput), "v"))

		// Vérifier pnpm (juste disponibilité)
		pnpmCmd := exec.Command("pnpm", "--version")
		_, err = pnpmCmd.Output()
		pnpmInstalled := err == nil

		// Comparer la version de Node
		nodeOk := tools.CompareVersion(installedNodeVersion, nodeVersion)

		if nodeOk && pnpmInstalled {
			fmt.Printf("✓ Node.js %s (requis: >= %s)\n", installedNodeVersion, nodeVersion)
			fmt.Printf("✓ pnpm est installé\n")
			fmt.Println("\n------ Initialisation du projet ------")

			if tools.AskYesNo(fmt.Sprintf("  Lancer la création du projet Astro ? [o/N]: "), true) {
				fmt.Printf("- Lancement: pnpm create astro@latest %s\n", nameServiceFront)
			} else {
				return fmt.Errorf("commande annulée par l'utilisateur")
			}
		} else {
			errMsg := "Prérequis manquants:\n"
			if !nodeOk {
				errMsg += fmt.Sprintf("  ✗ Node.js installé: %s (requis: >= %s)\n", installedNodeVersion, nodeVersion)
			}
			if !pnpmInstalled {
				errMsg += "  ✗ pnpm n'est pas installé. Installer avec: npm install -g pnpm\n"
			}
			return fmt.Errorf("%v", errMsg)
		}
	} else {
		return fmt.Errorf("commande annulée: les valeurs définis ne conviennent pas")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(starter2)

	starter2.Flags().StringVar(&hostTraefikFront, "hostFront", "", "format: host.extension => (requis) ")
	_ = starter2.MarkFlagRequired("hostFront")

	starter2.Flags().StringVar(&hostTraefikApi, "hostApi", "", "format: host.extension => (requis) ")
	_ = starter2.MarkFlagRequired("hostApi")
}
