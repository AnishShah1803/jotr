package templates

import (
	"strings"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
)

func TestSubstituteVariables(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		vars     map[string]string
		builtins map[string]string
		want     string
	}{
		{
			name:     "single variable",
			content:  "Hello {$name}!",
			vars:     map[string]string{"name": "World"},
			builtins: map[string]string{},
			want:     "Hello World!",
		},
		{
			name:     "multiple variables",
			content:  "{$greeting} {$name}, time is {$time}",
			vars:     map[string]string{"greeting": "Hello", "name": "Alice"},
			builtins: map[string]string{"{$time}": "10:30"},
			want:     "Hello Alice, time is 10:30",
		},
		{
			name:     "built-in variables",
			content:  "Date: {$date}, Base: {$base_dir}",
			vars:     map[string]string{},
			builtins: map[string]string{"{$date}": "2024-01-15", "{$base_dir}": "/home/user/Notes"},
			want:     "Date: 2024-01-15, Base: /home/user/Notes",
		},
		{
			name:     "missing variables",
			content:  "Hello {$name} and {$other}!",
			vars:     map[string]string{"name": "Bob"},
			builtins: map[string]string{},
			want:     "Hello Bob and {$other}!",
		},
		{
			name:     "repeated variable",
			content:  "{$x} + {$x} = {$sum}",
			vars:     map[string]string{"x": "5", "sum": "10"},
			builtins: map[string]string{},
			want:     "5 + 5 = 10",
		},
		{
			name:     "no substitutions needed",
			content:  "Just plain text",
			vars:     map[string]string{},
			builtins: map[string]string{},
			want:     "Just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SubstituteVariables(tt.content, tt.vars, tt.builtins)
			if got != tt.want {
				t.Errorf("SubstituteVariables() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSubstitutePrompts(t *testing.T) {
	tests := []struct {
		name    string
		content string
		values  []string
		want    string
	}{
		{
			name:    "single prompt",
			content: "<prompt>Question?</prompt>",
			values:  []string{"Answer"},
			want:    "Answer",
		},
		{
			name:    "multiple prompts",
			content: "<prompt>First?</prompt> and <prompt>Second?</prompt>",
			values:  []string{"A1", "A2"},
			want:    "A1 and A2",
		},
		{
			name:    "prompt with surrounding text",
			content: "Name: <prompt>What's your name?</prompt>",
			values:  []string{"Alice"},
			want:    "Name: Alice",
		},
		{
			name:    "multiline content",
			content: "# Header\n<prompt>Question?</prompt>\n## Footer",
			values:  []string{"Response"},
			want:    "# Header\nResponse\n## Footer",
		},
		{
			name:    "more values than prompts",
			content: "<prompt>Q?</prompt>",
			values:  []string{"A1", "A2"},
			want:    "A1",
		},
		{
			name:    "fewer values than prompts",
			content: "<prompt>Q1?</prompt> <prompt>Q2?</prompt>",
			values:  []string{"A1"},
			want:    "A1 <prompt>Q2?</prompt>",
		},
		{
			name:    "no prompts",
			content: "Just text",
			values:  []string{},
			want:    "Just text",
		},
		{
			name:    "complex prompt with special chars",
			content: "<prompt>What's the (main) point?</prompt>",
			values:  []string{"To learn"},
			want:    "To learn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SubstitutePrompts(tt.content, tt.values)
			if got != tt.want {
				t.Errorf("SubstitutePrompts() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	cfg := &config.Config{}
	cfg.Paths.BaseDir = "/home/user/Notes"

	tmpl := &Template{
		Name: "test",
		Content: `<!-- Path: ~/Notes/{$date}-{$topic}.md -->

topic = <prompt>What is the topic?</prompt>

# Meeting: {$topic}

**Date:** {$date}

## Notes
<prompt>What are the key points?</prompt>`,
		BuiltIns: make(map[string]string),
	}

	userVars := map[string]string{
		"topic": "Project Planning",
	}
	promptValues := []string{
		"Project Planning",
		"Discussed timeline and milestones",
	}

	result := RenderTemplate(tmpl, userVars, promptValues, cfg)

	if !strings.Contains(result, "# Meeting: Project Planning") {
		t.Errorf("RenderTemplate() missing substituted topic, got: %s", result)
	}

	if !strings.Contains(result, "**Date:**") {
		t.Errorf("RenderTemplate() missing date line, got: %s", result)
	}

	if !strings.Contains(result, "Discussed timeline and milestones") {
		t.Errorf("RenderTemplate() missing prompt value, got: %s", result)
	}

	if strings.Contains(result, "<!-- Path:") {
		t.Errorf("RenderTemplate() should remove path comment, got: %s", result)
	}

	if strings.Contains(result, "<prompt>") {
		t.Errorf("RenderTemplate() should remove prompts, got: %s", result)
	}
}

func TestResolveBuiltIns(t *testing.T) {
	cfg := &config.Config{}
	cfg.Paths.BaseDir = "/home/user/Notes"

	tmpl := &Template{
		Name:     "test",
		BuiltIns: make(map[string]string),
	}

	ResolveBuiltIns(tmpl, cfg)

	now := time.Now()

	if tmpl.BuiltIns["{$date}"] != now.Format("2006-01-02") {
		t.Errorf("ResolveBuiltIns() date = %v, want %v", tmpl.BuiltIns["{$date}"], now.Format("2006-01-02"))
	}

	if tmpl.BuiltIns["{$datetime}"] != now.Format("2006-01-02 15:04") {
		t.Errorf("ResolveBuiltIns() datetime = %v, want %v", tmpl.BuiltIns["{$datetime}"], now.Format("2006-01-02 15:04"))
	}

	if tmpl.BuiltIns["{$base_dir}"] != "/home/user/Notes" {
		t.Errorf("ResolveBuiltIns() base_dir = %v, want %v", tmpl.BuiltIns["{$base_dir}"], "/home/user/Notes")
	}

	if tmpl.BuiltIns["{$weekday}"] != now.Format("Monday") {
		t.Errorf("ResolveBuiltIns() weekday = %v, want %v", tmpl.BuiltIns["{$weekday}"], now.Format("Monday"))
	}

	if tmpl.BuiltIns["{$time}"] != now.Format("15:04") {
		t.Errorf("ResolveBuiltIns() time = %v, want %v", tmpl.BuiltIns["{$time}"], now.Format("15:04"))
	}
}
