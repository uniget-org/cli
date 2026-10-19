package config

import (
	"runtime"
	"strings"
	"testing"

	"github.com/pterm/pterm"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"UNIGET_LOGLEVEL",
		"UNIGET_DEBUG",
		"UNIGET_TRACE",
		"UNIGET_USER",
		"UNIGET_AUTOUPDATE",
		"UNIGET_PREFIX",
		"UNIGET_TARGET",
		"UNIGET_INTEGRATEPROFILED",
		"UNIGET_INTEGRATEETC",
		"UNIGET_INTEGRATEALL",
		"UNIGET_CACHE",
		"UNIGET_CACHERETENTION",
		"UNIGET_CACHEDIRECTORY",
		"UNIGET_VERIFY_IMAGE_SIGNATURE",
	} {
		t.Setenv(name, "")
	}
}

func expectedArch(altArch string) string {
	switch altArch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return ""
	}
}

func TestNewDefaultConfig(t *testing.T) {
	clearConfigEnv(t)

	config := NewDefaultConfig()

	if config.AltArch != runtime.GOARCH {
		t.Errorf("expected AltArch to be %s, got %s", runtime.GOARCH, config.AltArch)
	}
	if wantArch := expectedArch(runtime.GOARCH); wantArch != "" && config.Arch != wantArch {
		t.Errorf("expected Arch to be %s, got %s", wantArch, config.Arch)
	}
	if config.LogLevel != pterm.LogLevelInfo.String() {
		t.Errorf("expected LogLevel to be %s, got %s", pterm.LogLevelInfo.String(), config.LogLevel)
	}
	if config.Debug != false {
		t.Errorf("expected Debug to be false")
	}
	if config.Trace != false {
		t.Errorf("expected Trace to be false")
	}
	if config.User != false {
		t.Errorf("expected User to be false")
	}
	if config.AutoUpdate != false {
		t.Errorf("expected AutoUpdate to be false")
	}
	if config.Prefix != "/" {
		t.Errorf("expected Prefix to be /, got %s", config.Prefix)
	}
	if config.Target != "usr/local" {
		t.Errorf("expected Target to be usr/local, got %s", config.Target)
	}
	if config.IntegrateProfileD != false {
		t.Errorf("expected IntegrateProfileD to be false")
	}
	if config.IntegrateEtc != false {
		t.Errorf("expected IntegrateEtc to be false")
	}
	if config.IntegrateAll != false {
		t.Errorf("expected IntegrateAll to be false")
	}
	if config.Cache != "none" {
		t.Errorf("expected Cache to be none, got %s", config.Cache)
	}
	if config.FileCacheRetention != 24*60*60 {
		t.Errorf("expected FileCacheRetention to be %d, got %d", 24*60*60, config.FileCacheRetention)
	}
	if config.FileCacheDirectoryName != "downloads" {
		t.Errorf("expected FileCacheDirectoryName to be downloads, got %s", config.FileCacheDirectoryName)
	}
	if config.VerifyImageSignature != true {
		t.Errorf("expected VerifyImageSignature to be true")
	}
}

func TestNewDefaultConfigWithEnv(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("UNIGET_LOGLEVEL", "debug")
	t.Setenv("UNIGET_DEBUG", "true")
	t.Setenv("UNIGET_TRACE", "1")
	t.Setenv("UNIGET_USER", "t")
	t.Setenv("UNIGET_AUTOUPDATE", "TRUE")
	t.Setenv("UNIGET_PREFIX", "/opt/uniget")
	t.Setenv("UNIGET_TARGET", "usr/bin")
	t.Setenv("UNIGET_CACHE", "file")
	t.Setenv("UNIGET_CACHERETENTION", "3600")
	t.Setenv("UNIGET_CACHEDIRECTORY", "cache")
	t.Setenv("UNIGET_VERIFY_IMAGE_SIGNATURE", "false")

	config := NewDefaultConfig()

	if config.LogLevel != "debug" {
		t.Errorf("expected LogLevel to be debug, got %s", config.LogLevel)
	}
	if config.Debug != true {
		t.Errorf("expected Debug to be true")
	}
	if config.Trace != true {
		t.Errorf("expected Trace to be true")
	}
	if config.User != true {
		t.Errorf("expected User to be true")
	}
	if config.AutoUpdate != true {
		t.Errorf("expected AutoUpdate to be true")
	}
	if config.Prefix != "/opt/uniget" {
		t.Errorf("expected Prefix to be /opt/uniget, got %s", config.Prefix)
	}
	if config.Target != "usr/bin" {
		t.Errorf("expected Target to be usr/bin, got %s", config.Target)
	}
	if config.Cache != "file" {
		t.Errorf("expected Cache to be file, got %s", config.Cache)
	}
	if config.FileCacheRetention != 3600 {
		t.Errorf("expected FileCacheRetention to be 3600, got %d", config.FileCacheRetention)
	}
	if config.FileCacheDirectoryName != "cache" {
		t.Errorf("expected FileCacheDirectoryName to be cache, got %s", config.FileCacheDirectoryName)
	}
	if config.VerifyImageSignature != false {
		t.Errorf("expected VerifyImageSignature to be false")
	}
}

