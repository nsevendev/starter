package stage2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func ReplacePackageJsonScripts(pathFilePackageJson string, newScripts map[string]string) error {
	// lire package.json
	data, err := os.ReadFile(pathFilePackageJson)
	if err != nil {
		return fmt.Errorf("lecture %s: %w", pathFilePackageJson, err)
	}

	// parser en map générique pour ne toucher qu'à "scripts"
	var pkg map[string]any
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	// remplacer ENTIEREMENT la clé "scripts"
	scripts := make(map[string]any, len(newScripts))
	for k, v := range newScripts {
		scripts[k] = v
	}
	pkg["scripts"] = scripts

	// réécrire le script SANS échapper les caractères HTML
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(pkg); err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(pathFilePackageJson, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("écriture %s: %w", pathFilePackageJson, err)
	}
	return nil
}
