package os

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func GetOsVendor(prefix string) (string, error) {
	f, err := os.Open(prefix + "/etc/os-release") // #nosec G304 -- Prefix is the subdir uniget operates on
	if err != nil {
		return "", fmt.Errorf("cannot read /etc/os-release: %w", err)
	}
	defer f.Close()

	var osVendor string
	s := bufio.NewScanner(f)
	for s.Scan() {
		re, err := regexp.Compile(`^ID=(.*)$`)
		if err != nil {
			return "", fmt.Errorf("cannot compile regexp: %w", err)
		}
		m := re.FindStringSubmatch(s.Text())
		if m != nil {
			osVendor = strings.Trim(m[1], `"`)
		}
	}

	return osVendor, nil
}
