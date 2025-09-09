package framework

import (
	"errors"
	"fmt"
	"github.com/nsevendev/starter/internal/docker"
	"os"
	"os/exec"
)

// RunAngularSsrCreate exécute la commande Angular CLI pour créer un nouveau projet avec SSR
// Utilise 'ng' si disponible, sinon 'npx @angular/cli@latest'
func RunAngularSsrCreate(projectName, workdir string) error {
	hasNg := docker.HasCommand("ng")
	hasNpx := docker.HasCommand("npx")

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
