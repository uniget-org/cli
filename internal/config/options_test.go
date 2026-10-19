package config

import (
	"testing"
)

func TestConfigOptions(t *testing.T) {
	config := &Config{}
	for _, opt := range []ConfigOption{
		WithDebug(true),
		WithTrace(true),
		WithLogLevel("debug"),
		WithAutoUpdate(true),
		WithPrefix("/opt/uniget"),
		WithTarget("usr/bin"),
		WithIntegrateProfileD(true),
		WithIntegrateEtc(true),
	} {
		opt(config)
	}

	if config.Debug != true {
		t.Errorf("expected Debug to be true")
	}
	if config.Trace != true {
		t.Errorf("expected Trace to be true")
	}
	if config.LogLevel != "debug" {
		t.Errorf("expected LogLevel to be debug, got %s", config.LogLevel)
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
	if config.IntegrateProfileD != true {
		t.Errorf("expected IntegrateProfileD to be true")
	}
	if config.IntegrateEtc != true {
		t.Errorf("expected IntegrateEtc to be true")
	}
}

func TestWithIntegrateAll(t *testing.T) {
	config := &Config{}
	WithIntegrateAll(true)(config)
	if config.IntegrateProfileD != true {
		t.Errorf("expected IntegrateProfileD to be true")
	}
	if config.IntegrateEtc != true {
		t.Errorf("expected IntegrateEtc to be true")
	}

	WithIntegrateAll(false)(config)
	if config.IntegrateProfileD != false {
		t.Errorf("expected IntegrateProfileD to be false")
	}
	if config.IntegrateEtc != false {
		t.Errorf("expected IntegrateEtc to be false")
	}
}
