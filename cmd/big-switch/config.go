package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	ReleaseManager struct {
		Token string `yaml:"token"`
	} `yaml:"releaseManager"`
	Services []struct {
		Name  string `yaml:"name"`
		Color uint32 `yaml:"color"`
	} `yaml:"services"`
}

func readConfig() (*Config, error) {
	c := &Config{}
	bytes, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(bytes, c)
	if err != nil {
		panic(err)
	}

	if c.ReleaseManager.Token == "" {
		return nil, fmt.Errorf("release manager token is missing")
	}
	for i, service := range c.Services {
		if len(service.Name) < 1 {
			return nil, fmt.Errorf("name of service must be specified for entry %d", i)
		}
		if service.Color == 0 {
			return nil, fmt.Errorf("color of service must be specified for entry %d", i)
		}
	}

	return c, nil
}
