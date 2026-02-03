package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/pterm/pterm"
	"gitlab.com/uniget-org/cli/pkg/cache"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

type Config struct {
	altArch string ``
	arch    string

	cacheRoot              string
	cacheDirectory         string
	libRoot                string
	libDirectory           string
	configRoot             string
	configDirectory        string
	hooksPreDirectory      string
	hooksPostDirectory     string
	profileDDirectory      string
	metadataImageTag       string
	metadataFileName       string
	metadataFile           string
	fileCacheDirectoryName string

	hooksPreInstallDirectory    string
	hooksPostInstallDirectory   string
	hooksPreUninstallDirectory  string
	hooksPostUninstallDirectory string

	registry            string `arg:"env=UNIGET_REGISTRY"`
	organization        string
	imageRepository     string `arg:"env=UNIGET_REPOSITORY"`
	toolSeparator       string `arg:"env=UNIGET_TOOLSEPARATOR"`
	registryImagePrefix string

	logLevel string `arg:"env=UNIGET_LOGLEVEL"`
	debug    bool   `arg:"env=UNIGET_DEBUG"`
	trace    bool   `arg:"env=UNIGET_TRACE"`

	prefix            string `arg:"env=UNIGET_PREFIX"`
	target            string `arg:"env=UNIGET_TARGET"`
	user              bool   `arg:"env=UNIGET_USER"`
	autoupdate        bool   `arg:"env=UNIGET_AUTOUPDATE"`
	integrateprofiled bool   `arg:"env=UNIGET_INTEGRATEPROFILED"`
	integrateetc      bool   `arg:"env=UNIGET_INTEGRATEETC"`
	integrateall      bool   `arg:"env=UNIGET_INTEGRATEALL"`
	cache             string `arg:"env=UNIGET_CACHE"`
	cachedirectory    string `arg:"env=UNIGET_CACHEDIRECTORY"`
	cacheretention    int    `arg:"env=UNIGET_CACHERETENTION"`

	pathRewriteRules []tool.PathRewrite
}

func DefaultConfig() *Config {
	config := Config{
		altArch: runtime.GOARCH,

		cacheRoot:              "var/cache",
		libRoot:                "var/lib",
		configRoot:             "etc",
		configDirectory:        "uniget",
		hooksPreDirectory:      "hooks/pre.d",
		hooksPostDirectory:     "hooks/post.d",
		metadataImageTag:       "main",
		metadataFileName:       "metadata.json",
		fileCacheDirectoryName: "downloads",

		hooksPreInstallDirectory:    "hooks/pre-install.d",
		hooksPostInstallDirectory:   "hooks/post-install.d",
		hooksPreUninstallDirectory:  "hooks/pre-install.d",
		hooksPostUninstallDirectory: "hooks/post-install.d",

		registry:      "ghcr.io",
		organization:  "uniget-org",
		toolSeparator: "/",

		logLevel: pterm.LogLevelInfo.String(),
		debug:    false,
		trace:    false,

		prefix:            "",
		target:            "usr/local",
		user:              false,
		autoupdate:        false,
		integrateprofiled: false,
		integrateetc:      false,
		integrateall:      false,
		cache:             "none",
		cacheretention:    26 * 60 * 60,
	}

	switch config.altArch {
	case "amd64":
		config.arch = "x86_64"
	case "arm64":
		config.arch = "aarch64"
	}

	return &config
}

