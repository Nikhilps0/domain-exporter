package model

import (
	"time"
)

type Config struct {
	ServerPort      string        `yaml:"server_port"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
	DomainTargets   []string      `yaml:"domains"`
}
