package cmd

import (
	"fmt"
	"os"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Health check for jotr setup",
	Long:  `Run a health check to verify jotr configuration and setup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHealthCheck()
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runHealthCheck() error {
	fmt.Println("🔍 jotr Health Check")
	fmt.Println("====================\n")

	allGood := true

	// Check 1: Config file exists
	fmt.Print("Config file... ")
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Error: %v\n", err)
		allGood = false
		return nil // Can't continue without config
	}
	fmt.Println("✓ OK")

	// Check 2: Base directory exists
	fmt.Print("Base directory... ")
	if _, err := os.Stat(cfg.Paths.BaseDir); os.IsNotExist(err) {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Directory not found: %s\n", cfg.Paths.BaseDir)
		allGood = false
	} else {
		fmt.Println("✓ OK")
		fmt.Printf("  %s\n", cfg.Paths.BaseDir)
	}

	// Check 3: Diary directory exists
	fmt.Print("Diary directory... ")
	if _, err := os.Stat(cfg.DiaryPath); os.IsNotExist(err) {
		fmt.Println("⚠️  WARNING")
		fmt.Printf("  Directory not found: %s\n", cfg.DiaryPath)
		fmt.Println("  Will be created when you create your first daily note")
	} else {
		fmt.Println("✓ OK")
		fmt.Printf("  %s\n", cfg.DiaryPath)
	}

	// Check 4: Todo file
	fmt.Print("Todo file... ")
	if !notes.FileExists(cfg.TodoPath) {
		fmt.Println("⚠️  WARNING")
		fmt.Printf("  File not found: %s\n", cfg.TodoPath)
		fmt.Println("  Will be created when you sync tasks")
	} else {
		fmt.Println("✓ OK")
		fmt.Printf("  %s\n", cfg.TodoPath)
	}

	// Check 5: PDP file (optional)
	if cfg.PDPPath != "" {
		fmt.Print("PDP file... ")
		if !notes.FileExists(cfg.PDPPath) {
			fmt.Println("⚠️  WARNING")
			fmt.Printf("  File not found: %s\n", cfg.PDPPath)
		} else {
			fmt.Println("✓ OK")
			fmt.Printf("  %s\n", cfg.PDPPath)
		}
	}

	// Check 6: Editor
	fmt.Print("Editor... ")
	editor := os.Getenv("EDITOR")
	if editor == "" {
		fmt.Println("⚠️  WARNING")
		fmt.Println("  EDITOR environment variable not set")
		fmt.Println("  Will default to 'nvim'")
	} else {
		fmt.Println("✓ OK")
		fmt.Printf("  %s\n", editor)
	}

	// Check 7: AI command (if enabled)
	if cfg.AI.Enabled {
		fmt.Print("AI command... ")
		if cfg.AI.Command == "" {
			fmt.Println("⚠️  WARNING")
			fmt.Println("  AI enabled but no command configured")
		} else {
			fmt.Println("✓ OK")
			fmt.Printf("  %s\n", cfg.AI.Command)
		}
	}

	// Summary
	fmt.Println()
	if allGood {
		fmt.Println("✅ All checks passed!")
	} else {
		fmt.Println("⚠️  Some checks failed. Please review the issues above.")
	}

	// Show config summary
	fmt.Println("\nConfiguration Summary:")
	fmt.Printf("  Task Section: %s\n", cfg.Format.TaskSection)
	fmt.Printf("  Capture Section: %s\n", cfg.Format.CaptureSection)
	fmt.Printf("  Daily Note Sections: %v\n", cfg.Format.DailyNoteSections)
	fmt.Printf("  Include Weekends: %v\n", cfg.Streaks.IncludeWeekends)

	return nil
}

