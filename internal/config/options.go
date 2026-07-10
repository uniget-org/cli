package config

type ConfigOption func(*Config)

func WithDebug(debug bool) ConfigOption {
	return func(c *Config) {
		c.Debug = debug
	}
}

func WithTrace(trace bool) ConfigOption {
	return func(c *Config) {
		c.Trace = trace
	}
}

func WithLogLevel(logLevel string) ConfigOption {
	return func(c *Config) {
		c.LogLevel = logLevel
	}
}

func WithAutoUpdate(autoUpdate bool) ConfigOption {
	return func(c *Config) {
		c.AutoUpdate = autoUpdate
	}
}

func WithPrefix(prefix string) ConfigOption {
	return func(c *Config) {
		c.Prefix = prefix
	}
}

func WithTarget(target string) ConfigOption {
	return func(c *Config) {
		c.Target = target
	}
}

func WithIntegrateProfileD(integrate bool) ConfigOption {
	return func(c *Config) {
		c.IntegrateProfileD = integrate
	}
}

func WithIntegrateEtc(integrate bool) ConfigOption {
	return func(c *Config) {
		c.IntegrateEtc = integrate
	}
}

func WithIntegrateAll(integrate bool) ConfigOption {
	return func(c *Config) {
		c.IntegrateProfileD = integrate
		c.IntegrateEtc = integrate
	}
}
