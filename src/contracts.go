package main

import (
	"contract-testing/src/serialization"
	"contract-testing/src/serialization/openapi"
	"fmt"
	"io/ioutil"
	"strings"
)

type FailureReason string

const (
	FailureDebug        FailureReason = "debug"                   // Debug output
	FailureContract     FailureReason = "contract"                // An invalid contract
	FailureHttp         FailureReason = "http"                    // An error while running the HTTP request
	FailureIO           FailureReason = "io"                      // An error while fetching the data
	FailureHttpStatus   FailureReason = "unexpected.status"       // An unexpected HTTP status code
	FailureFormat       FailureReason = "format"                  // An invalid response format
	FailureSchema       FailureReason = "unexpected.schema"       // An invalid response Schema
	FailureContentType  FailureReason = "unexpected.content-type" // An unexpected content type
	FailureResponseTime FailureReason = "unexpected.responseTime" // The response time was longer than expected
)

type Failure struct {
	Reason  FailureReason
	Comment string
}

func (f Failure) String() string {
	if f.Comment != "" {
		return string(f.Reason) + ": " + f.Comment
	}
	return string(f.Reason)
}

type ContractResult struct {
	Name     string
	Failures []Failure
}

type ContractVerdict int

const (
	ContractPass ContractVerdict = 0b000
	ContractWarn ContractVerdict = 0b001
	ContractFail ContractVerdict = 0b010
)

func NewContractResult(name string) ContractResult {
	return ContractResult{
		Name:     name,
		Failures: make([]Failure, 0),
	}
}

func (c *ContractResult) failure(reason FailureReason, comment string) {
	if reason == "" {
		return
	}
	c.Failures = append(c.Failures, Failure{
		Reason:  reason,
		Comment: comment,
	})
}

func (c ContractResult) Pass(warningFailures *[]FailureReason) ContractVerdict {
	if warningFailures == nil && len(c.Failures) > 0 {
		return ContractFail
	}

	result := ContractPass

outer:
	for _, failure := range c.Failures {
		if failure.Reason == FailureDebug {
			continue
		}

		// Check if the failure is a warning only
		for _, warningFailure := range *warningFailures {
			if failure.Reason == warningFailure {
				result |= ContractWarn
				continue outer
			}
		}

		// If the failure is not only a warning, the contract has not passed
		result |= ContractFail
	}
	return result
}

func combineHeaders(contract serialization.Contract, suite serialization.Suite) map[string]string {
	headers := make(map[string]string)
	for k, v := range suite.Headers {
		headers[k] = v
	}
	for k, v := range contract.Headers {
		headers[k] = v
	}
	return headers
}

func RunContract(contract serialization.Contract, suite serialization.Suite, warningFailures *[]FailureReason) ContractResult {
	if len(contract.AnyOf) > 0 {
		failures := make([]Failure, 0)
		for _, subcontract := range contract.AnyOf {
			cr := RunContract(*subcontract, suite, warningFailures)
			if cr.Pass(warningFailures) <= ContractWarn {
				return cr
			}
			failures = append(failures, cr.Failures...)
		}
		return ContractResult{
			Name:     contract.Name,
			Failures: failures,
		}
	}
	if strings.HasPrefix(contract.Url, "file://") {
		return runFileContract(contract, suite)
	}
	return runHttpContract(contract, suite)
}

func runFileContract(contract serialization.Contract, suite serialization.Suite) ContractResult {
	cr := NewContractResult(contract.Name)
	if cr.Name == "" {
		cr.Name = contract.Url
	}

	content, err := ioutil.ReadFile(strings.TrimPrefix(contract.Url, "file://"))
	if err != nil {
		cr.failure(FailureIO, "")
		return cr
	}

	if contract.Expect.SchemaName != "" || contract.Expect.SchemaResolved != nil {
		cr.failure(checkSchemaOnJson(content, contract, suite))
		return cr
	}

	return cr
}

