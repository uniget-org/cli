package tool

import (
	"archive/tar"
	"io"
	"strings"

	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/logging"
)

type PathRewrite struct {
	Source    string
	Target    string
	Operation string
}

func applyPathRewrites(path string, rules []PathRewrite) string {
	logging.Debugf("Applying path rewrites to %s", path)

	newPath := path
	for _, rule := range rules {
		logging.Debugf("  Checking rule %v", rule)

		if rule.Operation == "REPLACE" {
			if strings.HasPrefix(newPath, rule.Source) {
				newPath = rule.Target + strings.TrimPrefix(newPath, rule.Source)
				logging.Debugf("    Applied rule")
			}

		} else if rule.Operation == "PREPEND" {
			if !strings.HasPrefix(newPath, rule.Target) {
				newPath = rule.Target + newPath
				logging.Debugf("    Applied rule")
			}

		} else {
			logging.Debugf("Operation %s not supported", rule.Operation)
		}

		if strings.HasPrefix(newPath, "/") || strings.HasPrefix(newPath, "./") {
			break
		}
	}

	logging.Debugf("  New path is %s", newPath)
	return newPath
}

func (tool *Tool) Inspect(w io.Writer, layer []byte, rules []PathRewrite) error {
	return archive.ProcessTarContents(layer, func(reader *tar.Reader, header *tar.Header) error {
		header.Name = applyPathRewrites(header.Name, rules)
		return archive.CallbackDisplayTarItem(reader, header)
	})
}

func (tool *Tool) Install(w io.Writer, layer []byte, rules []PathRewrite, patchFile func(path string)) error {
	return archive.ProcessTarContents(layer, func(reader *tar.Reader, header *tar.Header) error {
		header.Name = applyPathRewrites(header.Name, rules)
		err := archive.CallbackExtractTarItem(reader, header)
		if err != nil {
			return err
		}
		patchFile(header.Name)
		return nil
	})
}
