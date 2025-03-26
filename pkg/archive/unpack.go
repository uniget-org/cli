package archive

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/safearchive/tar"
	"github.com/google/safeopen"

	"github.com/uniget-org/cli/pkg/logging"
)

func Gunzip(layer []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(layer))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %s", err)
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			logging.Warning.Printfln("failed to close gzip reader: %s", err)
		}
	}()

	buffer, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read gzip: %s", err)
	}

	return buffer, nil
}

func ProcessTarContents(archive []byte, callback func(tar *tar.Reader, header *tar.Header) error) error {
	tarReader := tar.NewReader(bytes.NewReader(archive))

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to read next item: %s", err.Error())
		}

		err = callback(tarReader, header)
		if err != nil {
			return fmt.Errorf("failed to process item through callback: %s", err.Error())
		}
	}

	return nil
}

func CallbackDisplayTarItem(reader *tar.Reader, header *tar.Header) error {
	switch header.Typeflag {
	case tar.TypeDir:
	case tar.TypeReg:
		//nolint:errcheck
		fmt.Fprintf(logging.OutputWriter, "%s\n", header.Name)
	case tar.TypeSymlink:
		//nolint:errcheck
		fmt.Fprintf(logging.OutputWriter, "%s -> %s\n", header.Name, header.Linkname)
	case tar.TypeLink:
		//nolint:errcheck
		fmt.Fprintf(logging.OutputWriter, "%s -> %s\n", header.Name, header.Linkname)
	default:
		//nolint:errcheck
		fmt.Fprintf(logging.ErrorWriter, "Unknown: %s\n", header.Name)
	}
	return nil
}

func CallbackExtractTarItem(reader *tar.Reader, header *tar.Header) error {
	if header.Typeflag == tar.TypeDir {
		return nil
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory")
	}
	logging.Debugf("Current directory: %s", workDir)

	switch header.Typeflag {

	case tar.TypeReg:
		logging.Debugf("File: %s", header.Name)

		path := filepath.Clean(header.Name)

		dir := filepath.Dir(path)
		err := os.MkdirAll(dir, 0755) // #nosec G301 -- Tools must be world readable
		if err != nil {
			return fmt.Errorf("ExtractTarGz: MkdirAll() failed for %s: %s", dir, err.Error())
		}

		outFile, err := safeopen.CreateBeneath(workDir, path)
		if err != nil {
			return fmt.Errorf("ExtractTarGz: Create() failed for %s in %s: %s", path, workDir, err.Error())
		}
		defer func() {
			err := outFile.Close()
			if err != nil {
				logging.Warning.Printfln("failed to close file: %s", err)
			}
		}()
		if _, err := io.Copy(outFile, reader); err != nil {
			return fmt.Errorf("ExtractTarGz: Copy() failed for %s: %s", path, err.Error())
		}

		mode := os.FileMode(header.Mode) // #nosec G115 -- Must be addressed upstream
		Setuid := mode &^ 0777
		if (mode & Setuid) != 0 {
			logging.Warning.Printfln("Setuid bit cannot be set for %s", path)
		}
		err = outFile.Chmod(mode)
		if err != nil {
			return fmt.Errorf("ExtractTarGz: Chmod() failed for %s: %s", path, err.Error())
		}

	case tar.TypeSymlink:
		logging.Tracef("Untarring symlink %s", header.Name)

		// Check if symlink already exists
		_, err := os.Lstat(header.Name)
		if err == nil {
			logging.Debugf("Symlink %s already exists", header.Name)
		}
		// Continue if symlink does not exist
		if os.IsNotExist(err) {
			logging.Debugf("Symlink %s does not exist", header.Name)

			// Create directories for symlink
			dir := filepath.Dir(header.Name)
			logging.Tracef("Creating directory %s", dir)
			err := os.MkdirAll(dir, 0755) // #nosec G301 -- Tools must be world readable
			if err != nil {
				return fmt.Errorf("ExtractTarGz: MkdirAll() failed for %s: %s", dir, err.Error())
			}

			_, err = os.Stat(header.Linkname)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("ExtractTarGz: Stat() failed for TypeSymlink for %s: %s", header.Linkname, err.Error())
				}
			} else {
				err = os.Remove(header.Linkname)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Remove() failed for TypeSymlink for %s: %s", header.Linkname, err.Error())
				}
			}

			// Create symlink
			err = os.Symlink(header.Linkname, header.Name)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Symlink() failed for %s -> %s: %s", header.Linkname, header.Name, err.Error())
			}
		}

	case tar.TypeLink:
		logging.Tracef("Untarring link %s to %s", header.Name, header.Linkname)

		// Check if symlink already exists
		_, err := os.Lstat(header.Name)
		if err == nil {
			logging.Debugf("Link %s already exists", header.Name)
		}
		// Continue if symlink does not exist
		if os.IsNotExist(err) {
			logging.Debugf("Link %s does not exist", header.Name)

			// Create directories for symlink
			dir := filepath.Dir(header.Name)
			logging.Tracef("Creating directory %s", dir)
			err := os.MkdirAll(dir, 0755) // #nosec G301 -- Tools must be world readable
			if err != nil {
				return fmt.Errorf("ExtractTarGz: MkdirAll() failed for %s: %s", dir, err.Error())
			}

			_, err = os.Stat(header.Linkname)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("ExtractTarGz: Stat() failed for TypeLink for %s: %s", header.Linkname, err.Error())
				}
			} else {
				err = os.Remove(header.Linkname)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Remove() failed for TypeLink for %s: %s", header.Linkname, err.Error())
				}
			}

			// Create symlink
			err = os.Symlink(header.Linkname, header.Name)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Symlink() failed for %s -> %s: %s", header.Linkname, header.Name, err.Error())
			}
		}

	default:
		logging.Info.Printfln("Unknown: %s", header.Name)
	}
	return nil
}
