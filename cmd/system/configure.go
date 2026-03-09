package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
)

var ConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Run configuration wizard",
	Long: `Interactive wizard to set up jotr configuration.

Creates ~/.config/jotr/config.json with your preferences.

Examples:
  jotr configure              # Run configuration wizard
  jotr config                 # Using alias
  jotr cfg                    # Short alias`,
	Aliases: []string{"config", "cfg"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigWizard(cmd)
	},
}

func init() {
	ConfigureCmd.Flags().String("base-dir", "", "Path to your notes directory")
	ConfigureCmd.Flags().String("diary-dir", "", "Name of your diary folder (relative to base)")
	ConfigureCmd.Flags().String("todo-file", "", "Path to your todo file (relative to base, without .md extension)")
	ConfigureCmd.Flags().String("pdp-file", "", "Path to your PDP file (relative to base, without .md extension)")
}

// configPathFromHomeDir returns the config path relative to home directory.
func configPathFromHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "jotr", "config.json"), nil
}

// expandTildePath expands ~ to the user's home directory.
func expandTildePath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, path[1:]), nil
}

// validateBaseDir checks if the base directory is accessible.
// It allows non-existent directories (they can be created), but rejects
// files that aren't directories.
func validateBaseDir(baseDir string) error {
	if baseDir == "" {
		return fmt.Errorf("base directory cannot be empty")
	}
	// Check if the directory exists
	info, err := os.Stat(baseDir)
	if err == nil {
		// Path exists - make sure it's a directory
		if !info.IsDir() {
			return fmt.Errorf("base path is not a directory: %s", baseDir)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		// Path exists but cannot be accessed
		return fmt.Errorf("cannot access base directory: %w", err)
	}
	// Directory doesn't exist - that's fine, it can be created
	return nil
}

// applyDefaults sets default configuration values for fields not explicitly configured.
func applyDefaults(cfg *config.Config) {
	cfg.Format.TaskSection = "Important Things"
	cfg.Format.CaptureSection = "Captured"
	cfg.Format.DailyNoteSections = []string{"Notes", "Conversations/Activities"}
	cfg.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	cfg.Format.DailyNoteDirPattern = "{year}/{month_num}-{month_abbr}"

	cfg.AI.Enabled = true
	cfg.AI.Command = "auggie -p --quiet"

	cfg.Streaks.IncludeWeekends = false

	cfg.Summary.Sources = []string{"todo", "daily_notes"}

	cfg.Frontmatter.Fields = map[string]config.FrontmatterField{
		"status": {
			Type:        "enum",
			Values:      []string{"in-progress", "done", "blocked", "canceled"},
			Required:    false,
			Description: "Note status",
		},
		"priority": {
			Type:        "enum",
			Values:      []string{"P0", "P1", "P2", "P3"},
			Required:    false,
			Description: "Priority level",
		},
		"tags": {
			Type:        "list",
			Required:    false,
			Description: "Tags for categorization",
		},
	}

	cfg.NoteTemplates = make(map[string]interface{})
}

func runConfigWizard(cmd *cobra.Command) error {
	inReplMode := os.Getenv("JOTR_REPL_MODE") == "true"

	// Get flags
	baseDir, _ := cmd.Flags().GetString("base-dir")
	diaryDir, _ := cmd.Flags().GetString("diary-dir")
	todoFile, _ := cmd.Flags().GetString("todo-file")
	pdpFile, _ := cmd.Flags().GetString("pdp-file")

	// Check if any flags provided
	flagsProvided := baseDir != "" || diaryDir != "" || todoFile != "" || pdpFile != ""

	// In REPL mode, flags must be provided
	if inReplMode && !flagsProvided {
		return fmt.Errorf("configure command requires flags in REPL mode, use: configure --base-dir <path> [--diary-dir <name>] [--todo-file <path>] [--pdp-file <path>]")
	}

	// If flags provided, use flag-based configuration
	if flagsProvided {
		return runConfigWithFlags(baseDir, diaryDir, todoFile, pdpFile)
	}

	// Otherwise, use interactive wizard
	return runConfigInteractive(cmd)
}

func runConfigWithFlags(baseDir, diaryDir, todoFile, pdpFile string) error {
	cfg := &config.Config{}

	// Process baseDir
	if baseDir == "" {
		return fmt.Errorf("--base-dir is required")
	}

	// Expand tilde in baseDir and validate
	expandedBaseDir, err := expandTildePath(baseDir)
	if err != nil {
		return err
	}
	if err := validateBaseDir(expandedBaseDir); err != nil {
		return err
	}
	cfg.Paths.BaseDir = expandedBaseDir

	// Process diaryDir
	if diaryDir == "" {
		diaryDir = "Diary"
	}
	cfg.Paths.DiaryDir = diaryDir

	// Process todoFile
	if todoFile == "" {
		todoFile = "todo"
	}
	todoFile = strings.TrimSuffix(todoFile, ".md")
	cfg.Paths.TodoFilePath = todoFile

	// Process pdpFile
	if pdpFile != "" {
		pdpFile = strings.TrimSuffix(pdpFile, ".md")
		cfg.Paths.PDPFilePath = pdpFile
	}

	// Apply default configuration values
	applyDefaults(cfg)

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath, err := configPathFromHomeDir()
	if err != nil {
		return err
	}
	fmt.Printf("✓ Configuration saved to: %s\n", configPath)

	return nil
}

func runConfigInteractive(cmd *cobra.Command) error {
	reader := bufio.NewReader(cmd.InOrStdin())
	cfg := &config.Config{}

	fmt.Println("🎯 jotr Configuration Wizard")
	fmt.Println("============================")
	fmt.Println()

	// Step 1: Base directory
	fmt.Println("Step 1: Base Directory")
	fmt.Println("Enter the path to your notes directory:")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	fmt.Printf("Example: %s/Documents/Notes\n", homeDir)
	fmt.Print("> ")

	baseDir, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read base directory input: %w", err)
	}

	baseDir = strings.TrimSpace(baseDir)

	// Expand ~ to home directory
	baseDir, err = expandTildePath(baseDir)
	if err != nil {
		return err
	}

	// Validate the directory path
	if err := validateBaseDir(baseDir); err != nil {
		return err
	}

	cfg.Paths.BaseDir = baseDir
	fmt.Printf("✓ Base directory: %s\n\n", baseDir)

	// Step 2: Diary directory
	fmt.Println("Step 2: Diary Directory")
	fmt.Println("Enter the name of your diary folder (relative to base):")
	fmt.Println("Example: Diary, Journal, Daily")
	fmt.Print("> ")

	diaryDir, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read diary directory input: %w", err)
	}

	diaryDir = strings.TrimSpace(diaryDir)
	if diaryDir == "" {
		diaryDir = "Diary"
	}

	cfg.Paths.DiaryDir = diaryDir
	fmt.Printf("✓ Diary directory: %s\n\n", diaryDir)

	// Step 3: Todo file path
	fmt.Println("Step 3: Todo File Path")
	fmt.Println("Enter the path to your todo file (relative to base, without .md extension):")
	fmt.Println("Example: todo, Work/tasks, TODO")
	fmt.Print("> ")

	todoFilePath, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read todo file path input: %w", err)
	}

	todoFilePath = strings.TrimSpace(todoFilePath)
	if todoFilePath == "" {
		todoFilePath = "todo"
	}

	todoFilePath = strings.TrimSuffix(todoFilePath, ".md")
	cfg.Paths.TodoFilePath = todoFilePath
	fmt.Printf("✓ Todo file: %s.md\n", todoFilePath)

	todoBasename := filepath.Base(todoFilePath)
	stateFile := fmt.Sprintf(".%s_state.json", todoBasename)
	fmt.Printf("✓ State file will be: %s (auto-generated in same directory)\n\n", stateFile)

	// Step 4: PDP file (optional)
	fmt.Println("Step 4: PDP File (Optional)")
	fmt.Println("Enter the path to your PDP file (relative to base, without .md extension):")
	fmt.Println("Press Enter to skip")
	fmt.Print("> ")

	pdpFilePath, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read PDP file path input: %w", err)
	}

	pdpFilePath = strings.TrimSpace(pdpFilePath)
	if pdpFilePath != "" {
		pdpFilePath = strings.TrimSuffix(pdpFilePath, ".md")
		cfg.Paths.PDPFilePath = pdpFilePath
		fmt.Printf("✓ PDP file: %s.md\n\n", pdpFilePath)
	} else {
		fmt.Println("✓ PDP file: (none)")
	}

	// Apply default configuration values
	applyDefaults(cfg)

	fmt.Println("Saving configuration...")

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configPath, err := configPathFromHomeDir()
	if err != nil {
		return err
	}
	fmt.Printf("✓ Configuration saved to: %s\n\n", configPath)

	fmt.Println("🎉 Configuration complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("  jotr daily     # Create today's note")
	fmt.Println("  jotr --help    # See all commands")

	return nil
}
