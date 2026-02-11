package semver_test

import (
	"sort"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/semver"
)

func TestNewSemVer(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    semver.SemVer
		wantErr bool
	}{
		{
			name:    "full",
			version: "1.2.3-alpha.4",
			want: semver.SemVer{
				Major:         1,
				Minor:         2,
				Patch:         3,
				PrereleaseTag: "alpha",
				Prerelease:    4,
			},
			wantErr: false,
		},
		{
			name:    "simple",
			version: "1.2.3",
			want: semver.SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			wantErr: false,
		},
		{
			name:    "empty",
			version: "",
			want:    semver.SemVer{},
			wantErr: true,
		},
		{
			name:    "short",
			version: "1.2",
			want:    semver.SemVer{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := semver.NewSemVer(tt.version)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewSemVer() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewSemVer() succeeded unexpectedly")
			}
			if *got != tt.want {
				t.Errorf("NewSemVer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_String(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "full",
			version: "1.2.3-alpha.4",
		},
		{
			name:    "simple",
			version: "1.2.3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := semver.NewSemVer(tt.version)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := s.String()
			if got != tt.version {
				t.Errorf("String() = %v, version %v", got, tt.version)
			}
		})
	}
}

func TestSemVer_GetMajor(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    int
	}{
		{
			name:    "full",
			version: "1.2.3-alpha.4",
			want:    1,
		},
		{
			name:    "simple",
			version: "1.2.3",
			want:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := semver.NewSemVer(tt.version)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := s.GetMajor()
			if got != tt.want {
				t.Errorf("GetMajor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_GetMinor(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    int
	}{
		{
			name:    "full",
			version: "1.2.3-alpha.4",
			want:    2,
		},
		{
			name:    "simple",
			version: "1.2.3",
			want:    2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := semver.NewSemVer(tt.version)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := s.GetMinor()
			if got != tt.want {
				t.Errorf("GetMinor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_GetPatch(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    int
	}{
		{
			name:    "full",
			version: "1.2.3-alpha.4",
			want:    3,
		},
		{
			name:    "simple",
			version: "1.2.3",
			want:    3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := semver.NewSemVer(tt.version)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := s.GetPatch()
			if got != tt.want {
				t.Errorf("GetPatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_GetPrereleaseTag(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "full",
			version: "1.2.3-alpha.4",
			want:    "alpha",
		},
		{
			name:    "simple",
			version: "1.2.3",
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := semver.NewSemVer(tt.version)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := s.GetPrereleaseTag()
			if got != tt.want {
				t.Errorf("GetPrereleaseTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_GetPrerelease(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    int
	}{
		{
			name:    "full",
			version: "1.2.3-alpha.4",
			want:    4,
		},
		{
			name:    "simple",
			version: "1.2.3",
			want:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := semver.NewSemVer(tt.version)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := s.GetPrerelease()
			if got != tt.want {
				t.Errorf("GetPrerelease() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_LessThan(t *testing.T) {
	tests := []struct {
		name string
		v1   semver.SemVer
		v2   semver.SemVer
		want bool
	}{
		{
			name: "equal",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 3},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 3},
			want: false,
		},
		{
			name: "major1",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 3},
			v2:   semver.SemVer{Major: 2, Minor: 3, Patch: 4},
			want: true,
		},
		{
			name: "major2",
			v1:   semver.SemVer{Major: 2, Minor: 3, Patch: 4},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 3},
			want: false,
		},
		{
			name: "minor1",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 3},
			v2:   semver.SemVer{Major: 1, Minor: 3, Patch: 4},
			want: true,
		},
		{
			name: "minor2",
			v1:   semver.SemVer{Major: 1, Minor: 3, Patch: 4},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 3},
			want: false,
		},
		{
			name: "patch1",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 3},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 4},
			want: true,
		},
		{
			name: "patch2",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 4},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 3},
			want: false,
		},
		{
			name: "prerelease1",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "alpha", Prerelease: 1},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "alpha", Prerelease: 2},
			want: true,
		},
		{
			name: "prerelease2",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "alpha", Prerelease: 2},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "alpha", Prerelease: 1},
			want: false,
		},
		{
			name: "prerelease3",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "alpha", Prerelease: 1},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "beta", Prerelease: 1},
			want: true,
		},
		{
			name: "prerelease4",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "beta", Prerelease: 1},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "alpha", Prerelease: 1},
			want: false,
		},
		{
			name: "prerelease3",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "alpha", Prerelease: 1},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "gamma", Prerelease: 1},
			want: true,
		},
		{
			name: "prerelease4",
			v1:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "gamma", Prerelease: 1},
			v2:   semver.SemVer{Major: 1, Minor: 2, Patch: 3, PrereleaseTag: "alpha", Prerelease: 1},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.LessThan(&tt.v2)
			if got != tt.want {
				t.Errorf("LessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_BumpMajor(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "simple",
			version: "1.2.3",
			want:    "2.0.0",
		},
		{
			name:    "full",
			version: "1.2.3-rc.4",
			want:    "2.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := semver.NewSemVer(tt.version)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := s.BumpMajor()
			if got.String() != tt.want {
				t.Errorf("BumpMajor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_BumpMinor(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "simple",
			version: "1.2.3",
			want:    "1.3.0",
		},
		{
			name:    "full",
			version: "1.2.3-rc.4",
			want:    "1.3.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := semver.NewSemVer(tt.version)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := s.BumpMinor()
			if got.String() != tt.want {
				t.Errorf("BumpMinor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_BumpPatch(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "simple",
			version: "1.2.3",
			want:    "1.2.4",
		},
		{
			name:    "full",
			version: "1.2.3-rc.4",
			want:    "1.2.4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := semver.NewSemVer(tt.version)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := s.BumpPatch()
			if got.String() != tt.want {
				t.Errorf("BumpPatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_BumpPrerelease(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "simple",
			version: "1.2.3",
			want:    "1.2.3",
		},
		{
			name:    "full",
			version: "1.2.3-rc.4",
			want:    "1.2.3-rc.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := semver.NewSemVer(tt.version)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := s.BumpPrerelease()
			if got.String() != tt.want {
				t.Errorf("BumpPrerelease() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVerByVersion_Len(t *testing.T) {
	tests := []struct {
		name     string
		versions semver.SemVerByVersion
		want     int
	}{
		{
			name:     "empty",
			versions: []semver.SemVer{},
			want:     0,
		},
		{
			name: "len1",
			versions: []semver.SemVer{
				{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
			want: 1,
		},
		{
			name: "len5",
			versions: []semver.SemVer{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.versions.Len()
			if got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVerByVersion_Swap(t *testing.T) {
	tests := []struct {
		name     string
		versions semver.SemVerByVersion
		i        int
		j        int
	}{
		{
			name: "swap_0_1",
			versions: []semver.SemVer{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			i: 0,
			j: 1,
		},
		{
			name: "swap_1_1",
			versions: []semver.SemVer{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			i: 1,
			j: 1,
		},
		{
			name: "swap_3_1",
			versions: []semver.SemVer{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			i: 3,
			j: 1,
		},
		{
			name: "swap_0_4",
			versions: []semver.SemVer{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			i: 0,
			j: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iPre := tt.versions[tt.i]
			jPre := tt.versions[tt.j]
			tt.versions.Swap(tt.i, tt.j)
			if tt.versions[tt.i] != jPre || tt.versions[tt.j] != iPre {
				t.Errorf("Swap(%d, %d) = (%v, %v), want (%v, %v)", tt.i, tt.j, tt.versions[tt.i], tt.versions[tt.j], jPre, iPre)
			}
		})
	}
}

func TestSemVerByVersion_Less(t *testing.T) {
	tests := []struct {
		name     string
		versions semver.SemVerByVersion
		i        int
		j        int
		want     bool
	}{
		{
			name: "less_0_1",
			versions: []semver.SemVer{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			i:    0,
			j:    1,
			want: true,
		},
		{
			name: "less_3_1",
			versions: []semver.SemVer{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			i:    3,
			j:    1,
			want: true,
		},
		{
			name: "less_1_1",
			versions: []semver.SemVer{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			i:    1,
			j:    1,
			want: false,
		},
		{
			name: "less_0_4",
			versions: []semver.SemVer{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			i:    0,
			j:    4,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.versions.Less(tt.i, tt.j)
			if got != tt.want {
				t.Errorf("Less() = %v, want %v (compared %v, %v)", got, tt.want, tt.versions[tt.i], tt.versions[tt.j])
			}
		})
	}
}

func TestSemVerByVersion_Sort(t *testing.T) {
	tests := []struct {
		name     string
		versions semver.SemVerByVersion
		want     semver.SemVerByVersion
	}{
		{
			name: "test1",
			versions: []semver.SemVer{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			want: []semver.SemVer{
				{Major: 0, Minor: 1, Patch: 2},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
			},
		},
		{
			name: "sorted",
			versions: []semver.SemVer{
				{Major: 0, Minor: 1, Patch: 2},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
			},
			want: []semver.SemVer{
				{Major: 0, Minor: 1, Patch: 2},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
			},
		},
		{
			name: "reverse",
			versions: []semver.SemVer{
				{Major: 1, Minor: 3, Patch: 1},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 0, Minor: 1, Patch: 2},
			},
			want: []semver.SemVer{
				{Major: 0, Minor: 1, Patch: 2},
				{Major: 0, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 2, Patch: 4},
				{Major: 1, Minor: 3, Patch: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(semver.SemVerByVersion(tt.versions))
			for i := range tt.versions {
				if tt.versions[i] != tt.want[i] {
					t.Errorf("Sort() failed at i=%d, got %+v, want %+v", i, tt.versions, tt.want)
				}
			}
		})
	}
}
