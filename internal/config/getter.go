package config

import (
	"strings"

	"gitlab.com/uniget-org/cli/internal/constants"
)

// joinRoot joins a root directory with the installation prefix. Absolute
// roots are used as-is which supports XDG directories pointing outside of
// $HOME in user mode.
func (c *Config) joinRoot(root string) string {
	if strings.HasPrefix(root, "/") {
		return strings.TrimSuffix(root, "/")
	}

	prefix, _ := strings.CutSuffix(c.Prefix, "/")
	return prefix + "/" + root
}

func (c *Config) GetTarget() string {
	return c.joinRoot(c.Target)
}

func (c *Config) GetCacheRoot() string {
	return c.joinRoot(c.CacheRoot)
}

func (c *Config) GetLibRoot() string {
	return c.joinRoot(c.LibRoot)
}

func (c *Config) GetCacheDirectory() string {
	return c.joinRoot(c.CacheRoot) + "/" + constants.ProjectName
}

func (c *Config) GetLibDirectory() string {
	return c.joinRoot(c.LibRoot) + "/" + constants.ProjectName
}

func (c *Config) GetConfigDirectory() string {
	return c.joinRoot(c.ConfigRoot) + "/" + constants.ProjectName
}

func (c *Config) GetProfileDDirectory() string {
	return c.joinRoot(c.ConfigRoot) + "/profile.d"
}

func (c *Config) GetMetadataFile() string {
	return c.GetCacheDirectory() + "/" + constants.MetadataFileName
}

func (c *Config) GetHooksPreInstallDirectory() string {
	return c.GetConfigDirectory() + "/" + constants.HooksPreInstallDirectory
}

func (c *Config) GetHooksPostInstallDirectory() string {
	return c.GetConfigDirectory() + "/" + constants.HooksPostInstallDirectory
}

func (c *Config) GetHooksPreUninstallDirectory() string {
	return c.GetConfigDirectory() + "/" + constants.HooksPreUninstallDirectory
}

func (c *Config) GetHooksPostUninstallDirectory() string {
	return c.GetConfigDirectory() + "/" + constants.HooksPostUninstallDirectory
}
