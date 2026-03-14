package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/notes"
)

var PropertiesCmd = &cobra.Command{
	Use:   "properties [action] [note-name]",
	Short: "Manage typed note properties",
	Long: `View or edit typed properties in note frontmatter with type validation.

Supported Types:
  text       Plain text string
  list       Comma-separated values or YAML array
  number     Integer or floating-point number
  checkbox   Boolean (true/false, yes/no, x/-)
  date       Date in YYYY-MM-DD format
  datetime   ISO 8601 datetime format

Actions:
  list [note]              Show all properties (default)
  get [note] [key]         Get a specific property value
  set [note] key=val       Set a property (auto-detect type)
  set [note] key:type=val  Set a typed property
  remove [note] [key]      Remove a property
  stats                    Show vault-wide property statistics
  stats [property]         Show statistics for a specific property
  
Flags:
  --type string           Specify property type when setting
  --counts                Sort by usage count (for stats)
  
Examples:
  jotr properties MyNote
  jotr properties get MyNote status
  jotr properties set MyNote status=active
  jotr properties set MyNote priority:number=5
  jotr properties stats
  jotr properties stats category --counts`,
	Aliases: []string{"prop", "props"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("action required: list, get, set, remove, or stats")
		}

		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		action := args[0]
		switch action {
		case "list":
			if len(args) < 2 {
				return fmt.Errorf("note name required")
			}
			return showProperties(cmd.Context(), cfg, args[1])
		case "get":
			if len(args) < 3 {
				return fmt.Errorf("usage: properties get <note> <key>")
			}
			return getProperty(cmd.Context(), cfg, args[1], args[2])
		case "set":
			if len(args) < 3 {
				return fmt.Errorf("usage: properties set <note> key=value")
			}
			propType, _ := cmd.Flags().GetString("type")
			return setProperty(cmd.Context(), cfg, args[1], args[2], propType)
		case "remove":
			if len(args) < 3 {
				return fmt.Errorf("usage: properties remove <note> <key>")
			}
			return removeProperty(cmd.Context(), cfg, args[1], args[2])
		case "stats":
			property := ""
			if len(args) > 1 {
				property = args[1]
			}
			counts, _ := cmd.Flags().GetBool("counts")
			return showPropertyStats(cmd.Context(), cfg, property, counts)
		default:
			return showProperties(cmd.Context(), cfg, action)
		}
	},
}

func init() {
	PropertiesCmd.Flags().String("type", "", "property type (text, list, number, checkbox, date, datetime)")
	PropertiesCmd.Flags().Bool("counts", false, "sort by usage count")
}

func parsePropertyType(val string) (string, string) {
	parts := strings.SplitN(val, ":", 2)
	if len(parts) == 2 && isValidType(parts[0]) {
		return parts[0], parts[1]
	}
	return detectType(parts[len(parts)-1]), parts[len(parts)-1]
}

func detectType(val string) string {
	val = strings.TrimSpace(val)
	if val == "" {
		return "text"
	}
	if strings.EqualFold(val, "true") || strings.EqualFold(val, "false") ||
		strings.EqualFold(val, "yes") || strings.EqualFold(val, "no") ||
		val == "x" || val == "-" {
		return "checkbox"
	}
	if _, err := strconv.ParseInt(val, 10, 64); err == nil {
		return "number"
	}
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return "number"
	}
	if strings.Contains(val, ",") {
		return "list"
	}
	if _, err := time.Parse("2006-01-02", val); err == nil {
		return "date"
	}
	if _, err := time.Parse(time.RFC3339, val); err == nil {
		return "datetime"
	}
	return "text"
}

func isValidType(t string) bool {
	types := map[string]bool{
		"text": true, "list": true, "number": true,
		"checkbox": true, "date": true, "datetime": true,
	}
	return types[strings.ToLower(t)]
}

