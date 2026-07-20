package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
	RDAP      RDAPConfig      `yaml:"rdap"`
	TLS       TLSConfig       `yaml:"tls"`
	Domains   []DomainConfig  `yaml:"domains"`
}

type ServerConfig struct {
	ListenAddress string `yaml:"listen_address"`
}

type SchedulerConfig struct {
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	WorkerCount     int           `yaml:"worker_count"`
	RequestTimeout  time.Duration `yaml:"request_timeout"`
}

type RDAPConfig struct {
	BootstrapURL string `yaml:"bootstrap_url"`
	MaxRetries   int    `yaml:"max_retries"`
}

type TLSConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type DomainConfig struct {
	Name string `yaml:"name"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := defaultConfig()

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			ListenAddress: ":9118",
		},
		Scheduler: SchedulerConfig{
			RefreshInterval: 6 * time.Hour,
			WorkerCount:     5,
			RequestTimeout:  10 * time.Second,
		},
		RDAP: RDAPConfig{
			BootstrapURL: "https://data.iana.org/rdap/dns.json",
			MaxRetries:   3,
		},
		TLS: TLSConfig{
			Enabled: true,
			Port:    443,
		},
	}
}

func (c *Config) Validate() error {
	if c.Server.ListenAddress == "" {
		return fmt.Errorf("server.listen_address cannot be empty")
	}

	if c.Scheduler.WorkerCount <= 0 {
		return fmt.Errorf("scheduler.worker_count must be greater than zero")
	}

	if c.Scheduler.RefreshInterval <= 0 {
		return fmt.Errorf("scheduler.refresh_interval must be greater than zero")
	}

	if c.Scheduler.RequestTimeout <= 0 {
		return fmt.Errorf("scheduler.request_timeout must be greater than zero")
	}

	if c.RDAP.BootstrapURL == "" {
		return fmt.Errorf("rdap.bootstrap_url cannot be empty")
	}

	if c.RDAP.MaxRetries < 0 {
		return fmt.Errorf("rdap.max_retries cannot be negative")
	}

	if len(c.Domains) == 0 {
		return fmt.Errorf("at least one domain must be configured")
	}

	for i, d := range c.Domains {
		if d.Name == "" {
			return fmt.Errorf("domains[%d].name cannot be empty", i)
		}
	}

	return nil
}
