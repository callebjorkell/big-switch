package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ReleaseManager struct {
		Url    string `yaml:"url"`
		Token  string `yaml:"token"`
		Caller string `yaml:"caller"`
	} `yaml:"releaseManager"`
	Services []struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
		Color     uint32 `yaml:"color"`
	} `yaml:"services"`
}

func (c Config) ColorMap() map[string]uint32 {
	colors := make(map[string]uint32)
	for _, service := range c.Services {
		colors[service.Name] = service.Color
	}
	return colors
}

func parseConfig(content []byte) (*Config, error) {
	c := &Config{}
	err := yaml.Unmarshal(content, c)
	if err != nil {
		return nil, err
	}

	if c.ReleaseManager.Token == "" {
		return nil, fmt.Errorf("release manager token is missing")
	}
	if c.ReleaseManager.Url == "" {
		return nil, fmt.Errorf("release manager URL is missing")
	}
	if c.ReleaseManager.Caller == "" {
		return nil, fmt.Errorf("release manager caller is missing")
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
