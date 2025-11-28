package tool

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/google/safearchive/tar"

	"gitlab.com/uniget-org/cli/pkg/archive"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

type PathRewrite struct {
	Source    string
	Target    string
	Operation string
	Abort     bool
}

func applyPathRewrites(path string, rules []PathRewrite) string {
	logging.Debugf("Applying path rewrites to %s", path)

	newPath := path
	for _, rule := range rules {
		logging.Tracef("  Checking rule %v", rule)

		ruleWasApplied := false

		switch rule.Operation {
		case "REPLACE":
			if strings.HasPrefix(newPath, rule.Source) {
				newPath = rule.Target + strings.TrimPrefix(newPath, rule.Source)
				logging.Tracef("    Applied rule")
				ruleWasApplied = true
			}
		case "PREPEND":
			if !strings.HasPrefix(newPath, rule.Target) {
				newPath = rule.Target + newPath
				logging.Tracef("    Applied rule")
				ruleWasApplied = true
			}
		default:
			logging.Tracef("Operation %s not supported", rule.Operation)
		}

		if ruleWasApplied && rule.Abort {
			break
		}

		if strings.HasPrefix(newPath, "/") || strings.HasPrefix(newPath, "./") {
			break
		}
	}

	logging.Debugf("  New path is %s", newPath)
	return newPath
}

func (tool *Tool) Inspect(w io.Writer, layer io.ReadCloser, rules []PathRewrite) error {
	return archive.ProcessTarContents(layer, func(reader *tar.Reader, header *tar.Header) error {
		if header.Typeflag == tar.TypeDir {
			return nil
		}
		if len(rules) > 0 {
			header.Name = applyPathRewrites(header.Name, rules)
		}
		return archive.CallbackDisplayTarItem(reader, header)
	})
}

func (tool *Tool) Install(w io.Writer, layer io.ReadCloser, rules []PathRewrite, patchFile func(path string) string) ([]string, error) {
	installedFiles := []string{}

	err := archive.ProcessTarContents(layer, func(reader *tar.Reader, header *tar.Header) error {
		if header.Typeflag != tar.TypeDir {
			if header.Typeflag == tar.TypeLink && len(header.Linkname) > 0 {
				var err error

				absName, err := filepath.Abs(header.Name)
				if err != nil {
					return err
				}
				absLinkname, err := filepath.Abs(header.Linkname)
				if err != nil {
					return err
				}

				logging.Tracef("Name: %s, Linkname: %s", absName, absLinkname)
				header.Linkname, err = filepath.Rel(filepath.Dir(absName), absLinkname)
				if err != nil {
					return err
				}
				logging.Tracef("    Relative linkname is %s", header.Linkname)
			}
			header.Name = applyPathRewrites(header.Name, rules)

			err := archive.CallbackExtractTarItem(reader, header)
			if err != nil {
				return err
			}
			header.Name = patchFile(header.Name)

			installedFiles = append(installedFiles, header.Name)
		}
		return nil
	})
	if err != nil {
		return installedFiles, err
	}

	return installedFiles, nil
}
