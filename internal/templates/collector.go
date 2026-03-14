package templates

import "github.com/AnishShah1803/jotr/internal/utils"

func CollectVariableValues(vars []Variable) (map[string]string, error) {
	values := make(map[string]string)

	for _, v := range vars {
		value, err := utils.PromptUserRequired(v.Prompt)
		if err != nil {
			return nil, err
		}
		values[v.Name] = value
	}

	return values, nil
}

func CollectPromptValues(prompts []Prompt) ([]string, error) {
	var values []string

	for _, p := range prompts {
		value, err := utils.PromptUserRequired(p.Question)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, nil
}
