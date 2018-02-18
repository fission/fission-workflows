package main

import (
	"fmt"
	"os"

	"github.com/fission/fission-workflows/pkg/api/workflow/parse/yaml"
	"github.com/urfave/cli"
)

// TODO also validate with backend (optional)
var cmdValidate = cli.Command{
	Name:        "validate",
	Usage:       "Validate <file>",
	Description: "Validate a workflow",
	Action: func(c *cli.Context) error {
		// Get path from args
		if c.NArg() == 0 {
			fail("No file provided.")
		}

		var failed bool
		for _, path := range c.Args() {

			printErr := func(msg string) {
				fmt.Fprintf(os.Stderr, "%s: %s\n", path, msg)
				failed = true
			}

			// Get file
			file, err := os.Open(path)
			if err != nil {
				printErr(fmt.Sprintf("Failed to read file: %v", err))
				continue
			}

			// Read file into WorkflowSpec (assume yaml for now)
			_, err = yaml.Parse(file)
			if err != nil {
				printErr(fmt.Sprintf("Failed to parse yaml definition: %v", err))
				continue
			}
		}

		if failed {
			os.Exit(1)
		}

		return nil
	},
}
