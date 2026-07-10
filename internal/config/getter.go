package config

import "gitlab.com/uniget-org/cli/internal/constants"

func (c *Config) GetCacheDirectory() string {
	return c.CacheRoot + "/" + constants.ProjectName
}

func (c *Config) GetLibDirectory() string {
	return c.LibRoot + "/" + constants.ProjectName
}

func (c *Config) GetConfigDirectory() string {
	return c.ConfigRoot + "/" + constants.ProjectName
}

func (c *Config) GetProfileDDirectory() string {
	return c.ConfigRoot + "/profile.d"
}

func (c *Config) GetMetadataFile() string {
	return c.GetCacheDirectory() + "/" + constants.MetadataFileName
}
