package main

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/version"
	"github.com/urfave/cli"
)

var versionPrinter = commandContext(func(ctx Context) error {
	// Print client version
	fmt.Printf("client: %s\n", version.VERSION)

	// Print server version
	client := getClient(ctx)
	resp, err := client.Admin.Version(ctx)
	if err != nil {
		fmt.Printf("server: failed to get version (%v)\n", err)
	} else {
		fmt.Printf("server: %s\n", resp.Version)
	}
	return nil
})

var cmdVersion = cli.Command{
	Name:    "version",
	Usage:   "Print version of both client and server",
	Aliases: []string{"v"},
	Flags:   []cli.Flag{},
	Action:  versionPrinter,
}
