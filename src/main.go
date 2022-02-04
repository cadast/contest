package main

import (
	"contract-testing/src/serialization"
	"contract-testing/src/serialization/openapi"
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"log"
	"os"
)

func PassFail(b bool) aurora.Value {
	if b {
		return aurora.Green("PASS")
	}
	return aurora.Red("FAIL")
}

func main() {
	suite, err := serialization.LoadSuite("./../contract.yaml")
	if err != nil {
		log.Fatalln("Could not load Suite YAML", err)
	}

	doc, err := openapi.LoadDocument("./../products.yaml")
	if err != nil {
		log.Fatalln("Could not load OpenAPI Schema YAML", err)
	}
	suite.Schemas = doc.Components.Schemas

	fmt.Printf("Testing %d contracts...\n\n", len(suite.Contracts))
	successfulContracts := 0
	for _, contract := range suite.Contracts {
		res := RunContract(contract, *suite)

		if res.Pass {
			successfulContracts++
		}

		postfix := aurora.Reset("")
		if !res.Pass {
			postfix = aurora.Faint(fmt.Sprintf(" (%s)", res.Reason))
		}
		fmt.Printf("[%s] %s%s\n", PassFail(res.Pass), res.Name, postfix)
	}

	fmt.Println()
	fmt.Printf("%d/%d contracts passed.\n", successfulContracts, len(suite.Contracts))
	fmt.Printf("Final verdict: %s\n", aurora.Bold(PassFail(successfulContracts == len(suite.Contracts))))

	if successfulContracts < len(suite.Contracts) {
		os.Exit(1)
	}
}
