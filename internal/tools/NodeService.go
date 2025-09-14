package tools

import (
	"fmt"
	"os"
	"time"
)

func DeletePackageLock(packageLockPath string) {
	if err := os.Remove(packageLockPath); err != nil {
		// petit retry simple pour les fichiers aussi
		time.Sleep(100 * time.Millisecond)
		if err2 := os.Remove(packageLockPath); err2 != nil {
			fmt.Printf("- [KO] suppression app/package-lock.json - err: %v\n", err2)
		} else {
			fmt.Println("- [OK] suppression app/package-lock.json -")
		}
	} else {
		fmt.Println("- [OK] suppression app/package-lock.json -")
	}
}
func DeleteNodeModules(nodeModulesPath string) {
	if _, statErr := os.Stat(nodeModulesPath); statErr == nil {
		// Stratégie robuste: renommer le dossier pour le détacher, puis supprimer en retries.
		retries := 5
		delay := 200 * time.Millisecond
		for attempt := 1; attempt <= retries; attempt++ {
			// Tente un renommage pour éviter les recréations concurrentes (ex: .DS_Store)
			tmpPath := nodeModulesPath + fmt.Sprintf(".__to_delete__%d", time.Now().UnixNano())
			renameTarget := nodeModulesPath
			if err := os.Rename(nodeModulesPath, tmpPath); err == nil {
				renameTarget = tmpPath
			}

			// Supprime récursivement la cible (renommée si possible)
			if err := os.RemoveAll(renameTarget); err != nil {
				fmt.Printf("- [KO] tentative %d suppression app/node_modules - err: %v\n", attempt, err)
			} else {
				// Vérifie si le dossier d'origine existe encore (recréation potentielle)
				if _, still := os.Stat(nodeModulesPath); os.IsNotExist(still) {
					fmt.Println("- [OK] suppression app/node_modules -")
					break
				}
			}

			if attempt < retries {
				time.Sleep(delay)
				continue
			}

			// Dernière vérification après retries
			if _, still := os.Stat(nodeModulesPath); still == nil {
				fmt.Println("- [KO] suppression app/node_modules - persiste après plusieurs tentatives")
			} else {
				fmt.Println("- [OK] suppression app/node_modules -")
			}
		}
	}
}
