package archive

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/safearchive/tar"
	"github.com/google/safeopen"

	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
)

func ProcessTarContents(reader io.ReadCloser, callback func(tar *tar.Reader, header *tar.Header) error) error {
	tarReader := tar.NewReader(reader)
	//nolint:errcheck
	defer reader.Close()

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
		mode, err := myos.ConvertFileModeToString(header.Mode)
		if err != nil {
			return fmt.Errorf("unable to convert mode: %s", err)
		}
		//nolint:errcheck
		fmt.Fprintf(logging.OutputWriter, "-%s %s\n", mode, header.Name)
	case tar.TypeSymlink, tar.TypeLink:
		mode, err := myos.ConvertFileModeToString(header.Mode)
		if err != nil {
			return fmt.Errorf("unable to convert mode: %s", err)
		}
		//nolint:errcheck
		fmt.Fprintf(logging.OutputWriter, "l%s %s -> %s\n", mode, header.Name, header.Linkname)
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

	name := filepath.Clean(header.Name)
	if !filepath.IsLocal(name) {
		return fmt.Errorf("refusing to extract %q outside of %q", header.Name, workDir)
	}

	root, err := os.OpenRoot(workDir)
	if err != nil {
		return fmt.Errorf("unable to open working directory %s: %s", workDir, err)
	}
	//nolint:errcheck
	defer root.Close()

	switch header.Typeflag {

	case tar.TypeReg:
		logging.Debugf("File: %s", header.Name)

		err = ExtractFileFromTar(workDir, name, reader, header)
		if err != nil {
			return fmt.Errorf("ExtractTarGz: ExtractFileFromTar() failed for %s: %s", name, err.Error())
		}

	case tar.TypeSymlink, tar.TypeLink:
		logging.Tracef("Untarring (sym)link %s -> %s", header.Name, header.Linkname)

		err = myos.ValidateLinkTarget(workDir, name, header)
		if err != nil {
			return err
		}

		// Check if (sym)link already exists
		_, err := root.Lstat(name)
		if err == nil {
			logging.Debugf("Symlink %s already exists", header.Name)
			return nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("unable to check (sym)link %s: %s", header.Name, err)
		}

		// Create directories for (sym)link
		dir := filepath.Dir(name)
		logging.Tracef("Creating directory %s", dir)
		if dir != "." {
			err = root.MkdirAll(dir, 0755) // #nosec G301 -- Tools must be world readable
			if err != nil {
				return fmt.Errorf("ExtractTarGz: MkdirAll() failed for %s: %s", dir, err.Error())
			}
		}

		// Create (sym)link
		if header.Typeflag == tar.TypeLink {
			linkName := filepath.Clean(header.Linkname)
			err = root.Link(linkName, name)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Link() failed for %s -> %s: %s", linkName, name, err.Error())
			}
		} else {
			err = root.Symlink(header.Linkname, name)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Symlink() failed for %s -> %s: %s", header.Linkname, name, err.Error())
			}
		}

	default:
		logging.Info.Printfln("Unknown: %s", header.Name)
	}
	return nil
}

func ExtractFileFromTar(workDir string, path string, reader *tar.Reader, header *tar.Header) error {
	cleanPath := filepath.Clean(path)
	if !filepath.IsLocal(cleanPath) {
		return fmt.Errorf("refusing to extract %q outside of %q", path, workDir)
	}

	// Only permission bits are applied. Special bits from the archive
	// (setuid, setgid, sticky) are deliberately dropped so extraction
	// cannot install setuid/setgid binaries. Validating the mode before
	// creating the file rejects manipulated archives early.
	if header.Mode < 0 {
		return fmt.Errorf("ExtractTarGz: invalid mode %d for %s", header.Mode, path)
	}
	if header.Mode&^int64(os.ModePerm) != 0 {
		logging.Warning.Printfln("Ignoring non-permission mode bits for %s", path)
	}
	mode := os.FileMode(header.Mode & int64(os.ModePerm))

	err := myos.CreateSubdirectoriesForPath(workDir, cleanPath)
	if err != nil {
		return fmt.Errorf("ExtractTarGz: CreateSubdirectoriesForPath() failed for %s in %s: %s", path, workDir, err.Error())
	}
	outFile, err := safeopen.CreateBeneath(workDir, cleanPath)
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

	err = outFile.Chmod(mode)
	if err != nil {
		return fmt.Errorf("ExtractTarGz: Chmod() failed for %s: %s", path, err.Error())
	}

	return nil
}
