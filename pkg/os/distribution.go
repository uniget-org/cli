package os

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func GetOsVendorFromFile(filename string) (string, error) {
	f, err := os.Open(filename) // #nosec G304 -- Prefix is the subdir uniget operates on
	if err != nil {
		return "", fmt.Errorf("cannot read /etc/os-release: %w", err)
	}
	//nolint:errcheck
	defer f.Close()

	s := bufio.NewScanner(f)
	return GetOsVendorFromScanner(s)
}

func GetOsVendorFromScanner(s *bufio.Scanner) (string, error) {
	var osVendor string
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

func GetOsVendor(prefix string) (string, error) {
	return GetOsVendorFromFile(prefix + "/etc/os-release")
}
