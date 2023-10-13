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

	log.Debugf("Checking if %s is inside %s\n", candidate, absTarget)
	realpath, err := filepath.Abs(filepath.Join(absTarget, candidate))
	if err != nil {
		return fmt.Errorf("ExtractTarGz: Abs() failed: %s", err.Error())
	}
	log.Debugf("Realpath of %s is %s\n", candidate, realpath)

	relpath, err := filepath.Rel(absTarget, realpath)
	if err != nil {
		return fmt.Errorf("ExtractTarGz: Rel() failed: %s", err.Error())
	}
	log.Debugf("Relpath of %s is %s\n", realpath, relpath)

	if strings.Contains(relpath, "..") {
		return fmt.Errorf("ExtractTarGz: symlink target contains '..': %s", relpath)
	}

	return nil
}

func ExtractTarGz(gzipStream io.Reader) error {
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

		// TODO: Check if this is really necessary
		//if strings.Contains(header.Name, "..") {
		//	return fmt.Errorf("ExtractTarGz: filename contains '..': %s", header.Name)
		//}

		switch header.Typeflag {
		case tar.TypeDir:
			log.Tracef("Creating directory %s\n", header.Name)
			_, err := os.Stat(header.Name)
			if err != nil {
				err := os.Mkdir(header.Name, 0755)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Mkdir() failed: %s", err.Error())
				}
			}

		case tar.TypeReg:
			log.Tracef("Untarring file %s\n", header.Name)
			outFile, err := os.Create(header.Name)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			mode := os.FileMode(header.Mode)
			err = outFile.Chmod(mode)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Chmod() failed: %s", err.Error())
			}
			outFile.Close()

		case tar.TypeSymlink:
			log.Tracef("Untarring symlink %s\n", header.Name)
			_, err := os.Stat(header.Name)
			if err == nil {
				log.Debugf("Symlink %s already exists\n", header.Name)
			}
			if os.IsNotExist(err) {
				log.Debugf("Symlink %s does not exist\n", header.Name)

				err = pathIsInsideTarget(target, header.Name)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: pathIsInsideTarget() failed for %s: %s", header.Name, err.Error())
				}

				absHeaderLinkname := header.Linkname
				if !filepath.IsAbs(header.Linkname) {
					absHeaderLinkname = filepath.Join(filepath.Dir(header.Name), header.Linkname)
				}
				log.Tracef("Absolute symlink target is %s\n", absHeaderLinkname)
				err = pathIsInsideTarget(target, absHeaderLinkname)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: pathIsInsideTarget() failed for %s: %s", absHeaderLinkname, err.Error())
				}

				err = os.Symlink(header.Linkname, header.Name)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Symlink() failed: %s", err.Error())
				}
			}

		case tar.TypeLink:
			log.Tracef("Untarring link %s\n", header.Name)
			_, err := os.Stat(header.Name)
			if err == nil {
				log.Debugf("Link %s already exists\n", header.Name)
			}
			if os.IsNotExist(err) {
				log.Debugf("Target of link %s does not exist\n", header.Name)
				os.Remove(header.Name)

				err = os.Link(header.Linkname, header.Name)
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
