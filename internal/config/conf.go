package config

import (
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
		// TODO log problems
		return conf
	}

	err = yaml.Unmarshal(confFromFile, conf)
	if err != nil {
		// TODO log problems
		return conf
	}
	return conf
}

func defaultConfig() *Config {
	return &Config{
		Address: "localhost:8080",
	}
}
