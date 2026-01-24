package templates

import "github.com/AnishShah1803/jotr/internal/utils"

func CollectVariableValues(vars []Variable) (map[string]string, error) {
	values := make(map[string]string)

	for _, v := range vars {
		value := utils.PromptUserRequired(v.Prompt)
		values[v.Name] = value
	}

	return values, nil
}

func CollectPromptValues(prompts []Prompt) ([]string, error) {
	var values []string

	for _, p := range prompts {
		value := utils.PromptUserRequired(p.Question)
		values = append(values, value)
	}

	return values, nil
}
