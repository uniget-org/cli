package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/uniget-org/cli/pkg/logging"
)

func pathIsInsideTarget(target string, candidate string) error {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("pathIsInsideTarget(): Abs() failed: %s", err.Error())
	}

	logging.Debugf("Checking if %s works", filepath.Join(absTarget, candidate))
	cleanPath := filepath.Clean(filepath.Join(absTarget, candidate))
	logging.Debugf("Cleaned path is %s", cleanPath)

	realPath, err := filepath.EvalSymlinks(cleanPath)
	if os.IsNotExist(err) {
		logging.Tracef("Path does not exist (yet): %s", cleanPath)
		realPath = cleanPath

	} else if err != nil {
		return fmt.Errorf("pathIsInsideTarget(): EvalSymlinks() failed: %s", err.Error())
	}
	logging.Debugf("Realpath of %s is %s", candidate, realPath)

	relPath, err := filepath.Rel(absTarget, realPath)
	if err != nil {
		return fmt.Errorf("pathIsInsideTarget(): Rel() failed: %s", err.Error())
	}
	logging.Debugf("Relative path of %s is %s", realPath, relPath)

	if strings.Contains(relPath, "..") {
		return fmt.Errorf("pathIsInsideTarget(): symlink target contains '..': %s", relPath)
	}

	return nil
}

func Gunzip(layer []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(layer))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %s", err)
	}
	defer reader.Close()

	buffer, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read gzip: %s", err)
	}

	return buffer, nil
}

func ExtractTarGz(gzipStream io.Reader, patchPath func(path string) string, patchFile func(path string)) error {
	target := "."

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		err = ExtractItemOld(tarReader, header, patchPath, target, patchFile)
		if err != nil {
			return fmt.Errorf("ExtractTarGz: ExtractItem() failed: %s", err.Error())
		}
	}

	return nil
}