func formatPropertyValue(val string, propType string) string {
	if propType == "" {
		propType = detectType(val)
	}
	switch strings.ToLower(propType) {
	case "list":
		items := strings.Split(val, ",")
		for i, item := range items {
			items[i] = strings.TrimSpace(item)
		}
		return "[" + strings.Join(items, ", ") + "]"
	default:
		return quoteFrontmatterValue(val)
	}
}

func showProperties(ctx context.Context, cfg *config.LoadedConfig, noteName string) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	var targetNote string
	for _, note := range allNotes {
		if strings.Contains(strings.ToLower(filepath.Base(note)), strings.ToLower(noteName)) {
			targetNote = note
			break
		}
	}

	if targetNote == "" {
		return fmt.Errorf("note not found: %s", noteName)
	}

	content, err := os.ReadFile(targetNote)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	if len(lines) < 3 || lines[0] != "---" {
		fmt.Printf("No properties in %s\n", filepath.Base(targetNote))
		return nil
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		fmt.Printf("Invalid frontmatter in %s\n", filepath.Base(targetNote))
		return nil
	}

	fmt.Printf("Properties in %s:\n\n", filepath.Base(targetNote))

	for i := 1; i < endIdx; i++ {
		line := lines[i]
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			propType := detectType(value)
			fmt.Printf("  %s (%s): %s\n", key, propType, value)
		}
	}

	return nil
}

func setProperty(ctx context.Context, cfg *config.LoadedConfig, noteName string, propDef string, typeOverride string) error {
	key, value, propType := parsePropDef(propDef, typeOverride)

	if key == "" {
		return fmt.Errorf("invalid format, use: key=value or key:type=value")
	}

	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	var targetNote string
	for _, note := range allNotes {
		if strings.Contains(strings.ToLower(filepath.Base(note)), strings.ToLower(noteName)) {
			targetNote = note
			break
		}
	}

	if targetNote == "" {
		return fmt.Errorf("note not found: %s", noteName)
	}

	content, err := os.ReadFile(targetNote)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	newLines := []string{}
	if len(lines) > 0 && lines[0] == "---" {
		newLines = append(newLines, "---")
		updated := false

		for i := 1; i < len(lines); i++ {
			if lines[i] == "---" {
				if !updated {
					newLines = append(newLines, fmt.Sprintf("%s: %s", key, formatPropertyValue(value, propType)))
				}
				newLines = append(newLines, lines[i:]...)
				break
			}

			if strings.HasPrefix(lines[i], key+":") {
				newLines = append(newLines, fmt.Sprintf("%s: %s", key, formatPropertyValue(value, propType)))
				updated = true
			} else {
				newLines = append(newLines, lines[i])
			}
		}
	} else {
		newLines = append(newLines, "---")
		newLines = append(newLines, fmt.Sprintf("%s: %s", key, formatPropertyValue(value, propType)))
		newLines = append(newLines, "---")
		newLines = append(newLines, "")
		newLines = append(newLines, lines...)
	}

	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(targetNote, []byte(newContent), constants.FilePerm0644); err != nil {
		return err
	}

	fmt.Printf("✓ Set %s (%s) in %s\n", key, propType, filepath.Base(targetNote))
	return nil
}

func getProperty(ctx context.Context, cfg *config.LoadedConfig, noteName string, key string) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	var targetNote string
	for _, note := range allNotes {
		if strings.Contains(strings.ToLower(filepath.Base(note)), strings.ToLower(noteName)) {
			targetNote = note
			break
		}
	}

	if targetNote == "" {
		return fmt.Errorf("note not found: %s", noteName)
	}

	content, err := os.ReadFile(targetNote)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	if len(lines) < 3 || lines[0] != "---" {
		return fmt.Errorf("no properties in %s", filepath.Base(targetNote))
	}

	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			break
		}
		if strings.HasPrefix(lines[i], key+":") {
			parts := strings.SplitN(lines[i], ":", 2)
			fmt.Printf("%s\n", strings.TrimSpace(parts[1]))
			return nil
		}
	}

	return fmt.Errorf("property not found: %s", key)
}

