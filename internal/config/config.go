package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
    Server ServerConfig `yaml:"server"`
    Redis  RedisConfig  `yaml:"redis"`
}

type ServerConfig struct {
    Host    string        `yaml:"host"`
    Port    int           `yaml:"port"`
    Timeout TimeoutConfig `yaml:"timeout"`
}

type TimeoutConfig struct {
    Server int `yaml:"server"`
    Read   int `yaml:"read"`
    Write  int `yaml:"write"`
    Idle   int `yaml:"idle"`
}

type RedisConfig struct {
    Addr     string `yaml:"addr"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
}

func Load(path string) (Config, error) {
    var cfg Config
    data, err := os.ReadFile(path)
    if err != nil {
        return cfg, err
    }

    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return cfg, err
    }

    return cfg, nil
}

