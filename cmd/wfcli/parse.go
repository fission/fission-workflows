package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fission/fission-workflows/pkg/parse/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/urfave/cli"
)

var cmdParse = cli.Command{
	Name:        "parse",
	Aliases:     []string{"p"},
	Usage:       "parse <path-to-yaml> ",
	Description: "Read YAML definitions to the executable JSON format (deprecated)",
	Action: func(c *cli.Context) error {

		if c.NArg() == 0 {
			panic("Need a path to a yaml workflow definition")
		}

		for _, path := range c.Args() {

			fnName := strings.TrimSpace(path)

			f, err := os.Open(fnName)
			if err != nil {
				panic(err)
			}

			wfSpec, err := yaml.Parse(f)
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

			fmt.Println(jsonWf)
		}
		return nil
	},
}
