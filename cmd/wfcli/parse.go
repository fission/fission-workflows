package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fission/fission-workflows/pkg/parse"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/jsonpb"
	"github.com/urfave/cli"
)

var cmdParse = cli.Command{
	Name:    "parse",
	Aliases: []string{"p"},
	Usage:   "parse <path-to-source> ",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "type, t",
			Value: "yaml",
			Usage: "Indicate which parser plugin to use for the parsing (yaml|pb).",
		},
	},
	Description: "Read YAML definitions to the executable JSON format (deprecated)",
	Action: commandContext(func(ctx Context) error {

		if ctx.NArg() == 0 {
			panic("Need a path to a yaml Workflow definition")
		}

		parserType := ctx.String("type")
		if parserType != "" && !parse.Supports(parserType) {
			fmt.Printf("Unknown parser '%s'; will try all parsers.", parserType)
		}

		for _, path := range ctx.Args() {

			fnName := strings.TrimSpace(path)

			f, err := os.Open(fnName)
			if err != nil {
				panic(err)
			}

			wfSpec, err := parse.ParseWith(f, parserType)
			if err != nil {
				panic(err)
			}

			println(toFormattedJson(wfSpec))
		}
		return nil
	}),
}

func toFormattedJson(spec *types.WorkflowSpec) string {

	marshal := jsonpb.Marshaler{
		Indent: "  ",
	}
	jsonWf, err := marshal.MarshalToString(spec)
	if err != nil {
		panic(err)
	}

	return jsonWf
}
