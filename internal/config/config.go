package config

import "time"

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
}

type AppConfig struct {
	Name            string        `yaml:"name"`
	Env             string        `yaml:"env"`
	Host            string        `yaml:"host"`
	Port            string        `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

func (c AppConfig) Address() string {
	return c.Host + ":" + c.Port
}

type DatabaseConfig struct {
	DSN                 string        `yaml:"dsn"`
	MaxOpenConns        int           `yaml:"max_open_conns"`
	MaxIdleConns        int           `yaml:"max_idle_conns"`
	ConnMaxLifetime     time.Duration `yaml:"conn_max_lifetime"`
	HealthCheckTimeout  time.Duration `yaml:"health_check_timeout"`
	StartupPingTimeout  time.Duration `yaml:"startup_ping_timeout"`
	StartupPingInterval time.Duration `yaml:"startup_ping_interval"`
}
