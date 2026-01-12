package sidetable

import (
	"path/filepath"
	"strings"
	"text/template"
)

type templateContext struct {
	ProjectDir string
	PrivateDir string
	CommandDir string
	ConfigDir  string
	Args       []string
}

func renderDescription(description, projectDir, privateDirName, commandName, configDir string) (string, error) {
	if description == "" {
		return "", nil
	}

	ctx := delegateTemplateContext(projectDir, privateDirName, commandName, configDir)
	return executeTemplate(description, ctx)
}

func delegateTemplateContext(projectDir, privateDirName, commandName, configDir string) templateContext {
	privateDir := filepath.Join(projectDir, privateDirName)
	commandDir := filepath.Join(privateDir, commandName)
	return templateContext{
		ProjectDir: projectDir,
		PrivateDir: privateDir,
		CommandDir: commandDir,
		ConfigDir:  configDir,
		Args:       nil,
	}
}

func executeTemplate(raw string, ctx templateContext) (string, error) {
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
