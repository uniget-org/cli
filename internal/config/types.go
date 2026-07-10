package config

import (
	"os"
	"runtime"
	"strconv"

	"github.com/pterm/pterm"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

type Config struct {
	Arch                        string
	AltArch                     string
	LogLevel                    string
	Debug                       bool
	Trace                       bool
	User                        bool
	AutoUpdate                  bool
	Prefix                      string
	Target                      string
	CacheRoot                   string
	LibRoot                     string
	ConfigRoot                  string
	IntegrateProfileD           bool
	IntegrateEtc                bool
	IntegrateAll                bool
	PathRewriteRules            []tool.PathRewrite
	HooksPreInstallDirectory    string
	HooksPostInstallDirectory   string
	HooksPreUninstallDirectory  string
	HooksPostUninstallDirectory string
	Cache                       string
	FileCacheRetention          int
	FileCacheDirectoryName      string
}

func NewDefaultConfig(opts ...ConfigOption) *Config {
	config := &Config{
		AltArch:                runtime.GOARCH,
		LogLevel:               pterm.LogLevelInfo.String(),
		Debug:                  false,
		Trace:                  false,
		User:                   false,
		AutoUpdate:             false,
		Prefix:                 "",
		Target:                 "usr/local",
		IntegrateProfileD:      false,
		IntegrateEtc:           false,
		IntegrateAll:           false,
		Cache:                  "none",
		FileCacheRetention:     24 * 60 * 60,
		FileCacheDirectoryName: "downloads",
	}
	for _, opt := range opts {
		opt(config)
	}

	if len(os.Getenv("UNIGET_LOGLEVEL")) > 0 {
		config.LogLevel = os.Getenv("UNIGET_LOGLEVEL")
	}
	if len(os.Getenv("UNIGET_DEBUG")) > 0 {
		config.Debug, _ = strconv.ParseBool(os.Getenv("UNIGET_DEBUG"))
	}
	if len(os.Getenv("UNIGET_TRACE")) > 0 {
		config.Trace, _ = strconv.ParseBool(os.Getenv("UNIGET_TRACE"))
	}
	if len(os.Getenv("UNIGET_USER")) > 0 {
		config.User, _ = strconv.ParseBool(os.Getenv("UNIGET_USER"))
	}
	if len(os.Getenv("UNIGET_AUTOUPDATE")) > 0 {
		config.AutoUpdate, _ = strconv.ParseBool(os.Getenv("UNIGET_AUTOUPDATE"))
	}
	if len(os.Getenv("UNIGET_PREFIX")) > 0 {
		config.Prefix = os.Getenv("UNIGET_PREFIX")
	}
	if len(os.Getenv("UNIGET_TARGET")) > 0 {
		config.Target = os.Getenv("UNIGET_TARGET")
	}
	if len(os.Getenv("UNIGET_INTEGRATEPROFILED")) > 0 {
		config.IntegrateProfileD, _ = strconv.ParseBool(os.Getenv("UNIGET_INTEGRATEPROFILED"))
	}
	if len(os.Getenv("UNIGET_INTEGRATEETC")) > 0 {
		config.IntegrateEtc, _ = strconv.ParseBool(os.Getenv("UNIGET_INTEGRATEETC"))
	}
	if len(os.Getenv("UNIGET_INTEGRATEALL")) > 0 {
		config.IntegrateAll, _ = strconv.ParseBool(os.Getenv("UNIGET_INTEGRATEALL"))
	}
	if len(os.Getenv("UNIGET_CACHE")) > 0 {
		config.Cache = os.Getenv("UNIGET_CACHE")
	}
	if len(os.Getenv("UNIGET_CACHERETENTION")) > 0 {
		retention, err := strconv.Atoi(os.Getenv("UNIGET_CACHERETENTION"))
		if err == nil {
			config.FileCacheRetention = retention
		}
	}
	if len(os.Getenv("UNIGET_CACHEDIRECTORY")) > 0 {
		config.FileCacheDirectoryName = os.Getenv("UNIGET_CACHEDIRECTORY")
	}

	if config.IntegrateAll {
		config.IntegrateProfileD = true
		config.IntegrateEtc = true
	}

	switch config.AltArch {
	case "amd64":
		config.Arch = "x86_64"
	case "arm64":
		config.Arch = "aarch64"
	default:
		logging.Error.Printfln("Unsupported architecture: %s", config.AltArch)
		os.Exit(1)
	}

	return config
}

func (c *Config) SetGlobalConfig(opts ...ConfigOption) {
	c.CacheRoot = "/var/cache"
	c.LibRoot = "/var/lib"
	c.ConfigRoot = "/etc"

	for _, opt := range opts {
		opt(c)
	}

	c.addDefaultPathRewriteRules()
	c.PathRewriteRules = append(c.PathRewriteRules, tool.PathRewrite{
		Source:    "etc/systemd/",
		Target:    "/etc/systemd/",
		Operation: "REPLACE",
	})
	if c.IntegrateProfileD {
		c.PathRewriteRules = append(c.PathRewriteRules, tool.PathRewrite{
			Source:    "etc/profile.d/",
			Target:    "/etc/profile.d/",
			Operation: "REPLACE",
		})
	}
}

func (c *Config) SetUserConfig(opts ...ConfigOption) {
	c.Prefix = os.Getenv("HOME")
	c.Target = ".local"
	c.CacheRoot = os.Getenv("XDG_CACHE_HOME")
	c.LibRoot = os.Getenv("XDG_STATE_HOME")
	c.ConfigRoot = os.Getenv("XDG_CONFIG_HOME")

	for _, opt := range opts {
		opt(c)
	}

	if c.CacheRoot == "" {
		c.CacheRoot = ".cache"
	}
	if c.LibRoot == "" {
		c.LibRoot = ".local/state"
	}
	if c.ConfigRoot == "" {
		c.ConfigRoot = ".config"
	}

	c.addDefaultPathRewriteRules()
	c.PathRewriteRules = append(c.PathRewriteRules, []tool.PathRewrite{
		{
			Source:    "libexec/docker/cli-plugins/",
			Target:    ".docker/cli-plugins/",
			Operation: "REPLACE",
			Abort:     true,
		},
		{
			Source:    "etc/systemd/user/",
			Target:    ".config/systemd/user/",
			Operation: "REPLACE",
			Abort:     true,
		},
	}...)
	if c.IntegrateProfileD {
		c.PathRewriteRules = append(c.PathRewriteRules, tool.PathRewrite{
			Source:    "etc/profile.d/",
			Target:    ".config/profile.d/",
			Operation: "REPLACE",
			Abort:     true,
		})
	}
	if c.IntegrateEtc {
		c.PathRewriteRules = append(c.PathRewriteRules, tool.PathRewrite{
			Source:    "etc/",
			Target:    ".config/",
			Operation: "REPLACE",
			Abort:     true,
		})
	}
}