func runHttpContract(contract serialization.Contract, suite serialization.Suite) ContractResult {
	headers := combineHeaders(contract, suite)

	for key, value := range contract.Parameters {
		if strings.HasPrefix(key, "path:") {
			name := strings.TrimPrefix(key, "path:")
			contract.Url = strings.ReplaceAll(contract.Url, "{"+name+"}", value)
		} else if strings.HasPrefix(key, "header:") {
			name := strings.TrimPrefix(key, "header:")
			for headerKey, headerValue := range headers {
				headers[headerKey] = strings.ReplaceAll(headerValue, "{"+name+"}", value)
			}
		}
	}
	cr := NewContractResult(contract.Name)
	if contract.Name == "" {
		cr.Name = contract.Url
	} else {
		cr.Name = fmt.Sprintf("%s (%s)", contract.Name, contract.Url)
	}

	var body []byte
	if contract.Body != nil {
		var err error
		body, err = JsonMarshal(contract.Body)
		if err != nil {
			cr.failure(FailureContract, err.Error())
			return cr
		}
	}

	res, err := RunRequest(contract.Url, headers, body)
	if err != nil {
		cr.failure(FailureHttp, err.Error())
		return cr
	}

	if res.StatusCode != 200 && (contract.Expect.Status == 0 || contract.Expect.Status != res.StatusCode) {
		cr.failure(FailureHttpStatus, fmt.Sprintf("got %d not %d", res.StatusCode, contract.Expect.Status))
		return cr
	}

	if contract.Expect.ContentType != "" && !strings.HasPrefix(res.ContentType+";", contract.Expect.ContentType+";") {
		cr.failure(FailureContentType, fmt.Sprintf("got %s not %s", res.ContentType, contract.Expect.ContentType))
	}

	if contract.Debug {
		cr.failure(FailureDebug, string(res.Body))
	}

	if contract.Expect.SchemaName != "" || contract.Expect.SchemaResolved != nil {
		cr.failure(checkSchemaOnJson(res.Body, contract, suite))
	}

	if contract.Expect.ResponseTime > 0 && res.ResponseTime > contract.Expect.ResponseTime {
		cr.failure(FailureResponseTime, fmt.Sprintf("took %dms not %dms", res.ResponseTime, contract.Expect.ResponseTime))
	}

	return cr
}

// createArraySchema creates a new Schema of type array with the schema of the given name as Items.
// The suffix `[]` is trimmed from the given schemaName.
func createArraySchema(schemaName string, suite serialization.Suite) (openapi.Schema, bool) {
	schemaName = strings.TrimSuffix(schemaName, "[]")

	var itemSchema openapi.Schema
	itemSchema, found := suite.Schemas[schemaName]

	schema := openapi.Schema{
		Items: &itemSchema,
		Type:  openapi.SchemaTypeArray,
	}
	return schema, found
}

func checkSchemaOnJson(data []byte, contract serialization.Contract, suite serialization.Suite) (FailureReason, string) {
	schema, found := suite.Schemas[contract.Expect.SchemaName]
	if contract.Expect.SchemaResolved != nil {
		schema = *contract.Expect.SchemaResolved
		found = true
	}

	// Create a new array schema; possible future optimization is creating array schemas when loading the suite.
	if strings.HasSuffix(contract.Expect.SchemaName, "[]") {
		schema, found = createArraySchema(contract.Expect.SchemaName, suite)
	}

	// Check if the schema specified in the contract was found
	if !found {
		return FailureContract, ""
	}

	json, err := JsonUnmarshal(data)
	// Check if data was valid JSON
	if err != nil {
		return FailureFormat, ""
	}

	if schema.Title == "" {
		schema.Title = "root"
	}

	// Check for valid JSON schema
	messages := make([]string, 0)
	if valid := CheckSchema(schema, json, schema.Title, &messages); !valid {
		return FailureSchema, strings.Join(messages, ", ")
	}

	return "", ""
}
