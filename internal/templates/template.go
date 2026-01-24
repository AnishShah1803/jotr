package templates

type Variable struct {
	Name   string
	Prompt string
	Value  string
}

type Prompt struct {
	Question string
	Variable *string
}

type Template struct {
	Filename   string
	Priority   int
	Category   string
	Name       string
	Content    string
	TargetPath string
	Variables  []Variable
	Prompts    []Prompt
	BuiltIns   map[string]string
}
