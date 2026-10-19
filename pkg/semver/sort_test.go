package semver

import (
	"reflect"
	"sort"
	"testing"
)

func TestByVersionMethods(t *testing.T) {
	versions := ByVersion{"2.0.0", "1.0.0"}

	if versions.Len() != 2 {
		t.Errorf("expected len 2, got %d", versions.Len())
	}
	if !versions.Less(1, 0) {
		t.Errorf("expected 1.0.0 to be less than 2.0.0")
	}
	if versions.Less(0, 1) {
		t.Errorf("expected 2.0.0 not to be less than 1.0.0")
	}
	if versions.Less(0, 0) {
		t.Errorf("expected version not to be less than itself")
	}

	versions.Swap(0, 1)
	if versions[0] != "1.0.0" || versions[1] != "2.0.0" {
		t.Errorf("expected swap to exchange elements, got %v", versions)
	}
}

func TestSortByVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "numeric ordering ignores string sort",
			input:    []string{"1.10.0", "1.2.3", "1.0.0"},
			expected: []string{"1.0.0", "1.2.3", "1.10.0"},
		},
		{
			name:     "prerelease sorts before release",
			input:    []string{"1.0.0", "1.0.0-alpha"},
			expected: []string{"1.0.0-alpha", "1.0.0"},
		},
		{
			name:     "v prefix",
			input:    []string{"v1.0.0", "v0.1.0"},
			expected: []string{"v0.1.0", "v1.0.0"},
		},
		{
			name:     "major ordering",
			input:    []string{"3.0.0", "1.0.0", "2.0.0"},
			expected: []string{"1.0.0", "2.0.0", "3.0.0"},
		},
		{
			name:     "single element",
			input:    []string{"1.0.0"},
			expected: []string{"1.0.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := append([]string(nil), tt.input...)
			sort.Sort(ByVersion(got))
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
