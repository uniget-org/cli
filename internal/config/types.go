package config

import (
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

type Config struct {
	Arch                        string
	AltArch                     string
	LogLevel                    string `env:"UNIGET_LOGLEVEL"`
	Debug                       bool   `env:"UNIGET_DEBUG"`
	Trace                       bool   `env:"UNIGET_TRACE"`
	User                        bool   `env:"UNIGET_USER"`
	AutoUpdate                  bool   `env:"UNIGET_AUTOUPDATE"`
	Prefix                      string `env:"UNIGET_PREFIX"`
	Target                      string `env:"UNIGET_TARGET"`
	CacheRoot                   string
	LibRoot                     string
	ConfigRoot                  string
	IntegrateProfileD           bool `env:"UNIGET_INTEGRATEPROFILED"`
	IntegrateEtc                bool `env:"UNIGET_INTEGRATEETC"`
	IntegrateAll                bool `env:"UNIGET_INTEGRATEALL"`
	PathRewriteRules            []tool.PathRewrite
	HooksPreInstallDirectory    string
	HooksPostInstallDirectory   string
	HooksPreUninstallDirectory  string
	HooksPostUninstallDirectory string
	Cache                       string `env:"UNIGET_CACHE"`
	FileCacheRetention          int    `env:"UNIGET_CACHERETENTION"`
	FileCacheDirectoryName      string `env:"UNIGET_CACHEDIRECTORY"`
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

	t := reflect.TypeOf(*config)
	v := reflect.ValueOf(config).Elem()
	for i := 0; i < t.NumField(); i++ {
		fieldDef := t.Field(i)

		envVarName := fieldDef.Tag.Get("env")
		if len(envVarName) > 0 && len(os.Getenv(envVarName)) > 0 {
			field := v.FieldByName(fieldDef.Name)
			switch field.Kind() {
			case reflect.String:
				field.SetString(os.Getenv(envVarName))
			case reflect.Bool:
				val, err := strconv.ParseBool(os.Getenv(envVarName))
				if err == nil {
					field.SetBool(val)
				}
			case reflect.Int:
				val, err := strconv.Atoi(os.Getenv(envVarName))
				if err == nil {
					field.SetInt(int64(val))
				}
			default:
				panic("Unsupported type " + field.Kind().String() + " for field: " + fieldDef.Name)
			}
		}
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
	c.CacheRoot = strings.TrimPrefix(os.Getenv("XDG_CACHE_HOME"), os.Getenv("HOME")+"/")
	c.LibRoot = strings.TrimPrefix(os.Getenv("XDG_STATE_HOME"), os.Getenv("HOME")+"/")
	c.ConfigRoot = strings.TrimPrefix(os.Getenv("XDG_CONFIG_HOME"), os.Getenv("HOME")+"/")

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

func (c *Config) String() string {
	return "Config{" + "\n" +
		"  Arch: " + c.Arch + ", " + "\n" +
		"  AltArch: " + c.AltArch + ", " + "\n" +
		"  LogLevel: " + c.LogLevel + ", " + "\n" +
		"  Debug: " + strconv.FormatBool(c.Debug) + ", " + "\n" +
		"  Trace: " + strconv.FormatBool(c.Trace) + ", " + "\n" +
		"  User: " + strconv.FormatBool(c.User) + ", " + "\n" +
		"  AutoUpdate: " + strconv.FormatBool(c.AutoUpdate) + ", " + "\n" +
		"  Prefix: " + c.Prefix + ", " + "\n" +
		"  Target: " + c.Target + ", " + "\n" +
		"  CacheRoot: " + c.CacheRoot + ", " + "\n" +
		"  LibRoot: " + c.LibRoot + ", " + "\n" +
		"  ConfigRoot: " + c.ConfigRoot + ", " + "\n" +
		"  IntegrateProfileD: " + strconv.FormatBool(c.IntegrateProfileD) + ", " + "\n" +
		"  IntegrateEtc: " + strconv.FormatBool(c.IntegrateEtc) + ", " + "\n" +
		"  IntegrateAll: " + strconv.FormatBool(c.IntegrateAll) + ", " + "\n" +
		"  Cache: " + c.Cache + ", " + "\n" +
		"  FileCacheRetention: " + strconv.Itoa(c.FileCacheRetention) + ", " + "\n" +
		"  FileCacheDirectoryName: " + c.FileCacheDirectoryName + "\n" +
		"  CacheDirectory: " + c.GetCacheDirectory() + ", " + "\n" +
		"  LibDirectory: " + c.GetLibDirectory() + ", " + "\n" +
		"  ConfigDirectory: " + c.GetConfigDirectory() + ", " + "\n" +
		"  ProfileDDirectory: " + c.GetProfileDDirectory() + ", " + "\n" +
		"  MetadataFile: " + c.GetMetadataFile() + "\n" +
		"}"
}
