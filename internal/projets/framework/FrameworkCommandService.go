package framework

import (
	"errors"
	"fmt"
	"github.com/nsevendev/starter/internal/docker"
	"os"
	"os/exec"
	"path/filepath"
)

func RunAstroSsrCreate(nameServiceFront, workdir string) error {
	// creation commande installation projet Astro SSR
	cmd := exec.Command("pnpm", "create", "astro@latest", nameServiceFront)
	cmd.Dir = workdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("échec 'pnpm create astro': %w. Guide: installer Node.js >= 22 et npm", err)
	}

	return nil
}

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

	// creation commande installation projet Angular SSR avec le cli
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
			return fmt.Errorf("'ng' et 'npx' ont échoué: %v / %v. Guide: installer Node.js >= 22 et Angular CLI: npm install -g @angular/cli", err, err2)
		}
	}
	return nil
}

// InstallTailwindAndSetup installe @tailwindcss/postcss et configure PostCSS + styles.css dans l'app
func InstallTailwindAndSetup(appDir string) error {
	// installation tailwindcss/postcss dans le dossier de l'app
	cmd := exec.Command("npm", "install", "-D", "@tailwindcss/postcss")
	cmd.Dir = appDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("échec installation tailwindcss/postcss: %w", err)
	}
	fmt.Println("- [OK] installation @tailwindcss/postcss -")

	// créer postcss.config.json à la racine de l'app
	// a mettre à la place de la string en dessous { "plugins": { "@tailwindcss/postcss": {} }}
	// creer un fichier json à la place du js
	postcssConfig := "{\n  \"plugins\": {\n    \"@tailwindcss/postcss\": {}\n  }\n}\n"
	postcssConfigPath := filepath.Join(appDir, "postcss.config.json")
	if err := os.WriteFile(postcssConfigPath, []byte(postcssConfig), 0o644); err != nil {
		return fmt.Errorf("échec écriture %s: %w", postcssConfigPath, err)
	}
	fmt.Println("- [OK] création app/postcss.config.json -")

	// s'assurer que src/styles.css importe Tailwind
	stylesPath := filepath.Join(appDir, "src", "styles.css")
	content := "@import \"tailwindcss\";\n"
	if err := os.WriteFile(stylesPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("échec écriture %s: %w", stylesPath, err)
	}
	fmt.Println("- [OK] fichier app/src/styles.css écrasé avec import Tailwind -")

	return nil
}
