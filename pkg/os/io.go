package os

import (
	"fmt"
	"io"
	"os"
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
