package cache

import (
	"testing"
)

func TestCacheType_Values(t *testing.T) {
	// The iota-based ordering is part of the package contract because
	// CacheName is indexed by CacheType. Guard against accidental reordering.
	tests := []struct {
		name string
		got  CacheType
		want CacheType
	}{
		{name: "CacheNone", got: CacheNone, want: 0},
		{name: "CacheMemory", got: CacheMemory, want: 1},
		{name: "CacheFile", got: CacheFile, want: 2},
		{name: "CacheDocker", got: CacheDocker, want: 3},
		{name: "CacheContainerd", got: CacheContainerd, want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestCacheName_Lookup(t *testing.T) {
	tests := []struct {
		name string
		key  CacheType
		want string
	}{
		{name: "none", key: CacheNone, want: "none"},
		{name: "memory", key: CacheMemory, want: "memory"},
		{name: "file", key: CacheFile, want: "file"},
		{name: "docker", key: CacheDocker, want: "docker"},
		{name: "containerd", key: CacheContainerd, want: "containerd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := CacheName[tt.key]
			if !ok {
				t.Fatalf("CacheName[%v] missing entry", tt.key)
			}
			if got != tt.want {
				t.Errorf("CacheName[%v] = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestCacheName_Completeness(t *testing.T) {
	// Every declared CacheType constant must have a name in the lookup table.
	// Update this list when new cache types are added.
	all := []CacheType{
		CacheNone,
		CacheMemory,
		CacheFile,
		CacheDocker,
		CacheContainerd,
	}

	if got, want := len(CacheName), len(all); got != want {
		t.Errorf("CacheName has %d entries, want %d", got, want)
	}

	seen := make(map[string]CacheType, len(all))
	for _, k := range all {
		name, ok := CacheName[k]
		if !ok {
			t.Errorf("CacheName is missing entry for %v", k)
			continue
		}
		if name == "" {
			t.Errorf("CacheName[%v] is empty", k)
		}
		if prev, dup := seen[name]; dup {
			t.Errorf("duplicate CacheName %q for %v and %v", name, prev, k)
		}
		seen[name] = k
	}
}

func TestCacheName_UnknownType(t *testing.T) {
	// Reading an unknown key must return the zero value ("") without panicking.
	if got := CacheName[CacheType(999)]; got != "" {
		t.Errorf("CacheName[999] = %q, want empty string", got)
	}
}

func TestCacheStruct_GetName(t *testing.T) {
	tests := []struct {
		name string
		typ  CacheType
		want string
	}{
		{name: "none", typ: CacheNone, want: "none"},
		{name: "memory", typ: CacheMemory, want: "memory"},
		{name: "file", typ: CacheFile, want: "file"},
		{name: "docker", typ: CacheDocker, want: "docker"},
		{name: "containerd", typ: CacheContainerd, want: "containerd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CacheStruct{Type: tt.typ}
			if got := c.GetName(); got != tt.want {
				t.Errorf("GetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCacheStruct_GetName_UnknownType(t *testing.T) {
	c := &CacheStruct{Type: CacheType(999)}
	if got := c.GetName(); got != "" {
		t.Errorf("GetName() for unknown type = %q, want empty string", got)
	}
}
