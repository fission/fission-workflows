package main

import (
	"github.com/urfave/cli"
)

var cmdAdmin = cli.Command{
	Name:  "admin",
	Usage: "Administrative commands",
	Subcommands: []cli.Command{
		cmdStatus,
		cmdVersion,
		{
			Name:  "halt",
			Usage: "Stop the Workflow engine from evaluating anything",
			Action: commandContext(func(ctx Context) error {
				client := getClient(ctx)
				err := client.Admin.Halt(ctx)
				if err != nil {
					panic(err)
				}
				return nil
			}),
		},
		{
			Name:  "resume",
			Usage: "Resume the Workflow engine evaluations",
			Action: commandContext(func(ctx Context) error {
				client := getClient(ctx)
				err := client.Admin.Resume(ctx)
				if err != nil {
					panic(err)
				}
				return nil
			}),
		},
	},
}
