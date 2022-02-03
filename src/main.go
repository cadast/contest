package main

import (
	"contract-testing/src/serialization"
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"log"
	"os"
)

func main() {
	suite, err := serialization.LoadSuite("./../contract.yaml")
	if err != nil {
		log.Fatalln("Could not load Suite YAML", err)
	}
	log.Println(suite.Contracts)

	fmt.Printf("Testing %d contracts.\n", len(suite.Contracts))
	successfulContracts := 0
	for _, contract := range suite.Contracts {
		res := RunContract(contract, *suite)

		prefix := aurora.Red("FAIL")
		if res.Pass {
			prefix = aurora.Green("PASS")
			successfulContracts++
		}

		postfix := aurora.Reset("")
		if !res.Pass {
			postfix = aurora.Faint(fmt.Sprintf(" (%s)", res.Reason))
		}
		fmt.Printf("[%s] %s%s\n", prefix, res.Url, postfix)
	}

	verdict := aurora.Bold(aurora.Red("FAIL"))
	if successfulContracts == len(suite.Contracts) {
		verdict = aurora.Bold(aurora.Green("PASS"))
	}
	fmt.Println()
	fmt.Printf("%d/%d contracts passed.\n", successfulContracts, len(suite.Contracts))
	fmt.Printf("Final verdict: %s\n", verdict)

	if successfulContracts < len(suite.Contracts) {
		os.Exit(1)
	}
}
