package serialization

import (
	"contract-testing/src/serialization/openapi"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Expect struct {
	Status int    `yaml:"status"`
	Schema string `yaml:"schema"`
}

type Contract struct {
	Url     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Expect  Expect            `yaml:"expect"`
	Name    string            `yaml:"name"`
}

type Suite struct {
	Contracts []Contract        `yaml:"contracts"`
	Headers   map[string]string `yaml:"headers"`
	Schemas   map[string]openapi.Schema
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
