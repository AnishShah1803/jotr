package templates

import (
	"strings"
	"testing"
)

func TestParseFilename(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		wantPriority int
		wantCategory string
		wantName     string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid filename",
			filename:     "1-1-meeting.md",
			wantPriority: 1,
			wantCategory: "1",
			wantName:     "meeting",
			wantErr:      false,
		},
		{
			name:         "valid filename with multi-word name",
			filename:     "10-2-journal-entry.md",
			wantPriority: 10,
			wantCategory: "2",
			wantName:     "journal-entry",
			wantErr:      false,
		},
		{
			name:        "invalid - no extension",
			filename:    "meeting",
			wantErr:     true,
			errContains: "invalid template filename format",
		},
		{
			name:        "invalid - missing priority",
			filename:    "a-b-c.md",
			wantErr:     true,
			errContains: "invalid priority",
		},
		{
			name:        "invalid - too few parts",
			filename:    "1-name.md",
			wantErr:     true,
			errContains: "invalid template filename format",
		},
		{
			name:         "valid - empty category",
			filename:     "1--name.md",
			wantPriority: 1,
			wantCategory: "",
			wantName:     "name",
			wantErr:      false,
		},
		{
			name:         "valid - complex name with many dashes",
			filename:     "5-3-my-complex-template-name.md",
			wantPriority: 5,
			wantCategory: "3",
			wantName:     "my-complex-template-name",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPriority, gotCategory, gotName, err := ParseFilename(tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFilename() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseFilename() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFilename() unexpected error: %v", err)
				return
			}

			if gotPriority != tt.wantPriority {
				t.Errorf("ParseFilename() priority = %v, want %v", gotPriority, tt.wantPriority)
			}
			if gotCategory != tt.wantCategory {
				t.Errorf("ParseFilename() category = %v, want %v", gotCategory, tt.wantCategory)
			}
			if gotName != tt.wantName {
				t.Errorf("ParseFilename() name = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestParsePathComment(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantPath    string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid path comment",
			content: `<!-- Path: ~/Notes/Meetings/{$date}-{$topic}.md -->

# Meeting Notes`,
			wantPath: "~/Notes/Meetings/{$date}-{$topic}.md",
			wantErr:  false,
		},
		{
			name: "no path comment",
			content: `# Meeting Notes

Some content`,
			wantPath: "",
			wantErr:  false,
		},
		{
			name: "path comment with spaces",
			content: `<!-- Path: ~/Notes/My File.md -->
Content`,
			wantPath: "~/Notes/My File.md",
			wantErr:  false,
		},
		{
			name: "empty path comment",
			content: `<!-- Path: -->
Content`,
			wantErr:     true,
			errContains: "empty path",
		},
		{
			name: "path comment on later line",
			content: `# Some header

<!-- Path: ~/Notes/test.md -->
More content`,
			wantPath: "~/Notes/test.md",
			wantErr:  false,
		},
		{
			name: "multiple comments - finds first",
			content: `<!-- Path: ~/Notes/first.md -->
<!-- Path: ~/Notes/second.md -->
Content`,
			wantPath: "~/Notes/first.md",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, err := ParsePathComment(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePathComment() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParsePathComment() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParsePathComment() unexpected error: %v", err)
				return
			}

			if gotPath != tt.wantPath {
				t.Errorf("ParsePathComment() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestParseVariables(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantVars    []Variable
		wantErr     bool
		errContains string
	}{
		{
			name: "single variable",
			content: `topic = <prompt>What is the topic?</prompt>
# Meeting`,
			wantVars: []Variable{
				{Name: "topic", Prompt: "What is the topic?"},
			},
			wantErr: false,
		},
		{
			name: "multiple variables",
			content: `topic = <prompt>What is the topic?</prompt>
location = <prompt>Where is it happening?</prompt>
time = <prompt>What time?</prompt>
# Meeting`,
			wantVars: []Variable{
				{Name: "topic", Prompt: "What is the topic?"},
				{Name: "location", Prompt: "Where is it happening?"},
				{Name: "time", Prompt: "What time?"},
			},
			wantErr: false,
		},
		{
			name: "duplicate variable",
			content: `topic = <prompt>First?</prompt>
topic = <prompt>Second?</prompt>
# Meeting`,
			wantErr:     true,
			errContains: "variable defined multiple times",
		},
		{
			name: "no variables",
			content: `# Meeting
Some content`,
			wantVars: []Variable(nil),
			wantErr:  false,
		},
		{
			name: "variable with spaces around equals",
			content: `topic  =  <prompt>Question?</prompt>
Content`,
			wantVars: []Variable{
				{Name: "topic", Prompt: "Question?"},
			},
			wantErr: false,
		},
		{
			name: "variable on non-first line",
			content: `# Meeting

Some context
topic = <prompt>What?</prompt>
More content`,
			wantVars: []Variable{
				{Name: "topic", Prompt: "What?"},
			},
			wantErr: false,
		},
		{
			name: "variable with special characters in prompt",
			content: `title = <prompt>What's the title? (use spaces)</prompt>
Content`,
			wantVars: []Variable{
				{Name: "title", Prompt: "What's the title? (use spaces)"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVars, err := ParseVariables(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseVariables() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseVariables() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseVariables() unexpected error: %v", err)
				return
			}

			if len(gotVars) != len(tt.wantVars) {
				t.Errorf("ParseVariables() got %d variables, want %d", len(gotVars), len(tt.wantVars))
				return
			}

			for i, got := range gotVars {
				want := tt.wantVars[i]
				if got.Name != want.Name || got.Prompt != want.Prompt {
					t.Errorf("ParseVariables()[%d] = %+v, want %+v", i, got, want)
				}
			}
		})
	}
}

func TestParsePrompts(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantLen int
		wantQ   []string
	}{
		{
			name: "single prompt",
			content: `<prompt>What did you learn?</prompt>
Content`,
			wantLen: 1,
			wantQ:   []string{"What did you learn?"},
		},
		{
			name: "multiple prompts",
			content: `<prompt>First question?</prompt>
Some text
<prompt>Second question?</prompt>
More text
<prompt>Third question?</prompt>`,
			wantLen: 3,
			wantQ:   []string{"First question?", "Second question?", "Third question?"},
		},
		{
			name: "no prompts",
			content: `# Meeting
Some content`,
			wantLen: 0,
			wantQ:   []string(nil),
		},
		{
			name: "mixed with variables",
			content: `topic = <prompt>Topic?</prompt>
# Meeting
<prompt>Attendees?</prompt>
<prompt>Agenda?</prompt>`,
			wantLen: 3,
			wantQ:   []string{"Topic?", "Attendees?", "Agenda?"},
		},
		{
			name: "prompt with special characters",
			content: `<prompt>What's the (main) point? - maybe use "quotes"</prompt>
Content`,
			wantLen: 1,
			wantQ:   []string{`What's the (main) point? - maybe use "quotes"`},
		},
		{
			name: "nested prompts in lists",
			content: `- Item 1 <prompt>Detail 1?</prompt>
- Item 2 <prompt>Detail 2?</prompt>`,
			wantLen: 2,
			wantQ:   []string{"Detail 1?", "Detail 2?"},
		},
		{
			name: "multiline content with prompts",
			content: `# Header

Some paragraph text.

## Section
<prompt>What's next?</prompt>

More content.
<prompt>And then?</prompt>`,
			wantLen: 2,
			wantQ:   []string{"What's next?", "And then?"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrompts := ParsePrompts(tt.content)

			if len(gotPrompts) != tt.wantLen {
				t.Errorf("ParsePrompts() got %d prompts, want %d", len(gotPrompts), tt.wantLen)
				return
			}

			if tt.wantQ != nil {
				for i, got := range gotPrompts {
					if got.Question != tt.wantQ[i] {
						t.Errorf("ParsePrompts()[%d] question = %q, want %q", i, got.Question, tt.wantQ[i])
					}
					if got.Variable != nil {
						t.Errorf("ParsePrompts()[%d] variable should be nil (freeform), got %v", i, *got.Variable)
					}
				}
			}
		})
	}
}

func TestParseTemplate(t *testing.T) {
	tests := []struct {
		name        string
		filepath    string
		content     string
		want        *Template
		wantErr     bool
		errContains string
	}{
		{
			name:     "complete template",
			filepath: "/templates/1-1-meeting.md",
			content: `<!-- Path: ~/Notes/Meetings/{$date}-{$topic}.md -->

topic = <prompt>What is the meeting topic?</prompt>

# Meeting Notes: {$topic}

**Date:** {$date}
**Time:** {$time}

## Attendees
<prompt>Who attended this meeting?</prompt>

## Agenda
<prompt>What are the agenda items?</prompt>`,
			want: &Template{
				Filename:   "1-1-meeting.md",
				Priority:   1,
				Category:   "1",
				Name:       "meeting",
				TargetPath: "~/Notes/Meetings/{$date}-{$topic}.md",
				Variables: []Variable{
					{Name: "topic", Prompt: "What is the meeting topic?"},
				},
				Prompts: []Prompt{
					{Question: "What is the meeting topic?", Variable: nil},
					{Question: "Who attended this meeting?", Variable: nil},
					{Question: "What are the agenda items?", Variable: nil},
				},
				BuiltIns: make(map[string]string),
			},
			wantErr: false,
		},
		{
			name:     "template without path",
			filepath: "/templates/2-2-journal.md",
			content: `# Journal Entry

<prompt>What's on your mind?</prompt>`,
			want: &Template{
				Filename:   "2-2-journal.md",
				Priority:   2,
				Category:   "2",
				Name:       "journal",
				TargetPath: "",
				Variables:  []Variable(nil),
				Prompts: []Prompt{
					{Question: "What's on your mind?", Variable: nil},
				},
				BuiltIns: make(map[string]string),
			},
			wantErr: false,
		},
		{
			name:     "invalid filename",
			filepath: "/templates/meeting.md",
			content:  `# Meeting`,
			wantErr:  true,
		},
		{
			name:     "duplicate variables",
			filepath: "/templates/1-1-test.md",
			content: `topic = <prompt>First?</prompt>
topic = <prompt>Second?</prompt>`,
			wantErr:     true,
			errContains: "template 1-1-test.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTemplate(tt.filepath, tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTemplate() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseTemplate() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTemplate() unexpected error: %v", err)
				return
			}

			if got.Filename != tt.want.Filename {
				t.Errorf("ParseTemplate() Filename = %v, want %v", got.Filename, tt.want.Filename)
			}
			if got.Priority != tt.want.Priority {
				t.Errorf("ParseTemplate() Priority = %v, want %v", got.Priority, tt.want.Priority)
			}
			if got.Category != tt.want.Category {
				t.Errorf("ParseTemplate() Category = %v, want %v", got.Category, tt.want.Category)
			}
			if got.Name != tt.want.Name {
				t.Errorf("ParseTemplate() Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.TargetPath != tt.want.TargetPath {
				t.Errorf("ParseTemplate() TargetPath = %v, want %v", got.TargetPath, tt.want.TargetPath)
			}
			if len(got.Variables) != len(tt.want.Variables) {
				t.Errorf("ParseTemplate() Variables count = %v, want %v", len(got.Variables), len(tt.want.Variables))
			}
			if len(got.Prompts) != len(tt.want.Prompts) {
				t.Errorf("ParseTemplate() Prompts count = %v, want %v", len(got.Prompts), len(tt.want.Prompts))
			}
		})
	}
}
