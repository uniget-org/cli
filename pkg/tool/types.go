package tool

type Renovate struct {
	Datasource       string   `json:"datasoruce"`
	Package          string   `json:"package"`
    ExtractVersion   string   `json:"extractVersion,omitempty"`
	Versioning       string   `json:"versioning,omitempty"`
}

type Tool struct {
	Name                string     `json:"name"`
	Version             string     `json:"version"`
	Binary              string     `json:"binary,omitempty"`
	Check               string     `json:"check,omitempty"`
	Tags                []string   `json:"tags"`
	BuildDependencies   []string   `json:"build_dependencies,omitempty"`
	RuntimeDependencies []string   `json:"runtime_dependencies,omitempty"`
    Platforms           []string   `json:"platforms,omitempty`
	Homepage            string     `json:"homepage"`
	Description         string     `json:"description"`
	Renovate            Renovate   `json:"renovate,omitempty"`
}

type Tools struct {
	Tools []Tool `json:"tools"`
}

type ToolStatus struct {
	Name           string
	BinaryPresent  bool
	Version        string
	VersionMatches bool
}