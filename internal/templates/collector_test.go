package templates

import (
	"testing"
)

func TestCollectVariableValues(t *testing.T) {
	t.Skip("CollectVariableValues requires stdin mocking - tested via integration tests")

	vars := []Variable{
		{Name: "topic", Prompt: "What is the topic?"},
		{Name: "location", Prompt: "Where is it?"},
	}

	values, err := CollectVariableValues(vars)

	if err != nil {
		t.Errorf("CollectVariableValues() unexpected error: %v", err)
	}

	if len(values) != 2 {
		t.Errorf("CollectVariableValues() got %d values, want 2", len(values))
	}
}

func TestCollectPromptValues(t *testing.T) {
	t.Skip("CollectPromptValues requires stdin mocking - tested via integration tests")

	prompts := []Prompt{
		{Question: "First question?"},
		{Question: "Second question?"},
	}

	values, err := CollectPromptValues(prompts)

	if err != nil {
		t.Errorf("CollectPromptValues() unexpected error: %v", err)
	}

	if len(values) != 2 {
		t.Errorf("CollectPromptValues() got %d values, want 2", len(values))
	}
}
