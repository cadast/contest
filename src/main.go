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

func PassFail(b bool) aurora.Value {
	if b {
		return aurora.Green("PASS")
	}
	return aurora.Red("FAIL")
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
	suiteFileP := flag.String("suite", "", "The path to the suite to run on")
	var schemaFilesP multiStringFlag
	flag.Var(&schemaFilesP, "schema", "Path to an OpenAPI 3.0 schema file (multiple allowed)")
	flag.Parse()

	checkFilePointer(suiteFileP)
	if len(schemaFilesP) == 0 {
		fmt.Printf("You need to supply at least one schema file.\n")
		os.Exit(1)
	}
	for _, s := range schemaFilesP {
		checkFilePointer(&s)
	}

	suite, err := serialization.LoadSuite(*suiteFileP)
	if err != nil {
		log.Fatalln("Could not load Suite YAML", err)
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
			log.Fatalln("Could not create contracts for spec file ", specFile.Path, ": ", err)
		}

		suite.Contracts = append(suite.Contracts, contracts...)
	}

	fmt.Printf("Testing %d contracts...\n\n", len(suite.Contracts))
	successfulContracts := 0
	for _, contract := range suite.Contracts {
		res := RunContract(contract, *suite)

		if res.Pass {
			successfulContracts++
		}

		postfix := aurora.Reset("")
		if !res.Pass {
			comment := ": " + res.Comment
			if len(comment) == 2 {
				comment = ""
			}
			postfix = aurora.Faint(fmt.Sprintf(" (%s%s)", res.Reason, comment))
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
