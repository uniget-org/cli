package config

import (
	"strings"

	"gitlab.com/uniget-org/cli/internal/constants"
)

func (c *Config) GetTarget() string {
	prefix, _ := strings.CutSuffix(c.Prefix, "/")
	return prefix + "/" + c.Target
}

func (c *Config) GetCacheRoot() string {
	prefix, _ := strings.CutSuffix(c.Prefix, "/")
	return prefix + "/" + c.CacheRoot
}

func (c *Config) GetLibRoot() string {
	prefix, _ := strings.CutSuffix(c.Prefix, "/")
	return prefix + "/" + c.LibRoot
}

func (c *Config) GetCacheDirectory() string {
	prefix, _ := strings.CutSuffix(c.Prefix, "/")
	return prefix + "/" + c.CacheRoot + "/" + constants.ProjectName
}

func (c *Config) GetLibDirectory() string {
	prefix, _ := strings.CutSuffix(c.Prefix, "/")
	return prefix + "/" + c.LibRoot + "/" + constants.ProjectName
}

func (c *Config) GetConfigDirectory() string {
	prefix, _ := strings.CutSuffix(c.Prefix, "/")
	return prefix + "/" + c.ConfigRoot + "/" + constants.ProjectName
}

func (c *Config) GetProfileDDirectory() string {
	prefix, _ := strings.CutSuffix(c.Prefix, "/")
	return prefix + "/" + c.ConfigRoot + "/profile.d"
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