func TestNewDefaultConfigIgnoresInvalidEnv(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("UNIGET_DEBUG", "notabool")
	t.Setenv("UNIGET_CACHERETENTION", "notanumber")
	t.Setenv("UNIGET_PREFIX", "")

	config := NewDefaultConfig()

	if config.Debug != false {
		t.Errorf("expected Debug to remain false")
	}
	if config.FileCacheRetention != 24*60*60 {
		t.Errorf("expected FileCacheRetention to remain %d, got %d", 24*60*60, config.FileCacheRetention)
	}
	if config.Prefix != "/" {
		t.Errorf("expected Prefix to remain /, got %s", config.Prefix)
	}
}

func TestNewDefaultConfigWithIntegrateEnv(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("UNIGET_INTEGRATEALL", "true")

	config := NewDefaultConfig()

	if config.IntegrateAll != true {
		t.Errorf("expected IntegrateAll to be true")
	}
	if config.IntegrateProfileD != true {
		t.Errorf("expected IntegrateProfileD to be true")
	}
	if config.IntegrateEtc != true {
		t.Errorf("expected IntegrateEtc to be true")
	}
}

func TestSetGlobalConfig(t *testing.T) {
	clearConfigEnv(t)

	config := NewDefaultConfig()
	config.SetGlobalConfig()

	if config.CacheRoot != "var/cache" {
		t.Errorf("expected CacheRoot to be var/cache, got %s", config.CacheRoot)
	}
	if config.LibRoot != "var/lib" {
		t.Errorf("expected LibRoot to be var/lib, got %s", config.LibRoot)
	}
	if config.ConfigRoot != "etc" {
		t.Errorf("expected ConfigRoot to be etc, got %s", config.ConfigRoot)
	}
	if len(config.PathRewriteRules) != 5 {
		t.Fatalf("expected 5 path rewrite rules, got %d", len(config.PathRewriteRules))
	}
	// The etc/systemd/ rule is appended last, after the default PREPEND rule.
	last := config.PathRewriteRules[len(config.PathRewriteRules)-1]
	if last.Source != "etc/systemd/" || last.Target != "/etc/systemd/" || last.Operation != "REPLACE" {
		t.Errorf("expected final rule to rewrite etc/systemd/, got %+v", last)
	}
	preceding := config.PathRewriteRules[len(config.PathRewriteRules)-2]
	if preceding.Operation != "PREPEND" || preceding.Target != "usr/local/" {
		t.Errorf("expected preceding rule to prepend usr/local/, got %+v", preceding)
	}
}

func TestSetGlobalConfigWithIntegrateProfileD(t *testing.T) {
	clearConfigEnv(t)

	config := NewDefaultConfig()
	config.SetGlobalConfig(WithIntegrateProfileD(true))

	if len(config.PathRewriteRules) != 6 {
		t.Fatalf("expected 6 path rewrite rules, got %d", len(config.PathRewriteRules))
	}
	last := config.PathRewriteRules[len(config.PathRewriteRules)-1]
	if last.Source != "etc/profile.d/" || last.Target != "/etc/profile.d/" || last.Operation != "REPLACE" {
		t.Errorf("expected final rule to rewrite etc/profile.d/, got %+v", last)
	}
}

