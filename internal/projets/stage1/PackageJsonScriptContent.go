package stage1

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReplacePackageJSONScripts remplace la clé "scripts" dans un package.json
// par la map fournie (remplacement complet, pas de fusion).
// Ex: ReplacePackageJSONScripts("package.json", map[string]string{"start": "node server.js"})
func ReplacePackageJSONScripts(path string, newScripts map[string]string) error {
	// lire package.json
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("lecture %s: %w", path, err)
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

	// réécrire le script
	out, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	out = append(out, '\n')
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("écriture %s: %w", path, err)
	}
	return nil
}
