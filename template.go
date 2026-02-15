package sidetable

import (
	"strings"
	"text/template"
)

type templateContext struct {
	WorkspaceRoot string
	ToolDir       string
	ConfigDir     string
}

func evalTemplate(raw string, ctx templateContext) (string, error) {
	tpl, err := template.New("value").Option("missingkey=error").Parse(raw)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if execErr := tpl.Execute(&b, ctx); execErr != nil {
		return "", execErr
	}
	return b.String(), nil
}
