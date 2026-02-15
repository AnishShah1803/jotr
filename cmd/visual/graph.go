package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var (
	graphOutput string
	graphFormat string
)

// GraphCmd is the command for generating relationship graphs.
var GraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate relationship graph",
	Long: `Generate a visual graph of note relationships.
	
Requires graphviz (dot) to be installed.

Examples:
  jotr graph                        # Generate graph.png
  jotr graph --output notes.png     # Custom output
  jotr graph --format svg           # SVG format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		return generateGraph(cmd.Context(), cfg)
	},
}

func generateGraph(ctx context.Context, cfg *config.LoadedConfig) error {
	// Check if graphviz is installed
	if _, err := exec.LookPath("dot"); err != nil {
		return fmt.Errorf("graphviz (dot) is not installed\nInstall with: brew install graphviz")
	}

	// Find all notes
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	fmt.Printf("Analyzing %d notes...\n", len(allNotes))

	// Build graph data
	type Link struct {
		from string
		to   string
	}

	links := []Link{}
	linkRe := regexp.MustCompile(`\[\[([^\]]+)\]\]`)

	for _, notePath := range allNotes {
		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		fromNote := filepath.Base(notePath)
		fromNote = strings.TrimSuffix(fromNote, ".md")

		matches := linkRe.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) > 1 {
				toNote := match[1]
				links = append(links, Link{from: fromNote, to: toNote})
			}
		}
	}

	if len(links) == 0 {
		return fmt.Errorf("no links found between notes")
	}

	fmt.Printf("Found %d links\n", len(links))

	// Generate DOT file
	dotContent := "digraph notes {\n"
	dotContent += "  rankdir=LR;\n"
	dotContent += "  node [shape=box, style=rounded];\n"
	dotContent += "  edge [color=gray];\n\n"

	// Add nodes
	nodeSet := make(map[string]bool)
	for _, link := range links {
		nodeSet[link.from] = true
		nodeSet[link.to] = true
	}

	for node := range nodeSet {
		// Sanitize node ID for DOT syntax
		nodeID := sanitizeNodeID(node)

		// Escape quotes and limit length for label
		label := node
		if len(label) > 30 {
			label = label[:27] + "..."
		}

		label = strings.ReplaceAll(label, "\"", "\\\"")

		dotContent += fmt.Sprintf("  %s [label=\"%s\"];\n", nodeID, label)
	}

	dotContent += "\n"

	// Add edges
	for _, link := range links {
		fromID := sanitizeNodeID(link.from)
		toID := sanitizeNodeID(link.to)
		dotContent += fmt.Sprintf("  %s -> %s;\n", fromID, toID)
	}

	dotContent += "}\n"

	// Write DOT file
	dotFile := filepath.Join(cfg.Paths.BaseDir, ".graph.dot")
	if err := os.WriteFile(dotFile, []byte(dotContent), constants.FilePerm0644); err != nil {
		return fmt.Errorf("failed to write DOT file: %w", err)
	}

	// Generate graph image
	outputPath := graphOutput
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(cfg.Paths.BaseDir, outputPath)
	}

	cmd := exec.Command("dot", "-T"+graphFormat, dotFile, "-o", outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to generate graph: %w\n%s", err, output)
	}

	// Clean up DOT file
	if err := os.Remove(dotFile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove temporary DOT file: %v\n", err)
	}

	fmt.Printf("✓ Graph generated: %s\n", outputPath)
	fmt.Printf("  %d nodes, %d links\n", len(nodeSet), len(links))

	// Try to open the graph
	openGraph(outputPath)

	return nil
}

func openGraph(path string) {
	// Try to open the graph with default viewer
	var cmd *exec.Cmd

	switch {
	case utils.FileExists("/usr/bin/open"): // macOS
		cmd = exec.Command("open", path)
	case utils.FileExists("/usr/bin/xdg-open"): // Linux
		cmd = exec.Command("xdg-open", path)
	default:
		fmt.Println("  (Open the file manually to view)")
		return
	}

	if err := cmd.Start(); err == nil {
		fmt.Println("  Opening graph...")
	}
}

// sanitizeNodeID converts a node name to a valid DOT identifier.
func sanitizeNodeID(name string) string {
	// Replace problematic characters with underscores
	// DOT doesn't like: | < > { } [ ] " \ and some others
	replacer := strings.NewReplacer(
		"|", "_",
		"<", "_",
		">", "_",
		"{", "_",
		"}", "_",
		"[", "_",
		"]", "_",
		"\"", "_",
		"\\", "_",
		" ", "_",
		"-", "_",
		".", "_",
		",", "_",
		":", "_",
		";", "_",
		"(", "_",
		")", "_",
		"/", "_",
		"&", "_",
		"#", "_",
		"@", "_",
		"!", "_",
		"?", "_",
		"*", "_",
		"+", "_",
		"=", "_",
		"%", "_",
	)

	sanitized := replacer.Replace(name)

	// Ensure it starts with a letter or underscore
	if len(sanitized) > 0 && !isLetter(sanitized[0]) && sanitized[0] != '_' {
		sanitized = "n_" + sanitized
	}

	// If empty, use a default
	if sanitized == "" {
		sanitized = "node"
	}

	return sanitized
}

// isLetter checks if a byte is a letter.
func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
