package config

import "gitlab.com/uniget-org/cli/internal/constants"

func (c *Config) GetCacheDirectory() string {
	return c.Prefix + "/" + c.CacheRoot + "/" + constants.ProjectName
}

func (c *Config) GetLibDirectory() string {
	return c.Prefix + "/" + c.LibRoot + "/" + constants.ProjectName
}

func (c *Config) GetConfigDirectory() string {
	return c.Prefix + "/" + c.ConfigRoot + "/" + constants.ProjectName
}

func (c *Config) GetProfileDDirectory() string {
	return c.Prefix + "/" + c.ConfigRoot + "/profile.d"
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
