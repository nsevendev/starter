package tools

import (
	"fmt"
	"os"
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
		return fmt.Errorf("écriture du fichier %s: %w", path, err)
	}
	return nil
}
