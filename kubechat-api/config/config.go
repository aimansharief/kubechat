package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Cluster struct {
	Name            string   `yaml:"name"`
	Kubeconfig      string   `yaml:"kubeconfig"`
	AllowedCommands []string `yaml:"allowed_commands"`
	ReadOnly        bool     `yaml:"read_only"`
}

type Config struct {
	Clusters []Cluster `yaml:"clusters"`
}

func Load(path string) Config {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}
	return cfg
}
