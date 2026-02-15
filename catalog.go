package sidetable

import (
	"errors"
	"sort"
)

// EntryKind describes catalog entry type.
type EntryKind string

const (
	EntryKindTool  EntryKind = "tool"
	EntryKindAlias EntryKind = "alias"
)

// Entry is a listable tool or alias.
type Entry struct {
	Name        string
	Kind        EntryKind
	Target      string
	Description string
}

// Catalog contains all listable entries.
type Catalog struct {
	Entries []Entry
}

// Catalog returns tools and aliases available in this workspace.
func (w *Workspace) Catalog() (*Catalog, error) {
	if w == nil || w.config == nil {
		return nil, errors.New("workspace is not initialized")
	}

	entries := make([]Entry, 0, len(w.config.Tools)+len(w.config.Aliases))
	for _, name := range w.config.ToolNames() {
		tool := w.config.Tools[name]
		entries = append(entries, Entry{
			Name:        name,
			Kind:        EntryKindTool,
			Description: tool.Description,
		})
	}

	aliasNames := make([]string, 0, len(w.config.Aliases))
	for name := range w.config.Aliases {
		aliasNames = append(aliasNames, name)
	}
	sort.Strings(aliasNames)
	for _, name := range aliasNames {
		alias := w.config.Aliases[name]
		entries = append(entries, Entry{
			Name:        name,
			Kind:        EntryKindAlias,
			Target:      alias.Tool,
			Description: alias.Description,
		})
	}

	return &Catalog{Entries: entries}, nil
}
