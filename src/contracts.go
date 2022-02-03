package main

import (
	"contract-testing/src/serialization"
)

type FailureReason string

const (
	FailureHttp       FailureReason = "http"
	FailureHttpStatus FailureReason = "http.status"
)

type ContractResult struct {
	Url     string
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
	cr := ContractResult{Url: contract.Url, Pass: false}

	res, err := RunRequest(suite.Contracts[0].Url, combineHeaders(contract, suite))
	if err != nil {
		cr.Reason = FailureHttp
		return cr
	}

	if res.StatusCode != 200 {
		cr.Reason = FailureHttpStatus
		return cr
	}

	cr.Pass = true
	return cr
}
