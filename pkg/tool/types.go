package tool

type Renovate struct {
	Datasource     string `json:"datasource"`
	Package        string `json:"package"`
	ExtractVersion string `json:"extractVersion,omitempty"`
	Versioning     string `json:"versioning,omitempty"`
}

type Messages struct {
	Internals string `json:"internals"`
	Usage     string `json:"usage"`
	Update    string `json:"update"`
}

type License struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

type Tool struct {
	SchemaVersion		string	 `json:"schema_version"`
	Name                string   `json:"name"`
	License			    License  `json:"license"`
	Version             string   `json:"version"`
	Binary              string   `json:"binary,omitempty"`
	Check               string   `json:"check,omitempty"`
	Tags                []string `json:"tags"`
	BuildDependencies   []string `json:"build_dependencies,omitempty"`
	RuntimeDependencies []string `json:"runtime_dependencies,omitempty"`
	Platforms           []string `json:"platforms,omitempty"`
	ConflictsWith       []string `json:"conflicts_with,omitempty"`
	Homepage            string   `json:"homepage,omitempty"`
	Repository          string   `json:"repository"`
	Description         string   `json:"description"`
	Messages            Messages `json:"messages,omitempty"`
	Renovate            Renovate `json:"renovate,omitempty"`
	Status              ToolStatus
}

type Tools struct {
	Tools []Tool `json:"tools"`
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
