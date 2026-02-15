package config

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/Oudwins/zog/zconst"

	z "github.com/Oudwins/zog"

	"github.com/sushichan044/sidetable/internal/builtin"
)

const (
	msgDirectoryRequired       = "directory is required"
	msgDirectoryMustBeRelative = "directory must be relative"

	msgToolRunRequired            = "tool run is required"
	msgToolRunMustNotContainSpace = "tool run must not contain spaces"
	msgToolConflictsWithBuiltin   = "tool conflicts with builtin command"

	msgAliasNameRequired         = "alias name is required"
	msgAliasMustNotContainSpaces = "alias must not contain spaces"
	msgAliasToolRequired         = "alias tool is required"
	msgAliasConflictsWithTool    = "alias conflicts with tool name"
	msgAliasConflictsWithBuiltin = "alias conflicts with builtin command"
	msgAliasTargetUnknown        = "alias tool not found"
)

var (
	toolSchema = z.Struct(z.Shape{
		"run": z.String().
			Required(z.Message(msgToolRunRequired)).
			TestFunc(func(val *string, _ z.Ctx) bool {
				return !strings.ContainsAny(*val, " \t\n\r")
			}, z.Message(msgToolRunMustNotContainSpace)),
	})
	toolNameSchema = z.String().
			TestFunc(func(val *string, _ z.Ctx) bool {
			return !builtin.IsReservedName(*val)
		}, z.Message(msgToolConflictsWithBuiltin))

	aliasSchema = z.Struct(z.Shape{
		"tool": z.String().Required(z.Message(msgAliasToolRequired)),
	})
	aliasNameSchema = z.String().
			Required(z.Message(msgAliasNameRequired)).
			TestFunc(func(val *string, _ z.Ctx) bool {
			return !strings.ContainsAny(*val, " \t\n\r")
		}, z.Message(msgAliasMustNotContainSpaces)).
		TestFunc(func(val *string, _ z.Ctx) bool {
			return !builtin.IsReservedName(*val)
		}, z.Message(msgAliasConflictsWithBuiltin))

	configSchema = z.Struct(z.Shape{
		"directory": z.String().
			Required(z.Message(msgDirectoryRequired)).
			TestFunc(func(val *string, _ z.Ctx) bool {
				return strings.TrimSpace(*val) != ""
			}, z.Message(msgDirectoryRequired)).
			TestFunc(func(val *string, _ z.Ctx) bool {
				return !filepath.IsAbs(*val)
			}, z.Message(msgDirectoryMustBeRelative)),
		"tools": z.EXPERIMENTAL_MAP[string, Tool](
			toolNameSchema,
			toolSchema,
		),
		"aliases": z.EXPERIMENTAL_MAP[string, Alias](
			aliasNameSchema,
			aliasSchema,
		),
	})
)

func (c *Config) validateWithSchema() z.ZogIssueList {
	if c == nil {
		return z.ZogIssueList{newCustomIssue(nil, "config is nil")}
	}

	issues := make(z.ZogIssueList, 0)
	issues = append(issues, configSchema.Validate(c)...)
	issues = append(issues, validateCrossRules(c)...)

	sort.SliceStable(issues, func(i, j int) bool {
		lhsPath := issues[i].PathString()
		rhsPath := issues[j].PathString()
		if lhsPath != rhsPath {
			return lhsPath < rhsPath
		}
		return issues[i].Message < issues[j].Message
	})

	return issues
}

func validateCrossRules(config *Config) z.ZogIssueList {
	issues := make(z.ZogIssueList, 0)

	aliasNames := make([]string, 0, len(config.Aliases))
	for aliasName := range config.Aliases {
		aliasNames = append(aliasNames, aliasName)
	}
	sort.Strings(aliasNames)

	for _, aliasName := range aliasNames {
		alias := config.Aliases[aliasName]
		aliasPath := []string{"aliases", bracketKey(aliasName)}

		if _, exists := config.Tools[aliasName]; exists {
			issues = append(issues, newCustomIssue(aliasPath, msgAliasConflictsWithTool))
		}

		aliasTool := strings.TrimSpace(alias.Tool)
		if aliasTool != "" {
			if _, exists := config.Tools[aliasTool]; !exists {
				aliasToolPath := []string{"aliases", bracketKey(aliasName), "tool"}
				issues = append(issues, newCustomIssue(aliasToolPath, msgAliasTargetUnknown))
			}
		}
	}

	return issues
}

func bracketKey(key string) string {
	return `["` + key + `"]`
}

func newCustomIssue(path []string, message string) *z.ZogIssue {
	return (&z.ZogIssue{}).
		SetCode(zconst.IssueCodeCustom).
		SetPath(path).
		SetMessage(message)
}

type validationIssueError struct {
	issue *z.ZogIssue
}

func newValidationIssueError(issue *z.ZogIssue) error {
	return validationIssueError{issue: issue}
}

func (e validationIssueError) Error() string {
	if e.issue == nil {
		return "invalid config"
	}

	msg := e.issue.Message
	if msg == "" {
		msg = e.issue.Error()
	}

	path := e.issue.PathString()
	if path == "" {
		return msg
	}

	return path + ": " + msg
}

func (e validationIssueError) Unwrap() error {
	return e.issue
}
