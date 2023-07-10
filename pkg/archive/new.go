package archive

import (
	// TODO: Check if https://github.com/mholt/archiver makes more sense
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

func ExtractTarGz(gzipStream io.Reader) error {
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
			outFile.Chmod(os.FileMode(header.Mode))
			outFile.Close()

		case tar.TypeSymlink:
			log.Tracef("Untarring symlink %s\n", header.Name)
			_, err := os.Stat(header.Name)
			if err == nil {
				log.Debugf("Symlink %s already exists\n", header.Name)
			}
			if os.IsNotExist(err) {
				log.Debugf("Target of symlink %s does not exist\n", header.Name)
				os.Remove(header.Name)

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
