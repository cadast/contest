package main

import (
	"contract-testing/src/serialization"
	"contract-testing/src/serialization/openapi"
	"flag"
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"log"
	"os"
	"strings"
)

func PassWarnFail(i ContractVerdict) aurora.Value {
	if i >= ContractFail {
		return aurora.Red("FAIL")
	} else if i >= ContractWarn {
		return aurora.Yellow("WARN")
	}
	return aurora.Green("PASS")
}

type multiStringFlag []string

func (i *multiStringFlag) String() string {
	return strings.Join(*i, ",")
}

func (i *multiStringFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func checkFilePointer(p *string) {
	if *p == "" {
		fmt.Printf("An empty file name is not allowed.\n")
		os.Exit(1)
	}
	if _, err := os.Stat(*p); err != nil {
		fmt.Printf("The file %s does not exist or is not accessible.\n", *p)
		os.Exit(1)
	}
}

func main() {
	suiteFileP := flag.String("suite", "./contest.yaml", "The path to the suite to run on")
	var schemaFilesP multiStringFlag
	flag.Var(&schemaFilesP, "schema", "Path to an OpenAPI 3.0 schema file (multiple allowed)")
	flag.Parse()

	checkFilePointer(suiteFileP)
	for _, s := range schemaFilesP {
		checkFilePointer(&s)
	}

	suite, err := serialization.LoadSuite(*suiteFileP)
	if err != nil {
		log.Fatalln("Could not load Suite YAML", err)
	}
	if *suiteFileP == "./contest.yaml" {
		fmt.Printf("Using testing suite from contest.yaml.\n\n")
	}

	// Load all schemas from OpenAPI documents
	suite.Schemas = make(map[string]openapi.Schema)
	for _, path := range schemaFilesP {
		doc, err := openapi.LoadDocument(path)
		if err != nil {
			log.Fatalln("Could not load OpenAPI Schema YAML", err)
		}
		for k, v := range doc.Components.Schemas {
			suite.Schemas[k] = *v
		}
	}

	// Load spec files and create contracts for all operations listed
	for _, specFile := range suite.SpecFiles {
		contracts, err := specFile.CreateContracts()
		if err != nil {
			log.Fatalln("Could not create contracts for spec file", specFile.Path, ":", err)
		}

		suite.Contracts = append(suite.Contracts, contracts...)
	}

	var warningFailureReasons []FailureReason
	for failureReason, severity := range suite.Severity {
		if strings.ToLower(severity) == "warn" {
			if warningFailureReasons == nil {
				warningFailureReasons = make([]FailureReason, 0)
			}
			warningFailureReasons = append(warningFailureReasons, FailureReason(failureReason))
		}
	}

	fmt.Printf("Testing %d contracts...\n\n", len(suite.Contracts))

	successfulContracts := 0
	verdict := ContractPass

	for _, contract := range suite.Contracts {
		res := RunContract(contract, *suite)
		pass := res.Pass(&warningFailureReasons)
		verdict |= pass

		if pass < ContractFail {
			successfulContracts++
		}

		postfix := ""
		if len(res.Failures) > 0 {
			for _, failure := range res.Failures {
				postfix += failure.String() + "; "
			}
			postfix = strings.TrimSuffix(postfix, "; ")
			postfix = " " + aurora.Faint("("+postfix+")").String()
		}
		fmt.Printf("[%s] %s%s\n", PassWarnFail(pass), res.Name, postfix)
	}

	fmt.Println()
	fmt.Printf("%d/%d contracts passed.\n", successfulContracts, len(suite.Contracts))
	fmt.Printf("Final verdict: %s\n", aurora.Bold(PassWarnFail(verdict)))

	if successfulContracts < len(suite.Contracts) {
		os.Exit(1)
	}
}
