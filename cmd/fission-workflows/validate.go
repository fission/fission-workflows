package main

import (
	"fmt"
	"os"

	"github.com/fission/fission-workflows/pkg/parse/protobuf"
	"github.com/fission/fission-workflows/pkg/parse/yaml"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/jsonpb"

	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/urfave/cli"
)

var cmdValidate = cli.Command{
	Name:        "validate",
	Usage:       "Validate [file ...]",
	Description: "Validate a workflow definition",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "type, t",
			Value: "yaml",
			Usage: "encoding of the file(s) [yaml|proto|json]",
		},
	},
	Action: commandContext(func(ctx Context) error {
		// Get path from args
		if ctx.NArg() == 0 {
			fail("No file provided.")
		}

		var failed bool
		for _, path := range ctx.Args() {
			if err := validateWorkflowDefinition(path, ctx.String("type")); err != nil {
				if _, err := fmt.Fprintf(os.Stderr, "%s: %s\n", path, err.Error()); err != nil {
					panic(err)
				}
				failed = true
			}
		}

		if failed {
			os.Exit(1)
		}

		return nil
	}),
}

func validateWorkflowDefinition(path string, fType string) error {
	// Get file
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Read file into workflowSpec (assume yaml for now)
	var spec *types.WorkflowSpec
	switch fType {
	case "yaml":
		spec, err = yaml.Parse(file)
		if err != nil {
			return fmt.Errorf("failed to parse yaml definition: %v", err)
		}
	case "proto":
		spec, err = protobuf.Parse(file)
		if err != nil {
			return fmt.Errorf("failed to parse protobuf definition: %v", err)
		}
	case "json":
		err := jsonpb.Unmarshal(file, spec)
		if err != nil {
			return fmt.Errorf("failed to parse json definition: %v", err)
		}
	default:
		return fmt.Errorf("unsupported workflow definition format: %v", fType)
	}

	// Validate workflowSpec
	err = validate.WorkflowSpec(spec)
	if err != nil {
		invalid, ok := err.(validate.Error)
		if ok {
			return fmt.Errorf(validate.Format(invalid))
		} else {
			return fmt.Errorf("unknown error: %v", err)
		}
	}
	return nil
}
