package tools

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// AskYesNo pose une question oui/non à l'utilisateur et retourne true pour oui, false pour non
// Si l'utilisateur appuie sur Entrée sans rien saisir, la valeur par défaut est utilisée
// prompt : le message à afficher
// defaultNo : si true, la valeur par défaut est non, sinon oui
func AskYesNo(prompt string, defaultNo bool) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return !defaultNo
	}
	switch input {
	case "o", "oui", "y", "yes":
		return true
	case "n", "non", "no":
		return false
	default:
		return false
	}
}
