package os

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/uniget-org/cli/pkg/logging"
)

func IsDirectoryEmpty(name string) bool {
	//gosec:disable G304 -- This is a false positive
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	//nolint:errcheck
	defer f.Close()

	_, err = f.Readdirnames(1)
	return err == io.EOF
}

func CreateSubdirectoriesForPath(workDir, path string) error {
	// Clean the path and reject anything that would escape workDir. Note that
	// filepath.Clean alone does not remove leading ".." elements.
	cleanPath := filepath.Clean(path)
	if !filepath.IsLocal(cleanPath) {
		return fmt.Errorf("refusing to create directories for path %q outside of %q", path, workDir)
	}

	// Get the directory part of the path
	dir := filepath.Dir(cleanPath)

	// If dir is "." then no subdirectories need to be created
	if dir == "." {
		return nil
	}

	// Create all directories relative to workDir. os.Root prevents escaping
	// workDir through ".." elements or symbolic links.
	root, err := os.OpenRoot(workDir)
	if err != nil {
		return fmt.Errorf("failed to open workDir %s: %w", workDir, err)
	}
	//nolint:errcheck
	defer root.Close()

	err = root.MkdirAll(dir, 0755) // #nosec G301 -- Tools must be world readable
	if err != nil {
		return fmt.Errorf("failed to create directories for path %s in workDir %s: %w", path, workDir, err)
	}

	logging.Debugf("Created directories for path: %s in workDir: %s", path, workDir)
	return nil
}

// pathIsWithin reports whether target is located inside base. Both paths are
// expected to be absolute and normalized.
func PathIsWithin(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}

	return rel == "." ||
		(rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && !filepath.IsAbs(rel))
}

// validateLinkTarget ensures a link cannot point outside of the extraction
// root. Symlink targets are resolved relative to the directory containing the
// link, while hard link targets are resolved relative to the extraction root
// as defined by the tar format.
func ValidateLinkTarget(workDir, name string, header *tar.Header) error {
	if len(header.Linkname) == 0 {
		return fmt.Errorf("refusing to create link %q with an empty target", header.Name)
	}

	var resolved string
	switch header.Typeflag {
	case tar.TypeLink:
		linkName := filepath.Clean(header.Linkname)
		if !filepath.IsLocal(linkName) {
			return fmt.Errorf("refusing to create link %q with target %q outside of %q", header.Name, header.Linkname, workDir)
		}
		resolved = filepath.Join(workDir, linkName)
	default:
		if filepath.IsAbs(header.Linkname) {
			resolved = filepath.Clean(header.Linkname)
		} else {
			resolved = filepath.Join(workDir, filepath.Dir(name), header.Linkname) // #nosec G305 -- Result is validated with PathIsWithin below
		}
	}

	if !PathIsWithin(workDir, resolved) {
		return fmt.Errorf("refusing to create link %q with target %q outside of %q", header.Name, header.Linkname, workDir)
	}

	return nil
}

// binaryFileIsInUse reports whether the binary of a tool can be opened for
// writing. Only paths beneath the target tree are probed, and os.Root is
// used so that a manipulated metadata file can never make uniget open files
// outside of the installation target, not even through symbolic links.
func BinaryFileIsInUse(targetDirectory string, binaryPath string) bool {
	absBinaryPath, err := filepath.Abs(binaryPath)
	if err != nil {
		logging.Debugf("Unable to determine absolute path of %s: %s", binaryPath, err)
		return false
	}
	absTargetDirectory, err := filepath.Abs(targetDirectory)
	if err != nil {
		logging.Debugf("Unable to determine absolute path of %s: %s", targetDirectory, err)
		return false
	}
	if !PathIsWithin(absTargetDirectory, absBinaryPath) {
		logging.Warning.Printfln("Binary %s is outside of target %s. Skipping check whether it is in use", binaryPath, targetDirectory)
		return false
	}
	relativeBinaryPath, err := filepath.Rel(absTargetDirectory, absBinaryPath)
	if err != nil {
		logging.Debugf("Unable to determine relative path of %s: %s", binaryPath, err)
		return false
	}

	root, err := os.OpenRoot(absTargetDirectory)
	if err != nil {
		logging.Debugf("Unable to open target %s: %s", targetDirectory, err)
		return false
	}
	//nolint:errcheck
	defer root.Close()

	file, err := root.OpenFile(relativeBinaryPath, os.O_WRONLY, 0600)
	if err != nil {
		// os.Root also refuses to follow symbolic links leaving the
		// target. Such an error is reported as in use to stay on the
		// safe side.
		logging.Debugf("Unable to open binary %s for writing: %s", binaryPath, err)
		return true
	}
	//nolint:errcheck
	defer file.Close()

	return false
}