func removeProperty(ctx context.Context, cfg *config.LoadedConfig, noteName string, key string) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	var targetNote string
	for _, note := range allNotes {
		if strings.Contains(strings.ToLower(filepath.Base(note)), strings.ToLower(noteName)) {
			targetNote = note
			break
		}
	}

	if targetNote == "" {
		return fmt.Errorf("note not found: %s", noteName)
	}

	content, err := os.ReadFile(targetNote)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	if len(lines) > 0 && lines[0] == "---" {
		newLines = append(newLines, lines[0])
		inFrontmatter := true
		for _, line := range lines[1:] {
			if inFrontmatter && line == "---" {
				inFrontmatter = false
				newLines = append(newLines, line)
			} else if inFrontmatter && strings.HasPrefix(line, key+":") {
				// drop this line
			} else {
				newLines = append(newLines, line)
			}
		}
	} else {
		newLines = lines
	}

	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(targetNote, []byte(newContent), constants.FilePerm0644); err != nil {
		return err
	}

	fmt.Printf("✓ Removed %s from %s\n", key, filepath.Base(targetNote))
	return nil
}

func showPropertyStats(ctx context.Context, cfg *config.LoadedConfig, property string, sortByCount bool) error {
	allNotes, err := notes.FindNotes(ctx, cfg.Paths.BaseDir)
	if err != nil {
		return err
	}

	propStats := make(map[string]map[string]int)
	propTypes := make(map[string]string)

	for _, notePath := range allNotes {
		content, err := os.ReadFile(notePath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		if len(lines) < 3 || lines[0] != "---" {
			continue
		}

		for i := 1; i < len(lines); i++ {
			if lines[i] == "---" {
				break
			}
			parts := strings.SplitN(lines[i], ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				propType := detectType(val)

				if property != "" && key != property {
					continue
				}

				if _, ok := propStats[key]; !ok {
					propStats[key] = make(map[string]int)
					propTypes[key] = propType
				}
				propStats[key][val]++
			}
		}
	}

	if len(propStats) == 0 {
		fmt.Println("No properties found in vault")
		return nil
	}

	if property != "" && propStats[property] != nil {
		fmt.Printf("Property '%s' statistics:\n\n", property)
		printPropertyValues(propStats[property], sortByCount)
		return nil
	}

	if property == "" {
		fmt.Println("Vault-wide property statistics:")
		for prop, vals := range propStats {
			count := 0
			for _, c := range vals {
				count += c
			}
			fmt.Printf("%s (%d uses)\n", prop, count)
		}
		return nil
	}

	fmt.Printf("Property not found: %s\n", property)
	return nil
}

func printPropertyValues(vals map[string]int, sortByCount bool) {
	type kv struct {
		key   string
		count int
	}
	var sorted []kv
	for k, v := range vals {
		sorted = append(sorted, kv{k, v})
	}

	if sortByCount {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].count > sorted[j].count
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].key < sorted[j].key
		})
	}

	for _, kv := range sorted {
		fmt.Printf("  %s: %d\n", kv.key, kv.count)
	}
}

func parsePropDef(propDef string, typeOverride string) (key string, value string, propType string) {
	if strings.Contains(propDef, ":") && strings.Contains(propDef, "=") {
		idxColon := strings.Index(propDef, ":")
		idxEq := strings.Index(propDef, "=")
		if idxColon < idxEq {
			key = strings.TrimSpace(propDef[:idxColon])
			propType = strings.TrimSpace(propDef[idxColon+1 : idxEq])
			value = strings.TrimSpace(propDef[idxEq+1:])
			if !isValidType(propType) {
				propType = detectType(value)
			}
			return
		}
	}

	parts := strings.SplitN(propDef, "=", 2)
	if len(parts) == 2 {
		key = strings.TrimSpace(parts[0])
		value = strings.TrimSpace(parts[1])
		if typeOverride != "" && isValidType(typeOverride) {
			propType = typeOverride
		} else {
			propType = detectType(value)
		}
		return
	}

	return "", "", ""
}
