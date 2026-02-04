package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	regex = `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
)

type SemVer struct {
	Major         int
	Minor         int
	Patch         int
	PrereleaseTag string
	Prerelease    int
}

type SemVerByVersion []SemVer

func (s SemVerByVersion) Len() int {
	return len(s)
}

func (s SemVerByVersion) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SemVerByVersion) Less(i, j int) bool {
	v1 := s[i]
	v2 := s[j]
	return v1.LessThan(&v2)
}

func NewSemVer(version string) (*SemVer, error) {
	re := regexp.MustCompile(regex)
	matches := re.FindStringSubmatch(version)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid semver string: %s", version)
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	prereleaseSegments := strings.Split(matches[4], ".")
	prereleaseTag := prereleaseSegments[0]
	prerelease, _ := strconv.Atoi(prereleaseSegments[1])

	return &SemVer{
		Major:         major,
		Minor:         minor,
		Patch:         patch,
		PrereleaseTag: prereleaseTag,
		Prerelease:    prerelease,
	}, nil
}

func (s *SemVer) String() string {
	return fmt.Sprintf("%d.%d.%d-%s.%d", s.Major, s.Minor, s.Patch, s.PrereleaseTag, s.Prerelease)
}

func (s *SemVer) GetMajor() int {
	return s.Major
}

func (s *SemVer) GetMinor() int {
	return s.Minor
}

func (s *SemVer) GetPatch() int {
	return s.Patch
}

func (s *SemVer) GetPrereleaseTag() string {
	return s.PrereleaseTag
}

func (s *SemVer) GetPrerelease() int {
	return s.Prerelease
}

func (s *SemVer) LessThan(other *SemVer) bool {
	if s.Major != other.Major {
		return s.Major < other.Major
	}
	if s.Minor != other.Minor {
		return s.Minor < other.Minor
	}
	if s.Patch != other.Patch {
		return s.Patch < other.Patch
	}
	if s.PrereleaseTag != other.PrereleaseTag {
		return s.PrereleaseTag < other.PrereleaseTag
	}
	return s.Prerelease < other.Prerelease
}

func (s *SemVer) BumpMajor() *SemVer {
	return &SemVer{
		Major:         s.Major + 1,
		Minor:         s.Minor,
		Patch:         s.Patch,
		PrereleaseTag: s.PrereleaseTag,
		Prerelease:    s.Prerelease,
	}
}

func (s *SemVer) BumpMinor() *SemVer {
	return &SemVer{
		Major:         s.Major,
		Minor:         s.Minor + 1,
		Patch:         s.Patch,
		PrereleaseTag: s.PrereleaseTag,
		Prerelease:    s.Prerelease,
	}
}

func (s *SemVer) BumpPatch() *SemVer {
	return &SemVer{
		Major:         s.Major,
		Minor:         s.Minor,
		Patch:         s.Patch + 1,
		PrereleaseTag: s.PrereleaseTag,
		Prerelease:    s.Prerelease,
	}
}

func (s *SemVer) BumpPrerelease() *SemVer {
	return &SemVer{
		Major:         s.Major,
		Minor:         s.Minor,
		Patch:         s.Patch,
		PrereleaseTag: s.PrereleaseTag,
		Prerelease:    s.Prerelease + 1,
	}
}
