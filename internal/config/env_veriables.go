package config

import (
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"platform/domain-exporter/internal/model"
)

type ENV struct {
	Port string
}

func EnvConfig() ENV {

	var env ENV

	env.Port = os.Getenv("APP_PORT")
	if env.Port == "" {

		env.Port = "8080"
	}

	return env

}

func LoadConfig() (*model.Config, error) {

	path := "configs/config.yaml"
	data, err := os.ReadFile(path)
	if err != nil {

		slog.Error("configuration Loading Failed", "Path", path, "Error", err)
		return nil, err
	}
	var cfg model.Config

	if err = yaml.Unmarshal(data, &cfg); err != nil {

		slog.Error("Reading yaml Failed", "Error", err)
		return nil, err
	}

	return &cfg, nil

}