func TestSetUserConfig(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("HOME", "/home/tester")

	config := NewDefaultConfig()
	config.SetUserConfig()

	if config.Prefix != "/home/tester" {
		t.Errorf("expected Prefix to be /home/tester, got %s", config.Prefix)
	}
	if config.Target != ".local" {
		t.Errorf("expected Target to be .local, got %s", config.Target)
	}
	if config.CacheRoot != ".cache" {
		t.Errorf("expected CacheRoot to be .cache, got %s", config.CacheRoot)
	}
	if config.LibRoot != ".local/state" {
		t.Errorf("expected LibRoot to be .local/state, got %s", config.LibRoot)
	}
	if config.ConfigRoot != ".config" {
		t.Errorf("expected ConfigRoot to be .config, got %s", config.ConfigRoot)
	}
	if len(config.PathRewriteRules) != 7 {
		t.Fatalf("expected 7 path rewrite rules, got %d", len(config.PathRewriteRules))
	}
	if config.PathRewriteRules[3].Source != "" ||
		config.PathRewriteRules[3].Target != ".local/" ||
		config.PathRewriteRules[3].Operation != "PREPEND" {
		t.Errorf("rule 3 is invalid: %+v", config.PathRewriteRules[3])
	}
	if config.PathRewriteRules[4].Source != "libexec/docker/cli-plugins/" ||
		config.PathRewriteRules[4].Target != ".docker/cli-plugins/" ||
		config.PathRewriteRules[4].Operation != "REPLACE" ||
		config.PathRewriteRules[4].Abort != true {
		t.Errorf("rule 4 is invalid: %+v", config.PathRewriteRules[4])
	}
	if config.PathRewriteRules[5].Source != "etc/systemd/user/" ||
		config.PathRewriteRules[5].Target != ".config/systemd/user/" ||
		config.PathRewriteRules[5].Abort != true {
		t.Errorf("rule 5 is invalid: %+v", config.PathRewriteRules[5])
	}
	if config.PathRewriteRules[6].Source != ".local/share/terminfo/" ||
		config.PathRewriteRules[6].Target != ".terminfo/" ||
		config.PathRewriteRules[6].Abort != true {
		t.Errorf("rule 6 is invalid: %+v", config.PathRewriteRules[6])
	}
}

func TestSetUserConfigWithXDGDirectories(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("HOME", "/home/tester")
	t.Setenv("XDG_CACHE_HOME", "/home/tester/cache")
	t.Setenv("XDG_STATE_HOME", "/home/tester/state")
	t.Setenv("XDG_CONFIG_HOME", "/home/tester/config")

	config := NewDefaultConfig()
	config.SetUserConfig()

	if config.CacheRoot != "cache" {
		t.Errorf("expected CacheRoot to be cache, got %s", config.CacheRoot)
	}
	if config.LibRoot != "state" {
		t.Errorf("expected LibRoot to be state, got %s", config.LibRoot)
	}
	if config.ConfigRoot != "config" {
		t.Errorf("expected ConfigRoot to be config, got %s", config.ConfigRoot)
	}
}

func TestSetUserConfigWithXDGOutsideHome(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("HOME", "/home/tester")
	t.Setenv("XDG_CACHE_HOME", "/var/cache/tester")

	config := NewDefaultConfig()
	config.SetUserConfig()

	if config.CacheRoot != "/var/cache/tester" {
		t.Errorf("expected CacheRoot to be /var/cache/tester, got %s", config.CacheRoot)
	}
}

func TestSetUserConfigWithIntegrations(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("HOME", "/home/tester")

	config := NewDefaultConfig()
	config.SetUserConfig(WithIntegrateProfileD(true), WithIntegrateEtc(true))

	if len(config.PathRewriteRules) != 9 {
		t.Fatalf("expected 9 path rewrite rules, got %d", len(config.PathRewriteRules))
	}
	profileD := config.PathRewriteRules[7]
	if profileD.Source != "etc/profile.d/" || profileD.Target != ".config/profile.d/" || profileD.Operation != "REPLACE" || profileD.Abort != true {
		t.Errorf("expected profile.d rule, got %+v", profileD)
	}
	etc := config.PathRewriteRules[8]
	if etc.Source != "etc/" || etc.Target != ".config/" || etc.Operation != "REPLACE" || etc.Abort != true {
		t.Errorf("expected etc rule, got %+v", etc)
	}
}

func TestString(t *testing.T) {
	clearConfigEnv(t)

	config := NewDefaultConfig(WithPrefix("/opt/uniget"))
	config.SetGlobalConfig()
	output := config.String()

	for _, want := range []string{
		"Prefix: /opt/uniget,",
		"Target: usr/local",
		"FileCacheRetention: " + "86400",
		"CacheDirectory: /opt/uniget/var/cache/uniget",
		"LibDirectory: /opt/uniget/var/lib/uniget",
		"ConfigDirectory: /opt/uniget/etc/uniget",
		"ProfileDDirectory: /opt/uniget/etc/profile.d",
		"MetadataFile: /opt/uniget/var/cache/uniget/metadata.json",
		"VerifyImageSignature: true",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected string output to contain %q", want)
		}
	}
}