func (config *Config) Update() error {
	arg.MustParse(&config)

	config.imageRepository = config.organization + "/tools"
	config.registryImagePrefix = config.registry + "/" + config.imageRepository + config.toolSeparator

	config.pathRewriteRules = []tool.PathRewrite{
		{
			Source:    "usr/local/",
			Target:    "",
			Operation: "REPLACE",
		},
		{
			Source:    "var/lib/uniget/",
			Target:    config.libDirectory + "/",
			Operation: "REPLACE",
			Abort:     true,
		},
		{
			Source:    "var/cache/uniget/",
			Target:    config.cacheDirectory + "/",
			Operation: "REPLACE",
			Abort:     true,
		},
	}

	if !config.user {
		config.cacheDirectory = config.cacheRoot + "/" + projectName
		config.libDirectory = config.libRoot + "/" + projectName
		config.profileDDirectory = config.configRoot + "/profile.d"
		config.configDirectory = config.configRoot + "/uniget"
		config.metadataFile = config.cacheDirectory + "/" + config.metadataFileName
		config.cachedirectory = config.cacheDirectory + "/" + config.fileCacheDirectoryName

	} else {
		config.prefix = os.Getenv("HOME")
		config.target = ".local"
		config.cacheRoot = ".cache"

		if os.Getenv("XDG_CACHE_HOME") != "" {
			if strings.HasPrefix(os.Getenv("XDG_CACHE_HOME"), os.Getenv("HOME")) {
				config.cacheRoot = strings.TrimPrefix(os.Getenv("XDG_CACHE_HOME"), os.Getenv("HOME")+"/")
			}
		}
		config.cacheDirectory = config.cacheRoot + "/" + projectName

		config.libRoot = ".local/state"
		if os.Getenv("XDG_STATE_HOME") != "" {
			if strings.HasPrefix(os.Getenv("XDG_STATE_HOME"), os.Getenv("HOME")) {
				config.libRoot = strings.TrimPrefix(os.Getenv("XDG_STATE_HOME"), os.Getenv("HOME")+"/")
			}
		}
		config.libDirectory = config.libRoot + "/" + projectName

		config.configRoot = ".config"
		if os.Getenv("XDG_CONFIG_HOME") != "" {
			if strings.HasPrefix(os.Getenv("XDG_CONFIG_HOME"), os.Getenv("HOME")) {
				config.configRoot = strings.TrimPrefix(os.Getenv("XDG_CONFIG_HOME"), os.Getenv("HOME")+"/")
			}
		}
		config.profileDDirectory = config.configRoot + "/profile.d"
		config.configDirectory = config.configRoot + "/uniget"

		config.metadataFile = config.cacheDirectory + "/" + config.metadataFileName
		config.cachedirectory = config.cacheDirectory + "/" + config.fileCacheDirectoryName
	}

	if len(config.prefix) > 0 {
		re, err := regexp.Compile(`^\/`)
		if err != nil {
			return fmt.Errorf("cannot compile regexp: %w", err)
		}
		if !re.MatchString(config.prefix) {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot determine working directory: %w", err)
			}
			config.prefix = wd + "/" + config.prefix
			logging.Debugf("Converted prefix to absolute path %s", config.prefix)
		}
	}

	if strings.HasPrefix(config.target, "/") {
		config.target = strings.TrimLeft(config.target, "/")
	}

	if !config.user {
		logging.Debugf("Adding path rewrite rules for system installation")

		config.pathRewriteRules = append(config.pathRewriteRules, tool.PathRewrite{
			Source:    "etc/systemd/",
			Target:    "/etc/systemd/",
			Operation: "REPLACE",
		})

		if config.integrateprofiled || config.integrateall {
			config.pathRewriteRules = append(config.pathRewriteRules, tool.PathRewrite{
				Source:    "etc/profile.d/",
				Target:    "/etc/profile.d/",
				Operation: "REPLACE",
			})
		}

	} else {
		logging.Debugf("Adding path rewrite rules for user installation")

		config.pathRewriteRules = append(config.pathRewriteRules, tool.PathRewrite{
			Source:    "libexec/docker/cli-plugins/",
			Target:    ".docker/cli-plugins/",
			Operation: "REPLACE",
			Abort:     true,
		})

		config.pathRewriteRules = append(config.pathRewriteRules, tool.PathRewrite{
			Source:    "etc/systemd/user/",
			Target:    ".config/systemd/user/",
			Operation: "REPLACE",
			Abort:     true,
		})

		if config.integrateprofiled || config.integrateall {
			config.pathRewriteRules = append(config.pathRewriteRules, tool.PathRewrite{
				Source:    "etc/profile.d/",
				Target:    ".config/profile.d/",
				Operation: "REPLACE",
				Abort:     true,
			})
		}

		if config.integrateetc || config.integrateall {
			config.pathRewriteRules = append(config.pathRewriteRules, tool.PathRewrite{
				Source:    "etc/",
				Target:    ".config/",
				Operation: "REPLACE",
				Abort:     true,
			})
		}
	}

	if len(config.target) > 0 {
		targetPath := config.target
		if !strings.HasSuffix(targetPath, "/") {
			targetPath += "/"
		}
		config.pathRewriteRules = append(config.pathRewriteRules, tool.PathRewrite{
			Source:    "",
			Target:    targetPath,
			Operation: "PREPEND",
		})
	}
	if config.debug {
		logging.Debug("Path rewrite rules:")
		for _, rule := range config.pathRewriteRules {
			logging.Debugf("  %s -> %s (%s)", rule.Source, rule.Target, rule.Operation)
		}
	}

	var err error
	switch config.cache {
	case "none":
		logging.Debug("Using no cache")
		toolCache = cache.NewNoneCache()

	case "file":
		logging.Debug("Using file cache")
		fileCacheDir := config.prefix + "/" + config.cachedirectory
		assertDirectory(fileCacheDir)
		toolCache = cache.NewFileCache(fileCacheDir, config.cacheretention)

	case "docker":
		if containers.DockerIsAvailable() {
			logging.Debug("Using docker cache")
			toolCache, err = cache.NewDockerCache()
			if err != nil {
				return fmt.Errorf("error creating Docker cache: %s", err)
			}
		} else {
			logging.Warning.Println("Docker is not available. Falling back to no cache")
			toolCache = cache.NewNoneCache()
		}

	case "containerd":
		if containers.ContainerdIsAvailable() {
			logging.Debug("Using containerd cache")
			toolCache, err = cache.NewContainerdCache(projectName)
			if err != nil {
				return fmt.Errorf("error creating Containerd cache: %s", err)
			}
		} else {
			logging.Warning.Println("Containerd is not available. Falling back to no cache")
			toolCache = cache.NewNoneCache()
		}

	default:
		return fmt.Errorf("unsupported cache backend: %s", config.cache)
	}

	return nil
}

func (config *Config) Debug() {
	logging.Debugf("user: %t", config.user)
	logging.Debugf("prefix: %s", config.prefix)
	logging.Debugf("target: %s", config.target)
	logging.Debugf("configRoot: %s", config.configRoot)
	logging.Debugf("configDirectory: %s", config.configDirectory)
	logging.Debugf("cacheRoot: %s", config.cacheRoot)
	logging.Debugf("cacheDirectory: %s", config.cacheDirectory)
	logging.Debugf("libRoot: %s", config.libRoot)
	logging.Debugf("libDirectory: %s", config.libDirectory)
	logging.Debugf("metadataFile: %s", config.metadataFile)
	logging.Debugf("registry: %s", config.registry)
	logging.Debugf("repository: %s", config.imageRepository)
	logging.Debugf("tool-separator: %s", config.toolSeparator)
	logging.Debugf("cache: %s", config.cache)
	logging.Debugf("cache-directory: %s", config.cachedirectory)
	logging.Debugf("cache-retention: %d", config.cacheretention)
}
