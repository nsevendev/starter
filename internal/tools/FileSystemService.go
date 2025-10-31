package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SanitizeName convertit une chaîne en un nom de fichier/dossier sûr
func SanitizeName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9-_]+`)
	s = re.ReplaceAllString(s, "")
	s = strings.Trim(s, "-")
	return s
}

// EnsureDir creation d'un dossier s'il n'existe pas
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("création du dossier %s: %w", path, err)
	}
	return nil
}

// WriteFileAlways créer/écrase le fichier s'il existe
func WriteFileAlways(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("écriture du fichier %s: %w", path, err)
	}
	fmt.Printf("  (écrasé) %s\n", path)
	return nil
}

// WriteFileIfAbsent créer le fichier uniquement s'il n'existe pas
func WriteFileIfAbsent(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  (skip) %s existe déjà\n", path)
		return nil
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("erreur écriture du fichier %s: %w", path, err)
	}
	return nil
}

// ReplaceInFile effectue un remplacement de texte dans un fichier
func ReplaceInFile(path, old, new string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("lecture du fichier %s: %w", path, err)
	}

	newContent := strings.ReplaceAll(string(content), old, new)

	if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("écriture du fichier %s: %w", path, err)
	}

	fmt.Printf("  ✓ %s modifié\n", path)
	return nil
}

// ReplaceInAllGoFiles parcourt récursivement un dossier et remplace du texte dans tous les fichiers .go
func ReplaceInAllGoFiles(rootDir, old, new string) error {
	var filesModified int

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignorer les dossiers
		if info.IsDir() {
			return nil
		}

		// Traiter uniquement les fichiers .go
		if filepath.Ext(path) != ".go" {
			return nil
		}

		// Lire le fichier
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("lecture du fichier %s: %w", path, err)
		}

		// Vérifier si le fichier contient le texte à remplacer
		contentStr := string(content)
		if !strings.Contains(contentStr, old) {
			return nil
		}

		// Remplacer le texte
		newContent := strings.ReplaceAll(contentStr, old, new)

		// Écrire le fichier modifié
		if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
			return fmt.Errorf("écriture du fichier %s: %w", path, err)
		}

		filesModified++
		fmt.Printf("  ✓ %s modifié\n", path)
		return nil
	})

	if err != nil {
		return err
	}

	if filesModified > 0 {
		fmt.Printf("  Total: %d fichier(s) Go modifié(s)\n", filesModified)
	}

	return nil
}
