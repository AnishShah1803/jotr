package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
)

var GitCmd = &cobra.Command{
	Use:   "git [action]",
	Short: "Git integration for notes",
	Long: `Git integration for version control of notes.
	
Actions:
  status        Show git status
  commit        Commit changes
  history       Show commit history
  diff          Show diff
  
Examples:
  jotr git status
  jotr git commit
  jotr git history`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("action required: status, commit, history, or diff")
		}

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		action := args[0]

		switch action {
		case "status":
			return gitStatus(cfg)
		case "commit":
			return gitCommit(cfg)
		case "history":
			return gitHistory(cfg)
		case "diff":
			return gitDiff(cfg)
		default:
			return fmt.Errorf("unknown action: %s", action)
		}
	},
}

func gitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func gitStatus(cfg *config.LoadedConfig) error {
	if !gitAvailable() {
		return fmt.Errorf("git is not installed")
	}

	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = cfg.Paths.BaseDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git status failed: %w\n%s", err, output)
	}

	if len(output) == 0 {
		fmt.Println("✓ No changes")
	} else {
		fmt.Println("Git Status:")
		fmt.Println()
		fmt.Print(string(output))
	}

	return nil
}

func gitCommit(cfg *config.LoadedConfig) error {
	if !gitAvailable() {
		return fmt.Errorf("git is not installed")
	}

	// Add all changes
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = cfg.Paths.BaseDir

	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Commit with auto-generated message
	message := fmt.Sprintf("Auto-commit: %s", time.Now().Format("2006-01-02"))
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = cfg.Paths.BaseDir

	output, err := commitCmd.CombinedOutput()
	if err != nil {
		// Check if there's nothing to commit
		if strings.Contains(string(output), "nothing to commit") {
			fmt.Println("✓ Nothing to commit")
			return nil
		}

		return fmt.Errorf("git commit failed: %w\n%s", err, output)
	}

	fmt.Println("✓ Changes committed")
	fmt.Print(string(output))

	return nil
}

func gitHistory(cfg *config.LoadedConfig) error {
	if !gitAvailable() {
		return fmt.Errorf("git is not installed")
	}

	cmd := exec.Command("git", "log", "--oneline", "-10")
	cmd.Dir = cfg.Paths.BaseDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git log failed: %w\n%s", err, output)
	}

	fmt.Println("Recent commits:")
	fmt.Println()
	fmt.Print(string(output))

	return nil
}

func gitDiff(cfg *config.LoadedConfig) error {
	if !gitAvailable() {
		return fmt.Errorf("git is not installed")
	}

	cmd := exec.Command("git", "diff")
	cmd.Dir = cfg.Paths.BaseDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git diff failed: %w\n%s", err, output)
	}

	if len(output) == 0 {
		fmt.Println("No changes")
	} else {
		fmt.Print(string(output))
	}

	return nil
}
