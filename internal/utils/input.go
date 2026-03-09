package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// isReplMode checks if the code is running in REPL mode.
func isReplMode() bool {
	return os.Getenv("JOTR_REPL_MODE") == "true"
}

// PromptUser prompts the user with the given message and returns the input.
// This is a safer alternative to fmt.Scanln that handles empty input and trimming.
// Returns an empty string if the user provides no input.
// Returns an error if called in REPL mode.
func PromptUser(prompt string) (string, error) {
	if isReplMode() {
		return "", errors.New("interactive prompts are not supported in REPL mode")
	}

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

// PromptUserRequired prompts the user with the given message and returns the input.
// Unlike PromptUser, this keeps prompting until non-empty input is received.
// Returns an error if called in REPL mode.
func PromptUserRequired(prompt string) (string, error) {
	if isReplMode() {
		return "", errors.New("interactive prompts are not supported in REPL mode")
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		trimmed := strings.TrimSpace(input)
		if trimmed != "" {
			return trimmed, nil
		}

		fmt.Print("Input required, please try again: ")
	}
}

// PromptYesNo prompts the user with a yes/no question.
// Returns true for "yes" (y/Y), false for "no" (n/N).
// Keeps prompting until a valid response is received.
// Returns an error if called in REPL mode.
func PromptYesNo(prompt string) (bool, error) {
	if isReplMode() {
		return false, errors.New("interactive prompts are not supported in REPL mode")
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		trimmed := strings.ToLower(strings.TrimSpace(input))
		if trimmed == "y" || trimmed == "yes" {
			return true, nil
		}
		if trimmed == "n" || trimmed == "no" || trimmed == "" {
			return false, nil
		}

		fmt.Print("Please enter 'y' or 'n': ")
	}
}

// PromptChoice prompts the user to choose from a list of options.
// Returns the index of the chosen option (0-based), or -1 for invalid input.
// Returns an error if called in REPL mode.
func PromptChoice(prompt string, min, max int) (int, error) {
	if isReplMode() {
		return -1, errors.New("interactive prompts are not supported in REPL mode")
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		trimmed := strings.TrimSpace(input)
		if trimmed == "" {
			return -1, nil
		}

		var choice int
		_, err = fmt.Sscanf(trimmed, "%d", &choice)
		if err != nil {
			fmt.Printf("Please enter a number between %d and %d: ", min, max)
			continue
		}

		if choice >= min && choice <= max {
			return choice, nil
		}

		fmt.Printf("Please enter a number between %d and %d: ", min, max)
	}
}
