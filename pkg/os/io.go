package os

import (
	"fmt"
	"io"
	"os"

	"gitlab.com/uniget-org/cli/pkg/logging"
	"golang.org/x/sys/unix"
)

func ConvertFileModeToString(mode int64) (string, error) {
	//result := fmt.Sprintf("%o ", mode)
	result := ""

	suid := false
	sgid := false
	sticky := false
	if mode > 0o7777 {
		return "", fmt.Errorf("unsupported mode %o", mode)
	}

	if mode >= 0o4000 {
		suid = true
		mode -= 0o4000
	}
	if mode >= 0o2000 {
		sgid = true
		mode -= 0o2000
	}
	if mode >= 0o1000 {
		sticky = true
		mode -= 0o1000
	}

	if mode >= 0o400 {
		result += "r"
		mode -= 0o400
	} else {
		result += "-"
	}
	if mode >= 0o200 {
		result += "w"
		mode -= 0o200
	} else {
		result += "-"
	}
	if mode >= 0o100 {
		result += "x"
		mode -= 0o100
	} else {
		result += "-"
	}
	if suid {
		result = result[0:len(result)-2] + "s"
	}

	if mode >= 0o40 {
		result += "r"
		mode -= 0o40
	} else {
		result += "-"
	}
	if mode >= 0o20 {
		result += "w"
		mode -= 0o20
	} else {
		result += "-"
	}
	if mode >= 0o10 {
		result += "x"
		mode -= 0o10
	} else {
		result += "-"
	}
	if sgid {
		result = result[0:len(result)-2] + "s"
	}

	if mode >= 0o4 {
		result += "r"
		mode -= 0o4
	} else {
		result += "-"
	}
	if mode >= 0o2 {
		result += "w"
		mode -= 0o2
	} else {
		result += "-"
	}
	if mode >= 0o1 {
		result += "x"
		//mode -= 0o1
	} else {
		result += "-"
	}
	if sticky {
		result = result[0:len(result)-2] + "s"
	}

	return result, nil
}

func SlurpFile(filePath string) ([]byte, error) {
	f, err := os.Open(filePath) // #nosec G304 -- Data input
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}
	defer func() {
		_ = f.Close()
	}()

	return io.ReadAll(f)
}

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src) // #nosec G304 - Low-level copy expects sanitation
	if err != nil {
		return fmt.Errorf("failed to open source file: %s", err)
	}
	defer func() {
		_ = srcFile.Close()
	}()

	dstFile, err := os.Create(dst) // #nosec G304 - Low-level copy expects sanitation
	if err != nil {
		return fmt.Errorf("failed to create destination file: %s", err)
	}
	defer func() {
		_ = dstFile.Close()
	}()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %s", err)
	}

	return nil
}

func CloneFile(src, dst string) (err error) {
	err = CopyFile(src, dst)
	if err != nil {
		return fmt.Errorf("failed to copy file: %s", err)
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
	if err != nil {
		return fmt.Errorf("failed to change file times: %s", err)
	}

	err = os.Chmod(dst, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to change file mode: %s", err)
	}

	return nil
}

func DirectoryExists(directory string) bool {
	_, err := os.Stat(directory)
	return err == nil
}

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func DirectoryIsWritable(directory string) bool {
	logging.Debugf("Checking if directory %s is writable", directory)
	return unix.Access(directory, unix.W_OK) == nil
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

func AssertDirectory(directory string) {
	logging.Debugf("Creating directory %s", directory)
	err := os.MkdirAll(directory, 0755) // #nosec G301 -- Directories will contain public information
	if err != nil {
		logging.Error.Printfln("Error creating directory %s: %s", directory, err)
		os.Exit(1)
	}
}
