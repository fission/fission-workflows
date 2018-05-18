package main

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/version"
	"github.com/urfave/cli"
)

var versionPrinter = commandContext(func(ctx Context) error {
	c := ctx.Bool("client")
	s := ctx.Bool("server")
	if !c && !s {
		c = true
		s = true
	}

	// Print client version
	if c {
		fmt.Printf("client: %s\n", version.VersionInfo().Json())
	}

	// Print server version
	if s {
		client := getClient(ctx)
		resp, err := client.Admin.Version(ctx)
		if err != nil {
			fmt.Printf("server: failed to get version (%v)\n", err)
		} else {
			fmt.Printf("server: %s\n", resp)
		}
	}
	return nil
})

var cmdVersion = cli.Command{
	Name:    "version",
	Usage:   "Print version of both client and server.",
	Aliases: []string{"v"},
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "c, client",
			Usage: "Only show the version of the client.",
		},
		cli.BoolFlag{
			Name:  "s, server",
			Usage: "Only show the version of the client.",
		},
	},
	Action: versionPrinter,
}
