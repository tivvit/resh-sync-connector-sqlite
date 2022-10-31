package config

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Address string `yaml:"address"`
}

func New(confPath string) *Config {
	conf := defaultConfig()

	confFromFile, err := os.ReadFile(confPath)
	if err != nil {
		log.Error().Err(err).
			Str("config file", confPath).
			Msg("loading configuration file failed")
		return conf
	}

	err = yaml.Unmarshal(confFromFile, conf)
	if err != nil {
		log.Error().Err(err).Msg("unmarshalling YAML failed")
		return conf
	}
	return conf
}

func defaultConfig() *Config {
	return &Config{
		Address: "localhost:8080",
	}
}
