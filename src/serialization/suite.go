package serialization

import (
	"contract-testing/src/serialization/openapi"
	"errors"
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Expect struct {
	Status         int    `yaml:"status"`
	SchemaName     string `yaml:"schema"`
	ContentType    string `yaml:"contentType"`
	SchemaResolved *openapi.Schema
	ResponseTime   int64 `yaml:"responseTime"`
}

type Contract struct {
	Url        string                 `yaml:"url"`
	Method     string                 `yaml:"method"`
	Headers    map[string]string      `yaml:"headers"`
	Expect     Expect                 `yaml:"expect"`
	Name       string                 `yaml:"name"`
	Parameters map[string]string      `yaml:"parameters"`
	Body       map[string]interface{} `yaml:"body"`
	Debug      bool                   `yaml:"debug"`
}

type SpecFile struct {
	Path       string               `yaml:"path"`
	BaseUrl    string               `yaml:"baseUrl"`
	Operations map[string]Operation `yaml:"operations"`
}

type Operation struct {
	Parameters map[string]string      `yaml:"parameters"`
	Body       map[string]interface{} `yaml:"body"`
}

type Suite struct {
	SpecFiles []SpecFile        `yaml:"specFiles"`
	Contracts []Contract        `yaml:"contracts"`
	Headers   map[string]string `yaml:"headers"`
	Schemas   map[string]openapi.Schema
	Severity  map[string]string `yaml:"severity"`
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
		Name:       operation.OperationId,
		Parameters: make(map[string]string, 0),
	}, nil
}

func (s SpecFile) CreateContracts() ([]Contract, error) {
	doc, err := openapi.LoadDocument(s.Path)
	if err != nil {
		return nil, errors.New("could not load OpenAPI schema file")
	}

	contracts := make([]Contract, 0, len(s.Operations))

	for operationId, sop := range s.Operations {
		url, method, op, found := doc.FindOperationById(operationId)
		if !found {
			return nil, fmt.Errorf("operation %s not found", operationId)
		}

		contract, err := NewContractFromOperation(s.BaseUrl+url, method, *op)
		if err != nil {
			return nil, err
		}

		// Copy parameters from the spec file operation to the contract
		for key, value := range sop.Parameters {
			contract.Parameters[key] = value
		}

		// Check if all parameters from the path are in the contracts parameters
		for _, parameter := range doc.Paths[url].Parameters {
			// Check if the contract has the parameter including the location part
			_, found := contract.Parameters[string(parameter.In)+":"+parameter.Name]
			if found {
				continue
			}

			// Check if the contract has a parameter with the same name but missing the location part
			_, found = contract.Parameters[parameter.Name]
			if found {
				// Copy existing parameter value to new parameter key with the location part
				contract.Parameters[string(parameter.In)+":"+parameter.Name] = contract.Parameters[parameter.Name]
			}

			if found || !parameter.Required {
				continue
			}
			fmt.Printf("[%s] Missing parameter required %s from operation %s\n", aurora.Yellow("WARN"), parameter.Name, operationId)
		}

		contract.Body = sop.Body

		contracts = append(contracts, *contract)
	}
	return contracts, nil
}
