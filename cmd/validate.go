package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and notes",
	Long:  `Validate jotr configuration and check for issues in notes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runValidation()
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidation() error {
	fmt.Println("🔍 Validating jotr...")
	fmt.Println()

	allGood := true

	// Check 1: Config file exists
	fmt.Print("Config file exists... ")
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "jotr", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Not found: %s\n", configPath)
		allGood = false
		return nil
	}
	fmt.Println("✓ OK")

	// Check 2: Valid JSON
	fmt.Print("Valid JSON... ")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Cannot read: %v\n", err)
		allGood = false
		return nil
	}

	var configData map[string]interface{}
	if err := json.Unmarshal(data, &configData); err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Invalid JSON: %v\n", err)
		allGood = false
		return nil
	}
	fmt.Println("✓ OK")

	// Check 3: Load config
	fmt.Print("Load configuration... ")
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Error: %v\n", err)
		allGood = false
		return nil
	}
	fmt.Println("✓ OK")

	// Check 4: Base directory
	fmt.Print("Base directory... ")
	if _, err := os.Stat(cfg.Paths.BaseDir); os.IsNotExist(err) {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Not found: %s\n", cfg.Paths.BaseDir)
		allGood = false
	} else {
		fmt.Println("✓ OK")
		fmt.Printf("  %s\n", cfg.Paths.BaseDir)
	}

	// Check 5: Diary directory
	fmt.Print("Diary directory... ")
	if _, err := os.Stat(cfg.DiaryPath); os.IsNotExist(err) {
		fmt.Println("⚠️  WARNING")
		fmt.Printf("  Not found: %s\n", cfg.DiaryPath)
	} else {
		fmt.Println("✓ OK")
	}

	// Check 6: Todo file
	fmt.Print("Todo file... ")
	if !notes.FileExists(cfg.TodoPath) {
		fmt.Println("⚠️  WARNING")
		fmt.Printf("  Not found: %s\n", cfg.TodoPath)
	} else {
		fmt.Println("✓ OK")
	}

	// Check 7: Validate notes structure
	fmt.Print("Notes structure... ")
	allNotes, err := notes.FindNotes(cfg.Paths.BaseDir)
	if err != nil {
		fmt.Println("⚠️  WARNING")
		fmt.Printf("  Error finding notes: %v\n", err)
	} else {
		fmt.Println("✓ OK")
		fmt.Printf("  Found %d notes\n", len(allNotes))
	}

	// Check 8: Check for broken links (basic)
	fmt.Print("Checking for issues... ")
	issueCount := 0
	for _, notePath := range allNotes {
		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		// Check for empty files
		if len(content) == 0 {
			issueCount++
		}
	}

	if issueCount > 0 {
		fmt.Printf("⚠️  %d issues found\n", issueCount)
	} else {
		fmt.Println("✓ OK")
	}

	// Summary
	fmt.Println()
	if allGood && issueCount == 0 {
		fmt.Println("✅ Validation passed!")
	} else {
		fmt.Println("⚠️  Validation completed with warnings")
	}

	return nil
}

