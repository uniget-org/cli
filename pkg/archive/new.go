package archive

import (
	"archive/tar"
    "compress/gzip"
	"fmt"
	"io"
	"os"
)

// TODO: Check if https://github.com/mholt/archiver makes more sense
func ExtractTarGz(gzipStream io.Reader) error {
    uncompressedStream, err := gzip.NewReader(gzipStream)
    if err != nil {
        return fmt.Errorf("ExtractTarGz: NewReader failed")
    }

    tarReader := tar.NewReader(uncompressedStream)

    for true {
        header, err := tarReader.Next()

        if err == io.EOF {
            break
        }

        if err != nil {
            return fmt.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
        }

        switch header.Typeflag {
        case tar.TypeDir:
            _, err := os.Stat(header.Name)
			if err != nil {
                err := os.Mkdir(header.Name, 0755)
				if err != nil {
					return fmt.Errorf("ExtractTarGz: Mkdir() failed: %s", err.Error())
				}
			}

        case tar.TypeReg:
            outFile, err := os.Create(header.Name)
            if err != nil {
                return fmt.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
            }
            if _, err := io.Copy(outFile, tarReader); err != nil {
                return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
            }
			outFile.Chmod(os.FileMode(header.Mode))
            outFile.Close()

        default:
            return fmt.Errorf("ExtractTarGz: uknown type: %s in %s", header.Typeflag, header.Name)
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
    for true {
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

        default:
            return nil, fmt.Errorf("ListTarGz: uknown type: %s in %s", header.Typeflag, header.Name)
        }

    }

	return result, nil
}