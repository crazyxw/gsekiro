package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type WebConfig struct {
	VKey          string `yaml:"vKey"`
	InvokeTimeout int    `yaml:"invokeTimeout"`
	Port          string `yaml:"port"`
}

type LogConfig struct {
	Filename string `yaml:"filename"`
	MaxAge   int    `yaml:"maxAge"`
}

type Config struct {
	Web WebConfig `yaml:"web"`
	Log LogConfig `yaml:"log"`
}

func (c *Config) loadFromFile() error {
	config, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(config, c)
	if err != nil {
		return err
	}
	return nil
}
