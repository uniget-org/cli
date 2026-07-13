package os

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gitlab.com/uniget-org/cli/pkg/logging"
	"golang.org/x/sys/unix"
)

func DirectoryIsWritable(directory string) bool {
	logging.Debugf("Checking if directory %s is writable", directory)
	return unix.Access(directory, unix.W_OK) == nil
}

func AssertDirectory(directory string) {
	logging.Debugf("Creating directory %s", directory)
	err := os.MkdirAll(directory, 0755) // #nosec G301 -- Directories will contain public information
	if err != nil {
		logging.Error.Printfln("Error creating directory %s: %s", directory, err)
		os.Exit(1)
	}
}

func AssertWritableDirectory(directory string) {
	if !DirectoryExists(directory) {
		AssertDirectory(directory)
	}
	if !DirectoryIsWritable(directory) {
		logging.Error.Printfln("Directory %s is not writable", directory)
		os.Exit(1)
	}
}

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
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Get the directory part of the path
	dir := filepath.Dir(cleanPath)

	// If dir is "." then no subdirectories need to be created
	if dir == "." {
		return nil
	}

	// Create the full path within workDir
	fullDir := filepath.Join(workDir, dir)

	// Create all directories
	err := os.MkdirAll(fullDir, 0755) // #nosec G301 -- Tools must be world readable
	if err != nil {
		return fmt.Errorf("failed to create directories for path %s in workDir %s: %w", path, workDir, err)
	}

	logging.Debugf("Created directories for path: %s in workDir: %s", path, workDir)
	return nil
}
