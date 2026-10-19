package semver

import (
	"reflect"
	"sort"
	"testing"
)

func TestByVersionLessWithUnparseable(t *testing.T) {
	versions := ByVersion{"1.0.0", "not-a-version", "garbage"}

	// Parseable versions sort before unparseable versions
	if !versions.Less(0, 1) {
		t.Errorf("expected parseable version to be less than unparseable version")
	}
	if versions.Less(1, 0) {
		t.Errorf("expected unparseable version not to be less than parseable version")
	}

	// Two unparseable versions: fall back to string comparison where
	// "garbage" sorts before "not-a-version"
	if !versions.Less(2, 1) {
		t.Errorf("expected garbage to be less than not-a-version by string order")
	}
	if versions.Less(1, 2) {
		t.Errorf("expected not-a-version not to be less than garbage by string order")
	}
}

func TestSortWithUnparseableVersions(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "unparseable versions don't panic and sort last",
			input:    []string{"1.0.0", "not-a-version", "2.0.0"},
			expected: []string{"1.0.0", "2.0.0", "not-a-version"},
		},
		{
			name:     "only unparseable versions fall back to string order",
			input:    []string{"banana", "apple"},
			expected: []string{"apple", "banana"},
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
