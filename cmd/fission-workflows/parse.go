package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fission/fission-workflows/pkg/parse"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/jsonpb"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var cmdParse = cli.Command{
	Name:  "parse",
	Usage: "parse <path-to-workflow-file> ",
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
			log.Fatal("No file provided.")
		}

		parserType := ctx.String("type")
		if parserType != "" && !parse.Supports(parserType) {
			log.Fatalf("Unknown parser '%s'", parserType)
		}

		for _, path := range ctx.Args() {

			fnName := strings.TrimSpace(path)

			f, err := os.Open(fnName)
			if err != nil {
				f.Close()
				panic(err)
			}

			wfSpec, err := parse.ParseWith(f, parserType)
			f.Close()
			if err != nil {
				panic(err)
			}
			fmt.Println(toFormattedJSON(wfSpec))
		}
		return nil
	}),
}

func toFormattedJSON(spec *types.WorkflowSpec) string {

	marshal := jsonpb.Marshaler{
		Indent: "  ",
	}
	jsonWf, err := marshal.MarshalToString(spec)
	if err != nil {
		panic(err)
	}

	return jsonWf
}
