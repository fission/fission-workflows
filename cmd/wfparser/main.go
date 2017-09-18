package main

import (
	"os"

	"io/ioutil"

	"strings"

	"fmt"
	"github.com/fission/fission-workflow/pkg/api/workflow/parse/yaml"
	"github.com/gogo/protobuf/jsonpb"
)

// Temporary cli tool to parse yaml workflow definitions
func main() {

	if len(os.Args) == 0 {
		panic("Need a path to a .yaml")
	}

	fnName := strings.TrimSpace(os.Args[1])
	fmt.Printf("Formatting: '%s'\n", fnName)

	if !strings.HasSuffix(fnName, "yaml") {
		panic("Only YAML workflow definitions are supported")
	}

	f, err := os.Open(fnName)
	if err != nil {
		panic(err)
	}

	wfDef, err := yaml.Parse(f)
	if err != nil {
		panic(err)
	}

	wfSpec, err := yaml.Transform(wfDef)
	if err != nil {
		panic(err)
	}

	marshal := jsonpb.Marshaler{
		Indent: "  ",
	}
	jsonWf, err := marshal.MarshalToString(wfSpec)
	if err != nil {
		panic(err)
	}

	outputFile := strings.Replace(fnName, "yaml", "json", -1)

	err = ioutil.WriteFile(outputFile, []byte(jsonWf), 0644)
	if err != nil {
		panic(err)
	}

	println(outputFile)
}
