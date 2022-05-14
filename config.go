package main

import (
	"bytes"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Commands []Command `yaml:"commands"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Commands, validation.Required),
	)
}

type Command struct {
	Command  string `yaml:"command"`
	ApiUrl   string `yaml:"apiUrl"`
	Template string `yaml:"template"`
}

func (c Command) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Command, validation.Required, is.Alphanumeric),
		validation.Field(&c.ApiUrl, validation.Required, is.URL),
		validation.Field(&c.Template, validation.Required),
	)
}

func parseConfig(contents []byte) (*Config, error) {
	cfg := &Config{}

	contents = bytes.TrimPrefix(contents, []byte("---"))

	if err := yaml.Unmarshal(contents, &cfg); err != nil {
		return cfg, err
	}

	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}
