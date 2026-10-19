package semver

import (
	goversion "github.com/hashicorp/go-version"
)

type ByVersion []string

func (s ByVersion) Len() int {
	return len(s)
}

func (s ByVersion) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByVersion) Less(i, j int) bool {
	v1, err1 := goversion.NewVersion(s[i])
	v2, err2 := goversion.NewVersion(s[j])

	switch {
	case err1 != nil && err2 != nil:
		// Neither version can be parsed. Fall back to comparing strings.
		return s[i] < s[j]
	case err1 != nil:
		// Unparseable versions sort after parseable versions.
		return false
	case err2 != nil:
		// Unparseable versions sort after parseable versions.
		return true
	default:
		return v1.LessThan(v2)
	}
}
