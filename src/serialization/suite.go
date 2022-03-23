package serialization

import (
	"contract-testing/src/serialization/openapi"
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
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

	AnyOf []*Contract `yaml:"anyOf"`
}

type SpecFile struct {
	Path       string               `yaml:"path"`
	BaseUrl    string               `yaml:"baseUrl"`
	Operations map[string]Operation `yaml:"operations"`
}

type Operation struct {
	Parameters    map[string]string      `yaml:"parameters"`
	ParameterSets []map[string]string    `yaml:"parameterSets"`
	Body          map[string]interface{} `yaml:"body"`
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
	if len(operation.Responses) == 0 {
		return nil, fmt.Errorf("could not find any response in operation %s", operation.OperationId)
	} else if len(operation.Responses) == 1 {
		for statusCode := range operation.Responses {
			if statusCode != "200" {
				log.Println("info: operation", operation, "does not have a 200 response.")
			}
			return NewContractFromOperationWithStatus(url, method, operation, statusCode)
		}
	}

	subcontracts := make([]*Contract, 0)
	for statusCode := range operation.Responses {
		subcontract, err := NewContractFromOperationWithStatus(url, method, operation, statusCode)
		if err != nil {
			return nil, err
		}
		subcontracts = append(subcontracts, subcontract)
	}
	return &Contract{
		Url:        url,
		Method:     method,
		Name:       operation.OperationId,
		AnyOf:      subcontracts,
		Parameters: make(map[string]string, 0),
	}, nil
}

func NewContractFromOperationWithStatus(url string, method string, operation openapi.Operation, statusCode string) (*Contract, error) {
	var statusCodeInt int64
	var err error
	if statusCodeInt, err = strconv.ParseInt(statusCode, 10, 64); err != nil {
		return nil, fmt.Errorf("invalid status code: %s", statusCode)
	}

	if _, found := operation.Responses[statusCode]; !found {
		return nil, fmt.Errorf("could not find %s response in operation %s", statusCode, operation.OperationId)
	}
	if _, found := operation.Responses[statusCode].Content["application/json"]; !found {
		return nil, fmt.Errorf("could not find application/json content in operation %s (%s)", operation.OperationId, statusCode)
	}

	schema := operation.Responses[statusCode].Content["application/json"].Schema

	return &Contract{
		Url:    url,
		Method: method,
		Expect: Expect{
			Status:         int(statusCodeInt),
			SchemaResolved: schema,
			ContentType:    "application/json",
		},
		Name:       fmt.Sprintf("%s[response:%s]", operation.OperationId, statusCode),
		Parameters: make(map[string]string, 0),
	}, nil
}

func (s SpecFile) CreateContracts() ([]Contract, error) {
	doc, err := openapi.LoadDocument(s.Path)
	if err != nil {
		return nil, err
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
		contract.Parameters = deepCopyStringMap(sop.Parameters)

		if sop.ParameterSets == nil {
			sop.ParameterSets = make([]map[string]string, 1)
			sop.ParameterSets[0] = sop.Parameters
		}

		contract.Body = sop.Body
		contract.copyAttributesToChildren()

		for i, parameterSet := range sop.ParameterSets {
			parameterSetContract := contract.deepCopy()
			if len(sop.ParameterSets) > 1 {
				parameterSetContract.UpdateName(fmt.Sprintf("%s[paramSet:%d]", contract.Name, i))
			}

			parameterSetContract.Parameters = deepCopyStringMap(parameterSet)
			parameterSetContract.checkPathParameters(doc.Paths[url].Parameters, operationId)
			parameterSetContract.copyAttributesToChildren()

			contracts = append(contracts, *parameterSetContract)
		}
	}
	return contracts, nil
}

// checkPathParameters checks if all parameters from the path are in the contracts parameters. If the location part of
// the parameter is missing in the Contract, it is added using the information from the path.
func (c *Contract) checkPathParameters(parameters []*openapi.Parameter, operationId string) {
	for _, parameter := range parameters {
		// Check if the contract has the parameter including the location part
		_, found := c.Parameters[string(parameter.In)+":"+parameter.Name]
		if found {
			continue
		}

		// Check if the contract has a parameter with the same name but missing the location part
		_, found = c.Parameters[parameter.Name]
		if found {
			// Copy existing parameter value to new parameter key with the location part
			c.Parameters[string(parameter.In)+":"+parameter.Name] = c.Parameters[parameter.Name]
		}

		if found || !parameter.Required {
			continue
		}
		fmt.Printf("[%s] Missing parameter required %s from operation %s\n", aurora.Yellow("WARN"), parameter.Name, operationId)
	}
}

// copyAttributesToChildren recursively copies Contract.Parameters and Contract.Body to its subcontracts (anyOf)
func (c *Contract) copyAttributesToChildren() {
	if c.AnyOf == nil {
		return
	}
	for _, contract := range c.AnyOf {
		contract.Parameters = c.Parameters
		contract.Body = c.Body

		contract.copyAttributesToChildren()
	}
}

// UpdateName replaces the name of the contract with new. Subcontracts are updated such that the old name of this
// contract is replaced by the new name, but the rest of the subcontract's name is preserved.
func (c *Contract) UpdateName(new string) {
	c.updateName(c.Name, new)
}

func (c *Contract) updateName(old string, new string) {
	if c.AnyOf == nil {
		return
	}
	c.Name = strings.ReplaceAll(c.Name, old, new)
	for _, contract := range c.AnyOf {

		contract.updateName(old, new)
	}
}

func (c *Contract) deepCopy() *Contract {
	copied := &Contract{
		Url:        c.Url,
		Method:     c.Method,
		Headers:    deepCopyStringMap(c.Headers),
		Expect:     c.Expect,
		Name:       c.Name,
		Parameters: deepCopyStringMap(c.Parameters),
		Body:       deepCopyMap(c.Body),
		Debug:      c.Debug,
		AnyOf:      make([]*Contract, len(c.AnyOf)),
	}
	copied.Body = deepCopyMap(c.Body)
	for k, v := range c.AnyOf {
		copied.AnyOf[k] = v.deepCopy()
	}

	return copied
}

func deepCopyInterface(m interface{}) interface{} {
	switch m.(type) {
	case map[string]string:
		log.Println("copy string map")
		return deepCopyStringMap(m.(map[string]string))
	case map[string]interface{}:
		log.Println("copy map")
		return deepCopyMap(m.(map[string]interface{}))
	case []interface{}:
		log.Println("copy array")
		return deepCopyArray(m.([]interface{}))
	default:
		log.Println("copy default")
		return m
	}
}

func deepCopyStringMap(m map[string]string) map[string]string {
	copied := make(map[string]string)
	for k, v := range m {
		copied[k] = v
	}
	return copied
}

func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	copied := make(map[string]interface{})
	for k, v := range m {
		copied[k] = deepCopyInterface(v)
	}
	return copied
}

func deepCopyArray(m []interface{}) []interface{} {
	copied := make([]interface{}, len(m))
	for k, v := range m {
		copied[k] = deepCopyInterface(v)
	}
	return copied
}
