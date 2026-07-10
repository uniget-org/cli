package config

import (
	"strings"

	"gitlab.com/uniget-org/cli/pkg/tool"
)

func (c *Config) addDefaultPathRewriteRules() {
	rules := []tool.PathRewrite{
		{
			Source:    "usr/local/",
			Target:    "",
			Operation: "REPLACE",
		},
		{
			Source:    "var/lib/uniget/",
			Target:    c.GetLibDirectory() + "/",
			Operation: "REPLACE",
			Abort:     true,
		},
		{
			Source:    "var/cache/uniget/",
			Target:    c.GetCacheDirectory() + "/",
			Operation: "REPLACE",
			Abort:     true,
		},
	}

	if len(c.Target) > 0 {
		targetPath := c.Target
		if !strings.HasSuffix(targetPath, "/") {
			targetPath += "/"
		}
		rules = append(rules, tool.PathRewrite{
			Source:    "",
			Target:    targetPath,
			Operation: "PREPEND",
		})
	}

	c.PathRewriteRules = rules
}
