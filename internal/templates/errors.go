package templates

import "fmt"

var (
	ErrTemplateNotFound  = fmt.Errorf("template not found")
	ErrInvalidFilename   = fmt.Errorf("invalid template filename format")
	ErrParseFailed       = fmt.Errorf("failed to parse template")
	ErrVariableConflict  = fmt.Errorf("variable defined multiple times")
	ErrInvalidPathSyntax = fmt.Errorf("invalid path syntax in <!-- Path: -->")
)

type TemplateError struct {
	Template string
	Err      error
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf("template %s: %s", e.Template, e.Err.Error())
}

func (e *TemplateError) Unwrap() error {
	return e.Err
}
