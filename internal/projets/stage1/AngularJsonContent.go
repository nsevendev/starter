package stage1

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// ServeOptions options pour la commande "ng serve"
type ServeOptions struct {
	Host         string
	Port         int
	Poll         int
	AllowedHosts []string
}

// PatchOptions options pour PatchAngularJSON
type PatchOptions struct {
	AngularJSONPath  string        // ex: "angular.json"
	ProjectOldName   string        // ex: "test"  (dans le vierge)
	ProjectNewName   string        // ex: "app"
	OutputPath       string        // ex: "dist/app"
	BudgetStyleWarn  string        // ex: "500kB"
	BudgetStyleErr   string        // ex: "1MB"
	Serve            *ServeOptions // nil = ne pas toucher
	DisableAnalytics bool          // true => "cli.analytics": false
}

// PatchAngularJSON modifie un fichier angular.json selon les options fournies.
// - renomme le projet (projects.<old> -> projects.<new>)
// - modifie build.options.outputPath
// - ajoute/modifie un budget de type "anyComponentStyle" en prod
// - modifie serve.options (host, port, poll, allowedHosts)
// - modifie serve.configurations.production.buildTarget
// - ajoute "cli.analytics": false si demandé
//
// Ne modifie que les clés indiquées, les autres restent intactes.
func PatchAngularJSON(opts PatchOptions) error {
	data, err := os.ReadFile(opts.AngularJSONPath)
	if err != nil {
		return fmt.Errorf("lecture %s: %w", opts.AngularJSONPath, err)
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	projects, ok := root["projects"].(map[string]any)
	if !ok {
		return errors.New(`clé "projects" absente ou invalide`)
	}

	// récup projet (et éventuellement renommer la clé)
	projKey := opts.ProjectOldName
	if _, exists := projects[projKey]; !exists {
		// Si l’ancien nom n’existe pas, peut-être que le fichier a déjà le nouveau ?
		if _, exists2 := projects[opts.ProjectNewName]; exists2 {
			projKey = opts.ProjectNewName
		} else {
			return fmt.Errorf("projet %q introuvable dans projects", opts.ProjectOldName)
		}
	}

	projVal, ok := projects[projKey].(map[string]any)
	if !ok {
		return fmt.Errorf("projects.%s n’est pas un objet", projKey)
	}

	// renommer la clé si nécessaire
	if projKey != opts.ProjectNewName {
		projects[opts.ProjectNewName] = projVal
		delete(projects, projKey)
		projKey = opts.ProjectNewName
	}

	// vers architect/build/options
	architect := mustObj(projVal, "architect")
	build := mustObj(architect, "build")
	buildOptions := mustObj(build, "options")

	// outputPath
	if opts.OutputPath != "" {
		buildOptions["outputPath"] = opts.OutputPath
	}

	// vers budgets anyComponentStyle (production)
	configs := mustObj(build, "configurations")
	prod := mustObj(configs, "production")
	budgets := mustArr(prod, "budgets")

	// cherche un budget type anyComponentStyle ; sinon en crée un
	foundStyle := false
	for i := range budgets {
		if b, ok := budgets[i].(map[string]any); ok {
			if b["type"] == "anyComponentStyle" {
				if opts.BudgetStyleWarn != "" {
					b["maximumWarning"] = opts.BudgetStyleWarn
				}
				if opts.BudgetStyleErr != "" {
					b["maximumError"] = opts.BudgetStyleErr
				}
				foundStyle = true
				break
			}
		}
	}
	if !foundStyle && (opts.BudgetStyleWarn != "" || opts.BudgetStyleErr != "") {
		newB := map[string]any{
			"type": "anyComponentStyle",
		}
		if opts.BudgetStyleWarn != "" {
			newB["maximumWarning"] = opts.BudgetStyleWarn
		}
		if opts.BudgetStyleErr != "" {
			newB["maximumError"] = opts.BudgetStyleErr
		}
		budgets = append(budgets, newB)
		prod["budgets"] = budgets
	}

	// serve options + buildTarget
	serve := mustObj(architect, "serve")
	// options
	if opts.Serve != nil {
		serveOpts := mustObj(serve, "options")
		if opts.Serve.Host != "" {
			serveOpts["host"] = opts.Serve.Host
		}
		if opts.Serve.Port > 0 {
			serveOpts["port"] = opts.Serve.Port
		}
		if opts.Serve.Poll > 0 {
			serveOpts["poll"] = opts.Serve.Poll
		}
		if len(opts.Serve.AllowedHosts) > 0 {
			// convert []string -> []any pour JSON générique
			hosts := make([]any, 0, len(opts.Serve.AllowedHosts))
			for _, h := range opts.Serve.AllowedHosts {
				hosts = append(hosts, h)
			}
			serveOpts["allowedHosts"] = hosts
		}
	}
	// configurations.buildTarget
	serveCfg := mustObj(serve, "configurations")
	serveCfg["production"] = map[string]any{"buildTarget": fmt.Sprintf("%s:build:production", projKey)}
	serveCfg["development"] = map[string]any{"buildTarget": fmt.Sprintf("%s:build:development", projKey)}

	// cli.analytics
	if opts.DisableAnalytics {
		cliObj, _ := root["cli"].(map[string]any)
		if cliObj == nil {
			cliObj = map[string]any{}
		}
		cliObj["analytics"] = false
		root["cli"] = cliObj
	}

	// ecriture
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	if err := os.WriteFile(opts.AngularJSONPath, out, 0o644); err != nil {
		return fmt.Errorf("écriture %s: %w", opts.AngularJSONPath, err)
	}
	return nil
}

// mustObj retourne m[key] en tant que map[string]any.
// Si la clé n'existe pas, elle est créée avec un objet vide.
// Si la clé existe mais n'est pas un objet, elle est remplacée par un objet vide.
func mustObj(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok || v == nil {
		n := map[string]any{}
		m[key] = n
		return n
	}
	if mm, ok := v.(map[string]any); ok {
		return mm
	}
	// si le type est mauvais, on remplace par un objet vide (safe pour patcher)
	n := map[string]any{}
	m[key] = n
	return n
}

// mustArr retourne m[key] en tant que []any.
// Si la clé n'existe pas, elle est créée avec un tableau vide.
// Si la clé existe mais n'est pas un tableau, elle est remplacée par un tableau vide.
func mustArr(m map[string]any, key string) []any {
	v, ok := m[key]
	if !ok || v == nil {
		n := []any{}
		m[key] = n
		return n
	}
	if arr, ok := v.([]any); ok {
		return arr
	}
	n := []any{}
	m[key] = n
	return n
}
