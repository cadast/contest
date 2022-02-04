package main

import (
	"contract-testing/src/serialization"
	"contract-testing/src/serialization/openapi"
	"strings"
)

type FailureReason string

const (
	FailureContract   FailureReason = "contract"
	FailureHttp       FailureReason = "http"
	FailureHttpStatus FailureReason = "http.status"
	FailureFormat     FailureReason = "format"
	FailureSchema     FailureReason = "schema"
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
	cr := ContractResult{Name: contract.Url, Pass: false}

	res, err := RunRequest(contract.Url, combineHeaders(contract, suite))
	if err != nil {
		cr.Reason = FailureHttp
		return cr
	}

	if res.StatusCode != 200 && (contract.Expect.Status == 0 || contract.Expect.Status != res.StatusCode) {
		cr.Reason = FailureHttpStatus
		return cr
	}

	if contract.Expect.Schema != "" {
		schema, found := suite.Schemas[contract.Expect.Schema]
		if strings.HasSuffix(contract.Expect.Schema, "[]") {
			schemaName := strings.TrimSuffix(contract.Expect.Schema, "[]")

			var itemSchema openapi.Schema
			itemSchema, found = suite.Schemas[schemaName]

			schema = openapi.Schema{
				Items: &itemSchema,
				Type:  openapi.SchemaTypeArray,
			}
		}

		if !found {
			cr.Reason = FailureContract
			return cr
		}

		json, err := JsonUnmarshal(res.Body)
		if err != nil {
			cr.Reason = FailureFormat
			return cr
		}

		if !CheckSchema(schema, json, schema.Title) {
			cr.Reason = FailureSchema
			return cr
		}
	}

	cr.Pass = true
	return cr
}
