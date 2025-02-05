package tool

type Renovate struct {
	Datasource     string `json:"datasource" yaml:"datasource"`
	Package        string `json:"package" yaml:"package"`
	ExtractVersion string `json:"extractVersion,omitempty" yaml:"extractVersion,omitempty"`
	Versioning     string `json:"versioning,omitempty" yaml:"versioning,omitempty"`
}

type Messages struct {
	Internals string `json:"internals" yaml:"internals"`
	Usage     string `json:"usage" yaml:"usage"`
	Update    string `json:"update" yaml:"update"`
}

type License struct {
	Name string `json:"name" yaml:"name"`
	Link string `json:"link" yaml:"link"`
}

type Source struct {
	Registry   string `json:"registry" yaml:"registry"`
	Repository string `json:"repository" yaml:"repository"`
}

type Tool struct {
	SchemaVersion       string     `json:"schema_version,omitempty" yaml:"schema_version,omitempty"`
	Name                string     `json:"name" yaml:"name"`
	License             License    `json:"license" yaml:"license"`
	Version             string     `json:"version" yaml:"version"`
	Binary              string     `json:"binary,omitempty" yaml:"binary,omitempty"`
	Check               string     `json:"check,omitempty" yaml:"check"`
	Tags                []string   `json:"tags" yaml:"tags"`
	BuildDependencies   []string   `json:"build_dependencies,omitempty" yaml:"build_dependencies,omitempty"`
	RuntimeDependencies []string   `json:"runtime_dependencies,omitempty" yaml:"runtime_dependencies,omitempty"`
	Platforms           []string   `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	ConflictsWith       []string   `json:"conflicts_with,omitempty" yaml:"conflicts_with,omitempty"`
	Homepage            string     `json:"homepage,omitempty" yaml:"homepage,omitempty"`
	Repository          string     `json:"repository" yaml:"repository"`
	Description         string     `json:"description" yaml:"description"`
	Messages            Messages   `json:"messages,omitempty" yaml:"messages,omitempty"`
	Renovate            Renovate   `json:"renovate,omitempty" yaml:"renovate,omitempty"`
	Sources             []Source   `json:"sources" yaml:"sources"`
	Status              ToolStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type Tools struct {
	Tools []Tool `json:"tools" yaml:"tools"`
}

type ToolStatus struct {
	BinaryPresent      bool
	Version            string
	VersionMatches     bool
	MarkerFilePresent  bool
	MarkerFileVersion  string
	SkipDueToConflicts bool
	IsRequested        bool
}
