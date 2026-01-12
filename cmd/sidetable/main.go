package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/sushichan044/sidetable/internal/config"
	"github.com/sushichan044/sidetable/internal/delegate"
	"github.com/sushichan044/sidetable/version"
)

func main() {
	exitCode, err := run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if exitCode == 0 {
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func run(args []string, stdout, stderr io.Writer) (int, error) {
	configPath, showVersion, showHelp, remaining, err := parseGlobalFlags(args)
	if err != nil {
		return 1, err
	}

	if showHelp {
		return executeBuiltin([]string{"--help"}, stdout, stderr)
	}

	if showVersion && len(remaining) == 0 {
		fmt.Fprintln(stdout, version.Get())
		return 0, nil
	}

	if len(remaining) == 0 || isBuiltIn(remaining[0]) {
		return executeBuiltin(args, stdout, stderr)
	}

	if showVersion {
		fmt.Fprintln(stdout, version.Get())
		return 0, nil
	}

	path, err := resolveConfigPath(configPath)
	if err != nil {
		return 1, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return 1, err
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return 1, err
	}

	spec, err := delegate.Build(cfg, remaining[0], remaining[1:], projectDir)
	if err != nil {
		return 1, err
	}

	return delegate.Execute(spec), nil
}

func parseGlobalFlags(args []string) (string, bool, bool, []string, error) {
	fs := pflag.NewFlagSet("sidetable", pflag.ContinueOnError)
	fs.SetInterspersed(false)
	fs.SetOutput(io.Discard)
	fs.ParseErrorsAllowlist.UnknownFlags = true

	var configPath string
	var showVersion bool
	var showHelp bool
	fs.StringVar(&configPath, "config", "", "config path")
	fs.BoolVar(&showVersion, "version", false, "show version")
	fs.BoolVar(&showHelp, "help", false, "show help")

	if err := fs.Parse(args); err != nil {
		return "", false, false, nil, err
	}

	return configPath, showVersion, showHelp, fs.Args(), nil
}

func executeBuiltin(args []string, stdout, stderr io.Writer) (int, error) {
	root := newRootCommand(stdout, stderr)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		return 1, err
	}
	return 0, nil
}

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
	var configPath string

	root := &cobra.Command{
		Use:           "sidetable",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)

	root.PersistentFlags().StringVar(&configPath, "config", "", "config path")

	root.AddCommand(newListCommand(&configPath))
	root.AddCommand(newVersionCommand(stdout))

	return root
}

func newVersionCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Fprintln(stdout, version.Get())
		},
	}
}

func newListCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available commands",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := resolveConfigPath(*configPath)
			if err != nil {
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			projectDir, err := os.Getwd()
			if err != nil {
				return err
			}

			var desc string
			for _, name := range cfg.CommandNames() {
				commandCfg := cfg.Commands[name]
				desc, err = renderDescription(commandCfg.Description, projectDir, cfg.Directory, name)
				if err != nil {
					return err
				}
				alias := commandCfg.Alias
				if alias != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\t%s\n", name, alias, desc)
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", name, desc)
			}
			return nil
		},
	}
}

func renderDescription(description, projectDir, privateDirName, commandName string) (string, error) {
	if description == "" {
		return "", nil
	}

	ctx := delegateTemplateContext(projectDir, privateDirName, commandName)
	return executeTemplate(description, ctx)
}

type templateContext struct {
	ProjectDir string
	PrivateDir string
	CommandDir string
	Args       []string
}

func delegateTemplateContext(projectDir, privateDirName, commandName string) templateContext {
	privateDir := filepath.Join(projectDir, privateDirName)
	commandDir := filepath.Join(privateDir, commandName)
	return templateContext{
		ProjectDir: projectDir,
		PrivateDir: privateDir,
		CommandDir: commandDir,
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

func resolveConfigPath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}
	return config.ResolvePath()
}

func isBuiltIn(name string) bool {
	switch name {
	case "help", "list", "version":
		return true
	default:
		return false
	}
}
