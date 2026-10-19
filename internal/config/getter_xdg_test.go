package config

import (
	"testing"

	"gitlab.com/uniget-org/cli/internal/constants"
)

func TestPathGettersRootPrefix(t *testing.T) {
	// A prefix of "/" previously led to double-slash paths like
	// //var/cache or //usr/local.
	config := &Config{
		Prefix:     "/",
		Target:     "usr/local",
		CacheRoot:  "var/cache",
		LibRoot:    "var/lib",
		ConfigRoot: "etc",
	}
	if got := config.GetCacheDirectory(); got != "/var/cache/"+constants.ProjectName {
		t.Errorf("expected /var/cache/%s, got %s", constants.ProjectName, got)
	}
	if got := config.GetTarget(); got != "/usr/local" {
		t.Errorf("expected /usr/local, got %s", got)
	}
}

func TestPathGettersWithAbsoluteRoots(t *testing.T) {
	// XDG directories outside of $HOME produce absolute roots. They must
	// be used as-is instead of being joined under the prefix.
	config := &Config{
		Prefix:     "/home/nicholas",
		Target:     ".local",
		CacheRoot:  "/tmp/other-cache",
		LibRoot:    "/tmp/other-state",
		ConfigRoot: "/tmp/other-config",
	}
	if got := config.GetCacheDirectory(); got != "/tmp/other-cache/"+constants.ProjectName {
		t.Errorf("expected /tmp/other-cache/%s, got %s", constants.ProjectName, got)
	}
	if got := config.GetLibDirectory(); got != "/tmp/other-state/"+constants.ProjectName {
		t.Errorf("expected /tmp/other-state/%s, got %s", constants.ProjectName, got)
	}
	if got := config.GetCacheRoot(); got != "/tmp/other-cache" {
		t.Errorf("expected /tmp/other-cache, got %s", got)
	}
}
