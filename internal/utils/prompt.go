package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmPrompt asks the user for confirmation
func ConfirmPrompt(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N] ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
