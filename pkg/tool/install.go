package tool

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/logging"

	"github.com/regclient/regclient/types/blob"
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

func (tool *Tool) InstallWithPathRewrites(registry, imageRepository string, prefix string, rules []PathRewrite, patchFile func(path string)) error {
	// Fetch manifest for tool
	toolRef := containers.NewToolRef(registry, imageRepository, tool.Name, strings.Replace(tool.Version, "+", "-", -1))
	err := containers.GetManifestOld(toolRef, func(blob blob.Reader) error {
		logging.Debugf("Extracting with prefix=%s", prefix)

		// Change working directory to prefix
		// so that unpacking can ignore the target directory
		installDir := prefix
		if len(prefix) == 0 {
			installDir = "/"
		}
		err := os.Chdir(installDir)
		if err != nil {
			return fmt.Errorf("error changing directory to %s: %s", prefix, err)
		}
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting working directory")
		}
		logging.Debugf("Current directory: %s", dir)

		// Unpack tool
		err = archive.ExtractTarGz(blob, func(path string) string {
			return applyPathRewrites(path, rules)
		}, patchFile)
		if err != nil {
			return fmt.Errorf("failed to extract layer: %s", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get manifest: %s", err)
	}

	return nil
}

func (tool *Tool) Install(registry, imageRepository string, prefix string, target string, libDirectory string, cacheDirectory string, patchFile func(path string)) error {
	// Fetch manifest for tool
	toolRef := containers.NewToolRef(registry, imageRepository, tool.Name, strings.Replace(tool.Version, "+", "-", -1))
	err := containers.GetManifestOld(toolRef, func(blob blob.Reader) error {
		logging.Debugf("Extracting with prefix=%s and target=%s", prefix, target)

		// Change working directory to prefix
		// so that unpacking can ignore the target directory
		installDir := prefix
		if len(prefix) == 0 {
			installDir = "/"
		}
		err := os.Chdir(installDir)
		if err != nil {
			return fmt.Errorf("error changing directory to %s: %s", prefix, err)
		}
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting working directory")
		}
		logging.Debugf("Current directory: %s", dir)

		// Unpack tool
		err = archive.ExtractTarGz(blob, func(path string) string {
			// Skip paths that are a prefix of usr/local/
			// Necessary as long as tools are still installed in hardcoded /usr/local
			if strings.HasPrefix("usr/local/", path) {
				pterm.Debug.Println("Path is prefix of usr/local/")
				return ""
			}

			// Remove prefix usr/local/ to support arbitrary target directories
			fixedPath := strings.TrimPrefix(path, "usr/local/")

			// Fix lib directory
			if strings.HasPrefix(fixedPath, "var/lib/uniget/") {
				logging.Debugf("Replacing lib directory with %s", libDirectory)
				fixedPath = libDirectory + "/" + strings.TrimPrefix(fixedPath, "var/lib/uniget/")

				// Fix cache directory
			} else if strings.HasPrefix(fixedPath, "var/cache/uniget/") {
				logging.Debugf("Replacing cache directory with %s", cacheDirectory)
				fixedPath = cacheDirectory + "/" + strings.TrimPrefix(fixedPath, "var/cache/uniget/")

				// Prepending target
			} else if len(target) > 0 {
				logging.Debugf("Prepending target to %s", fixedPath)
				fixedPath = target + "/" + fixedPath
			}

			return fixedPath
		}, patchFile)
		if err != nil {
			return fmt.Errorf("failed to extract layer: %s", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get manifest: %s", err)
	}

	return nil
}

func (tool *Tool) InspectOld(w io.Writer, registry, imageRepository string, raw bool) error {
	// Fetch manifest for tool
	toolRef := containers.NewToolRef(registry, imageRepository, tool.Name, strings.Replace(tool.Version, "+", "-", -1))
	err := containers.GetManifestOld(toolRef, func(blob blob.Reader) error {
		result, err := archive.ListTarGz(blob, func(path string) string {
			// Remove prefix usr/local/ to support arbitrary target directories
			// Necessary as long as tools are still installed in hardcoded /usr/local
			fixedPath := path
			if !raw {
				fixedPath = strings.TrimPrefix(path, "usr/local/")
			}
			return fixedPath
		})
		if err != nil {
			return fmt.Errorf("failed to extract layer: %s", err)
		}

		// Display contents of tool image
		for _, file := range result {
			//fmt.Printf("%s\n", file)
			fmt.Fprintf(w, "%s\n", file)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get manifest: %s", err)
	}

	return nil
}

func (tool *Tool) InspectWithPathRewritesOld(w io.Writer, registry, imageRepository string, raw bool, rules []PathRewrite) error {
	// Fetch manifest for tool
	toolRef := containers.NewToolRef(registry, imageRepository, tool.Name, strings.Replace(tool.Version, "+", "-", -1))
	err := containers.GetManifestOld(toolRef, func(blob blob.Reader) error {
		result, err := archive.ListTarGz(blob, func(path string) string {
			return applyPathRewrites(path, rules)
		})
		if err != nil {
			return fmt.Errorf("failed to extract layer: %s", err)
		}

		// Display contents of tool image
		for _, file := range result {
			fmt.Fprintf(w, "%s\n", file)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get manifest: %s", err)
	}

	return nil
}

func (tool *Tool) Inspect(w io.Writer, layer []byte) error {
	return archive.ProcessTarContents(layer, archive.CallbackDisplayTarItem)
}
