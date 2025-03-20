// Copyright 2019 Setin Sergei
// Licensed under the Apache License, Version 2.0 (the "License")

package mosh

import (
	"moshcast/log"
	"os"

	"gopkg.in/yaml.v3"
)

// Options represents the server configuration parsed from YAML
type Options struct {
	Name     string `yaml:"Name"`
	Admin    string `yaml:"Admin,omitempty"`
	Location string `yaml:"Location,omitempty"`
	Host     string `yaml:"Host"`

	Socket struct {
		Port int `yaml:"Port"`
	} `yaml:"Socket"`

	Limits struct {
		Clients                int32 `yaml:"Clients"`
		Sources                int32 `yaml:"Sources"`
		SourceIdleTimeOut      int   `yaml:"SourceIdleTimeOut"`
		EmptyBufferIdleTimeOut int   `yaml:"EmptyBufferIdleTimeOut"`
		WriteTimeOut           int   `yaml:"WriteTimeOut"`
	} `yaml:"Limits"`

	Auth struct {
		DBFile string `yaml:"DBFile"`
	} `yaml:"Auth"`

	Paths struct {
		Base    string `yaml:"Base"`
		Web     string `yaml:"Web"`
		Log     string `yaml:"Log"`
		Plugins string `yaml:"Plugins"`
	} `yaml:"Paths"`

	Logging struct {
		LogLevel log.LogsLevel `yaml:"LogLevel"`
		LogSize  int           `yaml:"LogSize"`
	} `yaml:"Logging"`

	Mounts []*mount `yaml:"Mounts"`
}

// Load reads the configuration from config.yaml into the Options struct
func (o *Options) Load() error {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, o)
}
