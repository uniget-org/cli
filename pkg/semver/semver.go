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
	v1, err := goversion.NewVersion(s[i])
	if err != nil {
		panic(err)
	}
	v2, err := goversion.NewVersion(s[j])
	if err != nil {
		panic(err)
	}
	return v1.LessThan(v2)
}