func ExtractItemOld(tarReader *tar.Reader, header *tar.Header, patchPath func(path string) string, target string, patchFile func(path string)) error {
	// Directories will be created when files/(sym)links are unpacked
	if header.Typeflag == tar.TypeDir {
		return nil
	}

	// Prevent path traversal attacks using ..
	if strings.Contains(header.Name, "..") {
		return fmt.Errorf("ExtractTarGz: filename contains '..': %s", header.Name)
	}

	logging.Tracef("Processing %s", header.Name)

	// Use callback function to patch path
	fixedHeaderName := patchPath(header.Name)
	logging.Tracef("  Stripped name is %s", fixedHeaderName)
	if len(fixedHeaderName) == 0 {
		logging.Tracef("  Skipping")
		return nil
	}

	switch header.Typeflag {

	// Unpack file
	case tar.TypeReg:
		logging.Tracef("Untarring file %s", fixedHeaderName)

		// Prevent path traversal attacks using absolute paths
		cleanFixedHeaderName := filepath.Clean(fixedHeaderName)
		if strings.HasPrefix(cleanFixedHeaderName, "/") {
			return fmt.Errorf("ExtractTarGz: filename starts with '/': %s", cleanFixedHeaderName)
		}

		// Create directories for file
		dir := filepath.Dir(fixedHeaderName)
		err := os.MkdirAll(dir, 0755) // #nosec G301 -- Tools must be world readable
		if err != nil {
			return fmt.Errorf("ExtractTarGz: MkdirAll() failed: %s", err.Error())
		}

		// Create file
		outFile, err := os.Create(fixedHeaderName)
		if err != nil {
			return fmt.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
		}
		defer outFile.Close()
		// Write contents
		if _, err := io.Copy(outFile, tarReader); err != nil /* #nosec G110 -- Tool images are a trusted source */ {
			return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
		}
		// Set permissions
		mode := os.FileMode(header.Mode) // #nosec G115 -- Must be addressed upstream
		Setuid := mode &^ 0777
		if (mode & Setuid) != 0 {
			logging.Warning.Printfln("Setuid bit cannot be set for %s", fixedHeaderName)
		}
		err = outFile.Chmod(mode)
		if err != nil {
			return fmt.Errorf("ExtractTarGz: Chmod() failed: %s", err.Error())
		}

		// Callback for patching file
		patchFile(fixedHeaderName)

	// Unpack symlink
	case tar.TypeSymlink:
		logging.Tracef("Untarring symlink %s", fixedHeaderName)

		// Check if symlink already exists
		_, err := os.Lstat(fixedHeaderName)
		if err == nil {
			logging.Debugf("Symlink %s already exists", fixedHeaderName)
		}
		// Continue if symlink does not exist
		if os.IsNotExist(err) {
			logging.Debugf("Symlink %s does not exist", fixedHeaderName)

			// Prevent path traversal attacks for symlink source and target
			absHeaderLinkname := header.Linkname
			if !filepath.IsAbs(header.Linkname) {
				absHeaderLinkname = filepath.Join(filepath.Dir(fixedHeaderName), header.Linkname) // #nosec G305 -- Following code prevents traversal
			}
			logging.Tracef("Absolute symlink target is %s", absHeaderLinkname)
			err = pathIsInsideTarget(target, absHeaderLinkname)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: pathIsInsideTarget() failed for %s: %s", absHeaderLinkname, err.Error())
			}
			err = pathIsInsideTarget(target, fixedHeaderName)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: pathIsInsideTarget() failed for %s: %s", fixedHeaderName, err.Error())
			}

			// Create directories for symlink
			dir := filepath.Dir(fixedHeaderName)
			logging.Tracef("Creating directory %s", dir)
			err := os.MkdirAll(dir, 0755) // #nosec G301 -- Tools must be world readable
			if err != nil {
				return fmt.Errorf("ExtractTarGz: MkdirAll() failed: %s", err.Error())
			}

			_, err = os.Stat(header.Linkname)
			if err != nil {
				if os.IsNotExist(err) {
					// Symlink target does not exist
				} else {
					return fmt.Errorf("ExtractTarGz: Stat() failed for TypeSymlink: %s", err.Error())
				}
			} else {
				err = os.Remove(header.Linkname)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Remove() failed for TypeSymlink: %s", err.Error())
				}
			}

			// Create symlink
			err = os.Symlink(header.Linkname, fixedHeaderName)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Symlink() failed: %s", err.Error())
			}
		}

	// Unpack hardlink
	case tar.TypeLink:
		logging.Tracef("Untarring link %s", fixedHeaderName)

		// Check if link already exists
		_, err := os.Stat(fixedHeaderName)
		if err == nil {
			logging.Debugf("Link %s already exists", fixedHeaderName)
		}
		// Continue if link does not exist
		if os.IsNotExist(err) {
			logging.Debugf("Target of link %s does not exist", fixedHeaderName)

			// Remove existing link
			err = os.Remove(fixedHeaderName)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Remove() failed for TypeLink: %s", err.Error())
			}

			// Create link
			err = os.Link(header.Linkname, fixedHeaderName)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Link() failed: %s", err.Error())
			}
		}

	// Fail on unhandled types
	default:
		return fmt.Errorf("ExtractTarGz: unknown type for entry %s: %b", header.Name, header.Typeflag)
	}

	return nil
}

