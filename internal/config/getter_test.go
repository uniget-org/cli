package config

import (
	"testing"

	"gitlab.com/uniget-org/cli/internal/constants"
)

const testPrefix = "/opt/uniget"

func newPathGetterTestConfig(prefix string) *Config {
	return &Config{
		Prefix:     prefix,
		Target:     "usr/local",
		CacheRoot:  "var/cache",
		LibRoot:    "var/lib",
		ConfigRoot: "etc",
	}
}

func TestPathGetters(t *testing.T) {
	tests := []struct {
		name string
		get  func(*Config) string
		want string
	}{
		{"GetTarget", (*Config).GetTarget, testPrefix + "/usr/local"},
		{"GetCacheRoot", (*Config).GetCacheRoot, testPrefix + "/var/cache"},
		{"GetLibRoot", (*Config).GetLibRoot, testPrefix + "/var/lib"},
		{"GetCacheDirectory", (*Config).GetCacheDirectory, testPrefix + "/var/cache/" + constants.ProjectName},
		{"GetLibDirectory", (*Config).GetLibDirectory, testPrefix + "/var/lib/" + constants.ProjectName},
		{"GetConfigDirectory", (*Config).GetConfigDirectory, testPrefix + "/etc/" + constants.ProjectName},
		{"GetProfileDDirectory", (*Config).GetProfileDDirectory, testPrefix + "/etc/profile.d"},
		{"GetMetadataFile", (*Config).GetMetadataFile, testPrefix + "/var/cache/" + constants.ProjectName + "/" + constants.MetadataFileName},
		{"GetHooksPreInstallDirectory", (*Config).GetHooksPreInstallDirectory, testPrefix + "/etc/" + constants.ProjectName + "/" + constants.HooksPreInstallDirectory},
		{"GetHooksPostInstallDirectory", (*Config).GetHooksPostInstallDirectory, testPrefix + "/etc/" + constants.ProjectName + "/" + constants.HooksPostInstallDirectory},
		{"GetHooksPreUninstallDirectory", (*Config).GetHooksPreUninstallDirectory, testPrefix + "/etc/" + constants.ProjectName + "/" + constants.HooksPreUninstallDirectory},
		{"GetHooksPostUninstallDirectory", (*Config).GetHooksPostUninstallDirectory, testPrefix + "/etc/" + constants.ProjectName + "/" + constants.HooksPostUninstallDirectory},
	}
	for _, prefix := range []string{testPrefix + "/", testPrefix} {
		config := newPathGetterTestConfig(prefix)
		for _, tt := range tests {
			t.Run(prefix+"|"+tt.name, func(t *testing.T) {
				if got := tt.get(config); got != tt.want {
					t.Errorf("expected %s, got %s", tt.want, got)
				}
			})
		}
	}
}
