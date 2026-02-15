package scaf_fold

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed all:_template
var templateFS embed.FS

func renderTemplate(filePath string, raw []byte, data TemplateData) ([]byte, error) {
	tmpl, err := template.New(filePath).Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", filePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", filePath, err)
	}

	return buf.Bytes(), nil
}
