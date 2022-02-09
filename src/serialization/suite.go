package serialization

import (
	"contract-testing/src/serialization/openapi"
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Expect struct {
	Status         int    `yaml:"status"`
	SchemaName     string `yaml:"schema"`
	ContentType    string `yaml:"contentType"`
	SchemaResolved *openapi.Schema
}

type Contract struct {
	Url     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Expect  Expect            `yaml:"expect"`
	Name    string            `yaml:"name"`
}

type SpecFile struct {
	Path       string               `yaml:"path"`
	BaseUrl    string               `yaml:"baseUrl"`
	Operations map[string]Operation `yaml:"operations"`
}

type Operation struct {
}

type Suite struct {
	SpecFiles []SpecFile        `yaml:"specFiles"`
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

func NewContractFromGet200Operation(url string, path openapi.Path) (*Contract, error) {
	if _, found := path.Operations["get"]; !found {
		return nil, errors.New("could not find 200 response in path " + url)
	}
	return NewContractFromOperation(url, "get", path.Operations["get"])
}

func NewContractFromOperation(url string, method string, operation openapi.Operation) (*Contract, error) {
	if _, found := operation.Responses["200"]; !found {
		return nil, errors.New("could not find 200 response in operation " + operation.OperationId)
	}
	if _, found := operation.Responses["200"].Content["application/json"]; !found {
		return nil, errors.New("could not find application/json content in operation " + operation.OperationId)
	}
	schema := operation.Responses["200"].Content["application/json"].Schema
	return &Contract{
		Url:     url,
		Method:  method,
		Headers: nil,
		Expect: Expect{
			Status:         200,
			SchemaResolved: schema,
			ContentType:    "application/json",
		},
		Name: operation.OperationId,
	}, nil
}
