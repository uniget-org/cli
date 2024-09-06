package containers

import "fmt"

type ToolRef struct {
	Registry      string
	Repository    string
	toolSeparator string
	Tool          string
	Version       string
}

func NewToolRef(registry, repository, tool, version string) *ToolRef {
	return &ToolRef{
		Registry:      registry,
		Repository:    repository,
		toolSeparator: "/",
		Tool:          tool,
		Version:       version,
	}
}

func (t *ToolRef) String() string {
	return fmt.Sprintf("%s/%s%s%s:%s", t.Registry, t.Repository, t.toolSeparator, t.Tool, t.Version)
}

func (t *ToolRef) Key() string {
	return fmt.Sprintf("%s-%s", t.Tool, t.Version)
}
