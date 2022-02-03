package serialization

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Contract struct {
	Url     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
}

type Suite struct {
	Contracts []Contract        `yaml:"contracts"`
	Headers   map[string]string `yaml:"headers"`
}

type wrapper struct {
	Suite Suite `yaml:"suite"`
}

func LoadSuite(path string) (*Suite, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	wrapper := wrapper{}
	err = yaml.Unmarshal(content, &wrapper)
	if err != nil {
		return nil, err
	}
	return &wrapper.Suite, nil
}
