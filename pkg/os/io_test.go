package os

import (
	"testing"
)

var (
	stringForMode = map[int64]string{
		0o777: "rwxrwxrwx",
		0o755: "rwxr-xr-x",
		0o644: "rw-r--r--",
	}
)

func TestFileMode(t *testing.T) {
	for mode, expected := range stringForMode {
		modeString, err := ConvertFileModeToString(mode)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
			continue
		}
		if modeString != expected {
			t.Errorf("expected %s, got %s", expected, modeString)
		}
	}
}
