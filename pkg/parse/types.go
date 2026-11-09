package parse

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

type ImageRefs struct {
	Refs       []ref.Ref
	BumpedRefs []ref.Ref
}

func (r *ImageRefs) Add(ref ref.Ref) {
	r.Refs = append(r.Refs, ref)
}

func (r *ImageRefs) Bump(tools *tool.Tools, callback func(toolName string, oldVersion string, newVersion string)) error {
	r.BumpedRefs = make([]ref.Ref, len(r.Refs))

	supportRegistries := []string{
		"ghcr.io",
		"registry.gitlab.com",
	}

	for index, reference := range r.Refs {
		logging.Debugf("Bumping image reference: %s", reference)

		if slices.Contains(supportRegistries, reference.Registry) && strings.HasPrefix(reference.Repository, "uniget-org/tools/") {
			toolName := reference.Repository[17:]
			oldVersion := reference.Tag

			tool, err := tools.GetByName(toolName)
			if err != nil {
				return fmt.Errorf("tool %s not found in metadata: %s", toolName, err)
			}

			if oldVersion != tool.Version {
				callback(toolName, oldVersion, tool.Version)
			}

			reference.Tag = tool.Version
			reference.Digest = ""
			reference.Reference = fmt.Sprintf("%s/%s:%s", reference.Registry, reference.Repository, reference.Tag)
			reference.Digest, err = containers.FindNewDigest(reference)
			if err != nil {
				return fmt.Errorf("failed to find new digest for %s: %w", reference, err)
			}

			refReplacement := fmt.Sprintf("%s/%s:%s@%s", reference.Registry, reference.Repository, reference.Tag, reference.Digest)

			r.BumpedRefs[index], err = ref.New(refReplacement)
			if err != nil {
				return fmt.Errorf("failed to create ref from %q: %v", refReplacement, err)
			}

		} else {
			r.BumpedRefs[index] = reference
		}
	}

	return nil
}

func (r *ImageRefs) Replace(file []byte) ([]byte, error) {
	for i, reference := range r.Refs {
		file = bytes.ReplaceAll(file, []byte(reference.Reference), []byte(r.BumpedRefs[i].Reference))
	}

	return file, nil
}