func ListTarGz(gzipStream io.Reader, patchPath func(path string) string) ([]string, error) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return nil, fmt.Errorf("ListTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	result := []string{}
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("ListTarGz: Next() failed: %s", err.Error())
		}

		logging.Tracef("Processing %s", header.Name)
		fixedHeaderName := patchPath(header.Name)
		logging.Tracef("  Stripped name is %s", fixedHeaderName)
		if len(fixedHeaderName) == 0 {
			logging.Tracef("  Skipping")
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:

		case tar.TypeReg:
			result = append(result, fixedHeaderName)

		case tar.TypeSymlink:
			result = append(result, fixedHeaderName+" -> "+header.Linkname)

		case tar.TypeLink:
			result = append(result, fixedHeaderName+" -> "+header.Linkname)

		case tar.TypeChar:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeChar", fixedHeaderName)

		case tar.TypeBlock:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeBlock", fixedHeaderName)

		case tar.TypeFifo:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeFifo", fixedHeaderName)

		case tar.TypeCont:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeCont", fixedHeaderName)

		case tar.TypeXHeader:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeXHeader", fixedHeaderName)

		case tar.TypeXGlobalHeader:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeXGlobalHeader", fixedHeaderName)

		case tar.TypeGNULongLink:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeGNULongLink", fixedHeaderName)

		case tar.TypeGNULongName:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeGNULongName", fixedHeaderName)

		case tar.TypeGNUSparse:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeGNUSparse", fixedHeaderName)

		default:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: %b", fixedHeaderName, header.Typeflag)
		}

	}

	return result, nil
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
		fmt.Fprintf(logging.OutputWriter, "%s\n", header.Name)
	case tar.TypeSymlink:
		fmt.Fprintf(logging.OutputWriter, "%s -> %s\n", header.Name, header.Linkname)
	case tar.TypeLink:
		fmt.Fprintf(logging.OutputWriter, "%s -> %s\n", header.Name, header.Linkname)
	default:
		fmt.Fprintf(logging.ErrorWriter, "Unknown: %s\n", header.Name)
	}
	return nil
}

func CallbackExtractTarItem(reader *tar.Reader, header *tar.Header) error {
	if header.Typeflag == tar.TypeDir {
		return nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory")
	}
	logging.Info.Printfln("Current directory: %s", dir)

	switch header.Typeflag {

	case tar.TypeReg:
		logging.Info.Printfln("File: %s", header.Name)

		path := filepath.Clean(header.Name)

		dir := filepath.Dir(path)
		err := os.MkdirAll(dir, 0755) // #nosec G301 -- Tools must be world readable
		if err != nil {
			return fmt.Errorf("ExtractTarGz: MkdirAll() failed: %s", err.Error())
		}

		outFile, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
		}
		defer outFile.Close()
		if _, err := io.Copy(outFile, reader); err != nil {
			return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
		}

		mode := os.FileMode(header.Mode) // #nosec G115 -- Must be addressed upstream
		Setuid := mode &^ 0777
		if (mode & Setuid) != 0 {
			logging.Warning.Printfln("Setuid bit cannot be set for %s", path)
		}
		err = outFile.Chmod(mode)
		if err != nil {
			return fmt.Errorf("ExtractTarGz: Chmod() failed: %s", err.Error())
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
				return fmt.Errorf("ExtractTarGz: MkdirAll() failed: %s", err.Error())
			}

			_, err = os.Stat(header.Linkname)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("ExtractTarGz: Stat() failed for TypeSymlink: %s", err.Error())
				}
			} else {
				err = os.Remove(header.Linkname)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Remove() failed for TypeSymlink: %s", err.Error())
				}
			}

			// Create symlink
			err = os.Symlink(header.Linkname, header.Name)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Symlink() failed: %s", err.Error())
			}
		}

	case tar.TypeLink:
		logging.Tracef("Untarring link %s", header.Name)

		// Check if link already exists
		_, err := os.Stat(header.Name)
		if err == nil {
			logging.Debugf("Link %s already exists", header.Name)
		}
		// Continue if link does not exist
		if os.IsNotExist(err) {
			logging.Debugf("Target of link %s does not exist", header.Name)

			// Remove existing link
			err = os.Remove(header.Name)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Remove() failed for TypeLink: %s", err.Error())
			}

			// Create link
			err = os.Link(header.Linkname, header.Name)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Link() failed: %s", err.Error())
			}
		}

	default:
		logging.Info.Printfln("Unknown: %s", header.Name)
	}
	return nil
}
