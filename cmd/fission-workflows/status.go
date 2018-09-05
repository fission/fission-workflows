package main

import (
	"fmt"

	"github.com/urfave/cli"
)

var cmdStatus = cli.Command{
	Name:    "status",
	Aliases: []string{"s"},
	Usage:   "Check workflow engine status",
	Action: commandContext(func(ctx Context) error {
		client := getClient(ctx)
		resp, err := client.Admin.Status(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf(resp.Status)

		return nil
	}),
}
