package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	projectName string
	version     string
)

var initTempAngssrGo = &cobra.Command{
	Use:   "init-temp-angssr-go",
	Short: "initialise un projet angular ssr avec go, mongo",
	Long:  `Initialise un projet angular ssr avec go, mongo, redis, docker, r2, mailer, etc...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// validation des flags
		if projectName == "" {
			return fmt.Errorf("le flag --name est requis")
		}
		if version == "" {
			return fmt.Errorf("le flag --version est requis (ex: --version=v1.0.0)")
		}

		// chemin du projet (dossier courant + nom du projet)
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("erreur lors de la récupération du dossier courant: %w", err)
		}
		projectPath := filepath.Join(currentDir, projectName)

		// check si le dossier existe deja
		if _, err := os.Stat(projectPath); err == nil {
			return fmt.Errorf("le dossier %s existe déjà", projectName)
		}

		// clone du repo avec la version specifiée
		fmt.Printf("Clonage du template (version %s)...\n", version)
		cloneCmd := exec.Command("git", "clone", "--branch", version, "--depth", "1",
			"https://github.com/nsevendev/temp-angssr-go.git", projectName)
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

		fmt.Printf("Projet %s créé avec succès !\n", projectName)
		return nil
	},
}

func init() {
	initTempAngssrGo.Flags().StringVar(&projectName, "name", "", "nom du projet (requis)")
	initTempAngssrGo.Flags().StringVar(&version, "version", "", "version du template (ex: v1.0.0) (requis)")

	rootCmd.AddCommand(initTempAngssrGo)
}
