package archive

import (
	// TODO: Check if https://github.com/mholt/archiver makes more sense
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func pathIsInsideTarget(target string, candidate string) error {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("ExtractTarGz: Abs() failed: %s", err.Error())
	}

	log.Debugf("Checking if %s works\n", filepath.Join(absTarget, candidate))
	cleanPath := filepath.Clean(filepath.Join(absTarget, candidate))
	log.Debugf("Cleaned path is %s\n", cleanPath)

	realPath, err := filepath.EvalSymlinks(cleanPath)
	if os.IsNotExist(err) {
		log.Tracef("Path does not exist (yet): %s\n", cleanPath)
		realPath = cleanPath

	} else if err != nil {
		return fmt.Errorf("ExtractTarGz: EvalSymlinks() failed: %s", err.Error())
	}
	log.Debugf("Realpath of %s is %s\n", candidate, realPath)

	relPath, err := filepath.Rel(absTarget, realPath)
	if err != nil {
		return fmt.Errorf("ExtractTarGz: Rel() failed: %s", err.Error())
	}
	log.Debugf("Relative path of %s is %s\n", realPath, relPath)

	if strings.Contains(relPath, "..") {
		return fmt.Errorf("ExtractTarGz: symlink target contains '..': %s", relPath)
	}

	return nil
}

func ExtractTarGz(gzipStream io.Reader, func patchPath(path string) string) error {
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
		}

		if err != nil {
			return fmt.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		if strings.Contains(header.Name, "..") {
			return fmt.Errorf("ExtractTarGz: filename contains '..': %s", header.Name)
		}

		log.Tracef("Processing %s\n", header.Name)
		strippedHeaderName := strings.TrimPrefix(header.Name, "usr/local/")
		log.Tracef("  Stripped name is %s\n", strippedHeaderName)
		if len(strippedHeaderName) == 0 {
			log.Tracef("  Skipping\n")
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			log.Tracef("Creating directory %s\n", strippedHeaderName)
			_, err := os.Stat(strippedHeaderName)
			if err != nil {
				err := os.Mkdir(strippedHeaderName, 0755) // #nosec G301 -- Tools must be world readable
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Mkdir() failed: %s", err.Error())
				}
			}

		case tar.TypeReg:
			log.Tracef("Untarring file %s\n", strippedHeaderName)
			outFile, err := os.Create(strippedHeaderName)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
			} // #nosec G110 -- Tool images are a trusted source
			mode := os.FileMode(header.Mode)
			err = outFile.Chmod(mode)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Chmod() failed: %s", err.Error())
			}
			err = outFile.Close()
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Failed to close %s: %s", strippedHeaderName, err.Error())
			}

		case tar.TypeSymlink:
			log.Tracef("Untarring symlink %s\n", strippedHeaderName)
			_, err := os.Stat(strippedHeaderName)
			if err == nil {
				log.Debugf("Symlink %s already exists\n", strippedHeaderName)
			}
			if os.IsNotExist(err) {
				log.Debugf("Symlink %s does not exist\n", strippedHeaderName)

				absHeaderLinkname := header.Linkname
				if !filepath.IsAbs(header.Linkname) {
					absHeaderLinkname = filepath.Join(filepath.Dir(strippedHeaderName), header.Linkname) // #nosec G305 -- Following code prevents traversal
				}
				log.Tracef("Absolute symlink target is %s\n", absHeaderLinkname)
				err = pathIsInsideTarget(target, absHeaderLinkname)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: pathIsInsideTarget() failed for %s: %s", absHeaderLinkname, err.Error())
				}

				err = pathIsInsideTarget(target, strippedHeaderName)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: pathIsInsideTarget() failed for %s: %s", strippedHeaderName, err.Error())
				}

				err = os.Symlink(header.Linkname, strippedHeaderName)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Symlink() failed: %s", err.Error())
				}
			}

		case tar.TypeLink:
			log.Tracef("Untarring link %s\n", strippedHeaderName)
			_, err := os.Stat(strippedHeaderName)
			if err == nil {
				log.Debugf("Link %s already exists\n", strippedHeaderName)
			}
			if os.IsNotExist(err) {
				log.Debugf("Target of link %s does not exist\n", strippedHeaderName)
				err = os.Remove(strippedHeaderName)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Remove() failed for TypeLink: %s", err.Error())
				}

				err = os.Link(header.Linkname, strippedHeaderName)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Link() failed: %s", err.Error())
				}
			}

		default:
			return fmt.Errorf("ExtractTarGz: unknown type for entry %s: %b", header.Name, header.Typeflag)
		}

	}

	return nil
}

func ListTarGz(gzipStream io.Reader) ([]string, error) {
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

		switch header.Typeflag {
		case tar.TypeDir:

		case tar.TypeReg:
			result = append(result, header.Name)

		case tar.TypeSymlink:
			result = append(result, header.Name+" -> "+header.Linkname)

		case tar.TypeLink:
			result = append(result, header.Name+" -> "+header.Linkname)

		case tar.TypeChar:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeChar", header.Name)

		case tar.TypeBlock:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeBlock", header.Name)

		case tar.TypeFifo:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeFifo", header.Name)

		case tar.TypeCont:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeCont", header.Name)

		case tar.TypeXHeader:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeXHeader", header.Name)

		case tar.TypeXGlobalHeader:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeXGlobalHeader", header.Name)

		case tar.TypeGNULongLink:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeGNULongLink", header.Name)

		case tar.TypeGNULongName:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeGNULongName", header.Name)

		case tar.TypeGNUSparse:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: TypeGNUSparse", header.Name)

		default:
			return nil, fmt.Errorf("ListTarGz: unknown type in entry %s: %b", header.Name, header.Typeflag)
		}

	}

	return result, nil
}
