package main

import (
	"contract-testing/src/serialization"
	"contract-testing/src/serialization/openapi"
	"io/ioutil"
	"strings"
)

type FailureReason string

const (
	FailureContract    FailureReason = "contract"                // An invalid contract
	FailureHttp        FailureReason = "http"                    // An error while running the HTTP request
	FailureIO          FailureReason = "io"                      // An error while fetching the data
	FailureHttpStatus  FailureReason = "unexpected.status"       // An unexpected HTTP status code
	FailureFormat      FailureReason = "format"                  // An invalid response format
	FailureSchema      FailureReason = "unexpected.schema"       // An invalid response Schema
	FailureContentType FailureReason = "unexpected.content-type" // An unexpected content type
)

type ContractResult struct {
	Name    string
	Pass    bool
	Reason  FailureReason
	Comment string
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

func RunContract(contract serialization.Contract, suite serialization.Suite) ContractResult {
	if strings.HasPrefix(contract.Url, "file://") {
		return runFileContract(contract, suite)
	}
	return runHttpContract(contract, suite)
}

func runFileContract(contract serialization.Contract, suite serialization.Suite) ContractResult {
	cr := ContractResult{Name: contract.Name, Pass: false}
	if cr.Name == "" {
		cr.Name = contract.Url
	}

	content, err := ioutil.ReadFile(strings.TrimPrefix(contract.Url, "file://"))
	if err != nil {
		cr.Reason = FailureIO
		return cr
	}

	if contract.Expect.SchemaName != "" || contract.Expect.SchemaResolved != nil {
		cr.Reason, cr.Comment = checkSchemaOnJson(content, contract, suite)
		if cr.Reason != "" {
			return cr
		}
	}

	cr.Pass = true
	return cr
}

func runHttpContract(contract serialization.Contract, suite serialization.Suite) ContractResult {
	cr := ContractResult{Name: contract.Name, Pass: false}
	if cr.Name == "" {
		cr.Name = contract.Url
	}

	res, err := RunRequest(contract.Url, combineHeaders(contract, suite))
	if err != nil {
		cr.Reason = FailureHttp
		return cr
	}

	if res.StatusCode != 200 && (contract.Expect.Status == 0 || contract.Expect.Status != res.StatusCode) {
		cr.Reason = FailureHttpStatus
		return cr
	}

	if contract.Expect.ContentType != "" && !strings.HasPrefix(res.ContentType, contract.Expect.ContentType+";") {
		cr.Reason = FailureContentType
		cr.Comment = "got \"" + res.ContentType + "\" not \"" + contract.Expect.ContentType + "\""
		return cr
	}

	if contract.Expect.SchemaName != "" || contract.Expect.SchemaResolved != nil {
		cr.Reason, cr.Comment = checkSchemaOnJson(res.Body, contract, suite)
		if cr.Reason != "" {
			return cr
		}
	}

	cr.Pass = true
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
