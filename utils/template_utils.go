package utils

import (
	"bytes"
	"fmt"
	"text/template"
)

func DoTemplate(name string, text string, data any) (string, error) {
	template, err := template.New(name).Parse(text)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s template: %w", name, err)
	}

	var resultBuffer bytes.Buffer

	err = template.Execute(&resultBuffer, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute %s template: %w", name, err)
	}

	result := resultBuffer.String()
	return result, nil
}
