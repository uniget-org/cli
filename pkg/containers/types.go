package containers

import (
	"context"
	"fmt"

	"github.com/regclient/regclient/types/ref"
	"github.com/uniget-org/cli/pkg/logging"
)

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

func FindToolRef(registries, repositories []string, tool, version string) (*ToolRef, error) {
	if len(registries) == 0 {
		return nil, fmt.Errorf("no registries provided")
	}
	if len(repositories) == 0 {
		return nil, fmt.Errorf("no repositories provided")
	}
	if len(registries) != len(repositories) {
		return nil, fmt.Errorf("number of registries and repositories do not match")
	}

	for index := range registries {
		toolRef := NewToolRef(registries[index], repositories[index], tool, version)
		logging.Tracef("Checking %s", toolRef)
		if toolRef.ManifestExists() {
			logging.Tracef("Found %s", toolRef)
			return toolRef, nil
		}
	}
	return nil, fmt.Errorf("tool %s:%s not found in sources", tool, version)
}

func (t *ToolRef) ManifestExists() bool {
	ref := t.GetRef()

	ctx := context.Background()
	rc := GetRegclient()
	//nolint:errcheck
	defer rc.Close(ctx, ref)

	b, err := HeadPlatformManifestForLocalPlatform(ctx, rc, ref)
	if err != nil {
		return false
	}

	return b
}

func (t *ToolRef) String() string {
	return fmt.Sprintf("%s/%s%s%s:%s", t.Registry, t.Repository, t.toolSeparator, t.Tool, t.Version)
}

func (t *ToolRef) Key() string {
	return fmt.Sprintf("%s-%s", t.Tool, t.Version)
}

func (t *ToolRef) GetRef() ref.Ref {
	r, err := ref.New(t.String())
	if err != nil {
		return ref.Ref{}
	}
	return r
}
