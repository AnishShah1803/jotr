package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptUser prompts the user with the given message and returns the input.
// This is a safer alternative to fmt.Scanln that handles empty input and trimming.
// Returns an empty string if the user provides no input.
func PromptUser(prompt string) string {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}

	return strings.TrimSpace(input)
}

// PromptUserRequired prompts the user with the given message and returns the input.
// Unlike PromptUser, this keeps prompting until non-empty input is received.
func PromptUserRequired(prompt string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		trimmed := strings.TrimSpace(input)
		if trimmed != "" {
			return trimmed
		}

		fmt.Print("Input required, please try again: ")
	}
}

// PromptYesNo prompts the user with a yes/no question.
// Returns true for "yes" (y/Y), false for "no" (n/N).
// Keeps prompting until a valid response is received.
func PromptYesNo(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		trimmed := strings.ToLower(strings.TrimSpace(input))
		if trimmed == "y" || trimmed == "yes" {
			return true
		}
		if trimmed == "n" || trimmed == "no" || trimmed == "" {
			return false
		}

		fmt.Print("Please enter 'y' or 'n': ")
	}
}

// PromptChoice prompts the user to choose from a list of options.
// Returns the index of the chosen option (0-based), or -1 for invalid input.
func PromptChoice(prompt string, min, max int) int {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		trimmed := strings.TrimSpace(input)
		if trimmed == "" {
			return -1
		}

		var choice int
		_, err = fmt.Sscanf(trimmed, "%d", &choice)
		if err != nil {
			fmt.Printf("Please enter a number between %d and %d: ", min, max)
			continue
		}

		if choice >= min && choice <= max {
			return choice
		}

		fmt.Printf("Please enter a number between %d and %d: ", min, max)
	}
}
