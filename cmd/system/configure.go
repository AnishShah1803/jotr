package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"

	"github.com/AnishShah1803/jotr/internal/utils"
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

// detectExistingNotesDir searches for existing notes directories.
// Priority: ~/Notes or ~/Obsidian first, then depth-1 and depth-2 subdirs of
// common parent dirs (Documents, Dropbox, workspace, work, home root).
// Shallower matches win — depth-1 beats depth-2.
func detectExistingNotesDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Priority 1: ~/Notes
	notesDir := filepath.Join(homeDir, "Notes")
	if _, err := os.Stat(notesDir); err == nil {
		return notesDir
	}

	// Priority 2: ~/Obsidian
	obsidianDir := filepath.Join(homeDir, "Obsidian")
	if _, err := os.Stat(obsidianDir); err == nil {
		return obsidianDir
	}

	// Priority 3: depth-1 then depth-2 search under common parent dirs.
	// Depth-1 hits win over depth-2 (shallower = more deliberate placement).
	parents := []string{"Documents", "Dropbox", "", "workspace", "work"}

	// isNotesMatch reports whether a directory name looks like a notes vault.
	isNotesMatch := func(name string) bool {
		n := strings.ToLower(name)
		return strings.Contains(n, "notes") || strings.Contains(n, "obsidian")
	}

	// depth-1: ~/parent/<entry>
	for _, parent := range parents {
		basePath := homeDir
		if parent != "" {
			basePath = filepath.Join(homeDir, parent)
		}
		if _, err := os.Stat(basePath); err != nil {
			continue
		}
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if isNotesMatch(entry.Name()) {
				return filepath.Join(basePath, entry.Name())
			}
		}
	}

	// depth-2: ~/parent/<sub>/<entry>  (e.g. ~/Documents/repos/Obsidian)
	for _, parent := range parents {
		basePath := homeDir
		if parent != "" {
			basePath = filepath.Join(homeDir, parent)
		}
		if _, err := os.Stat(basePath); err != nil {
			continue
		}
		subs, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}
		for _, sub := range subs {
			if !sub.IsDir() {
				continue
			}
			subPath := filepath.Join(basePath, sub.Name())
			entries, err := os.ReadDir(subPath)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				if isNotesMatch(entry.Name()) {
					return filepath.Join(subPath, entry.Name())
				}
			}
		}
	}

	return ""
}

// detectDiaryDir searches for common diary/journal directory names within baseDir
func detectDiaryDir(baseDir string) string {
	candidates := []string{"Diary", "Journal", "Daily", "journals", "diary"}
	for _, candidate := range candidates {
		path := filepath.Join(baseDir, candidate)
		if _, err := os.Stat(path); err == nil {
			return candidate
		}
	}
	return "Diary" // default
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
	// Resolve any relative components to absolute path
	expandedBaseDir, err = filepath.Abs(expandedBaseDir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
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
	// Detect existing notes directory
	detectedBaseDir := detectExistingNotesDir()

	// Always use smart defaults flow (works even if no directory detected)
	return runConfigInteractiveSmart(cmd, detectedBaseDir)
}

// runConfigInteractiveSmart presents detected defaults with single confirmation
func runConfigInteractiveSmart(cmd *cobra.Command, detectedBaseDir string) error {
	reader := bufio.NewReader(cmd.InOrStdin())
	cfg := &config.Config{}

	detected := detectedBaseDir != ""

	// Use detected or default values
	baseDir := detectedBaseDir
	if baseDir == "" {
		homeDir, _ := os.UserHomeDir()
		baseDir = filepath.Join(homeDir, "jotr-notes")
	}

	diaryDir := detectDiaryDir(baseDir)
	todoFile := "todo"
	pdpFile := "" // default to none

	fmt.Println("🎯 jotr Configuration Wizard")
	fmt.Println("============================")
	fmt.Println()

	if detected {
		// Show what was found
		fmt.Println("Detected existing notes folder:")
		fmt.Printf("  Base directory: %s\n", baseDir)
		fmt.Printf("  Diary folder:   %s/\n", diaryDir)
		fmt.Printf("  Todo file:      %s.md\n", todoFile)
		if pdpFile != "" {
			fmt.Printf("  PDP file:       %s.md\n", pdpFile)
		} else {
			fmt.Println("  PDP file:       (not configured)")
		}
		fmt.Println()
		fmt.Print("Use these settings? [Y/n/custom]: ")
	} else {
		// Nothing detected — offer the default fallback explicitly
		homeDir, _ := os.UserHomeDir()
		fmt.Println("No existing notes folder detected.")
		fmt.Println()
		fmt.Printf("Default: %s\n", baseDir)
		fmt.Printf("  (jotr will create this folder for you)\n")
		fmt.Println()
		fmt.Printf("Tip: if your notes live somewhere else (e.g. %s/Documents/MyNotes),\n", homeDir)
		fmt.Println("     choose 'custom' to specify the path.")
		fmt.Println()
		fmt.Print("Use default folder, or customise? [Y/custom]: ")
	}

	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	response = strings.TrimSpace(strings.ToLower(response))

	switch response {
	case "", "y", "yes":
		// Use detected defaults
		cfg.Paths.BaseDir = baseDir
		cfg.Paths.DiaryDir = diaryDir
		cfg.Paths.TodoFilePath = todoFile
		cfg.Paths.PDPFilePath = pdpFile
	case "n", "no", "custom":
		// Fall back to full interactive wizard, reusing the same reader to avoid buffer split
		return runConfigInteractiveFull(cmd, reader)
	default:
		return fmt.Errorf("invalid response: %s (expected Y, n, or custom)", response)
	}

	// Validate base directory
	if err := validateBaseDir(cfg.Paths.BaseDir); err != nil {
		return err
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

// runConfigInteractiveFull runs the complete 4-step wizard with path completion
func runConfigInteractiveFull(cmd *cobra.Command, existingReader *bufio.Reader) error {
	var reader *bufio.Reader
	if existingReader != nil {
		reader = existingReader
	} else {
		reader = bufio.NewReader(cmd.InOrStdin())
	}
	cfg := &config.Config{}

	fmt.Println("🎯 jotr Configuration Wizard")
	fmt.Println("============================")
	fmt.Println()

	// Step 1: Base directory (with TAB completion)
	fmt.Println("Step 1: Base Directory")
	fmt.Println("Enter the path to your notes directory (TAB for completion):")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	fmt.Printf("Example: %s/Documents/Notes\n", homeDir)

	baseDir, err := utils.PromptPathWithCompletion("> ", reader)
	if err != nil {
		return fmt.Errorf("failed to read base directory input: %w", err)
	}

	// Expand ~ to home directory
	baseDir, err = expandTildePath(baseDir)
	if err != nil {
		return err
	}
	// Resolve any relative components to absolute path
	baseDir = strings.TrimSpace(baseDir)
	if baseDir != "" {
		baseDir, err = filepath.Abs(baseDir)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}
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

	diaryDir, err := utils.PromptPathWithCompletion("> ", reader)
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

	todoFilePath, err := utils.PromptPathWithCompletion("> ", reader)
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

	pdpFilePath, err := utils.PromptPathWithCompletion("> ", reader)
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
