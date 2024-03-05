package tool

import (
	"fmt"
	"os"
	"strings"

	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/logging"

	"github.com/pterm/pterm"
	"github.com/regclient/regclient/types/blob"
)

func (tool *Tool) Install(registryImagePrefix string, prefix string, target string, libDirectory string, cacheDirectory string) error {
	// Fetch manifest for tool
	err := containers.GetManifest(fmt.Sprintf(registryImagePrefix+"%s:%s", tool.Name, strings.Replace(tool.Version, "+", "-", -1)), func(blob blob.Reader) error {
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
		})
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

func (tool *Tool) Inspect(registryImagePrefix string, raw bool) error {
	// Fetch manifest for tool
	err := containers.GetManifest(fmt.Sprintf(registryImagePrefix+"%s:%s", tool.Name, strings.Replace(tool.Version, "+", "-", -1)), func(blob blob.Reader) error {
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
			fmt.Printf("%s\n", file)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get manifest: %s", err)
	}

	return nil
}
