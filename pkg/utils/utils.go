package utils

import (
	"fmt"

	"github.com/charmbracelet/glamour"
	"gopkg.in/yaml.v3"
)

func All[K comparable, V any](m map[K]V, condition func(V) bool) bool {
	for _, value := range m {
		if !condition(value) {
			return false // Return false if any value doesn't meet the condition
		}
	}
	return true // Return true if all values meet the condition
}

func PrettyYAML(input interface{}) string {
	yamlData, err := yaml.Marshal(input)
	if err != nil {
		return fmt.Sprintf("error marshalling YAML: %v", err)
	}

	markdownContent := fmt.Sprintf("```yaml\n%s\n```", yamlData)

	renderedContent, err := glamour.Render(markdownContent, "dark")
	if err != nil {
		return fmt.Sprintf("error rendering markdown: %v", err)
	}

	return renderedContent
}
