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
	v1, _ := goversion.NewVersion(s[i])
	v2, _ := goversion.NewVersion(s[j])
	return v1.LessThan(v2)
}
