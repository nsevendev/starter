package docker

import (
	"bytes"
	"fmt"
	"github.com/nsevendev/starter/internal/tools"
	"os"
	"os/exec"
	"strings"
)

// HasCommand checks si une commande est disponible dans le PATH
func HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// HasDocker checks si Docker est installé
func HasDocker() bool {
	if !HasCommand("docker") {
		fmt.Println("[WARN] Docker introuvable. Installez Docker Desktop / Docker Engine.")
		return false
	}
	fmt.Println("[OK] Docker est présent.")
	return true
}

// HasDockerCompose checks si Docker Compose est installé
func HasDockerCompose() (bool, bool) {
	// pas de docker
	if !HasDocker() {
		return false, false
	}
	// essaie docker compose
	cmd := exec.Command("docker", "compose", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		for _, line := range lines {
			fmt.Printf("[OK] %s\n", line)
		}
		return true, false
	}
	// essaie ancien commande docker compose
	if HasCommand("docker-compose") {
		cmd2 := exec.Command("docker-compose", "--version")
		var out2 bytes.Buffer
		cmd2.Stdout = &out2
		cmd2.Stderr = &out2
		err2 := cmd2.Run()
		if err2 == nil {
			lines := strings.Split(strings.TrimSpace(out2.String()), "\n")
			for _, line := range lines {
				fmt.Printf("[OK] %s\n", line)
			}
			fmt.Println("[WARN] Vous avez une ancienne version de docker-compose")
			return false, true
		}
		fmt.Println("[WARN] Vous avez une ancienne version de docker-compose")
	}
	fmt.Println("[WARN] 'docker compose' ou 'docker-compose' introuvable.")
	return false, false
}

// DockerNetworkExists checks if a Docker network exists
func DockerNetworkExists(name string) bool {
	if !HasDocker() {
		return false
	}
	cmd := exec.Command("docker", "network", "inspect", name)
	return cmd.Run() == nil
}

// PrintDockerHints check docker et le reseau externe
func PrintDockerHints(project string) {
	network := "traefik-nseven"
	hasSub, hasBin := HasDockerCompose()

	if !HasDocker() {
		fmt.Println("[warn] Docker introuvable. Installez Docker Desktop / Docker Engine.")
	}
	if !hasSub && !hasBin {
		fmt.Println("[warn] 'docker compose' ou 'docker-compose' introuvable. Installez le plugin Compose ou utilisez Docker Desktop récent.")
	}
	if !DockerNetworkExists(network) {
		fmt.Printf("[INFO] Réseau externe '%s' absent. Créez-le avant de lancer: docker network create %s\n", network, network)
		if tools.AskYesNo(fmt.Sprintf("  Voulez vous creer le reseau %v => %v ? [o/N]: ", network), true) {
			cmd := exec.Command("docker", "network", "create", network)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("[ERROR] Échec de la création du réseau '%s': %v\n", network, err)
			} else {
				fmt.Printf("[OK] Réseau '%s' créé avec succès.\n", network)
			}
		} else {
			fmt.Printf("[INFO] Réseau '%s' non créé. Pensez à l'initialiser plus tard:\n  docker network create %s\n", network, network)
		}
	}
}
