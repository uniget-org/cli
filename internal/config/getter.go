package config

import "gitlab.com/uniget-org/cli/internal/constants"

func (c *Config) GetCacheDirectory() string {
	return c.Prefix + c.CacheRoot + "/" + constants.ProjectName
}

func (c *Config) GetLibDirectory() string {
	return c.Prefix + c.LibRoot + "/" + constants.ProjectName
}

func (c *Config) GetConfigDirectory() string {
	return c.Prefix + c.ConfigRoot + "/" + constants.ProjectName
}

func (c *Config) GetProfileDDirectory() string {
	return c.Prefix + c.ConfigRoot + "/profile.d"
}

func (c *Config) GetMetadataFile() string {
	return c.GetCacheDirectory() + "/" + constants.MetadataFileName
}
