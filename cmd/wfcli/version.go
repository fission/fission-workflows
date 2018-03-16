package main

import (
	"fmt"

	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/admin_api"
	"github.com/fission/fission-workflows/pkg/version"
	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"
)

var versionPrinter = func(c *cli.Context) error {
	fmt.Printf("client: %s\n", version.VERSION)

	u := parseUrl(c.GlobalString("url"))
	client := createTransportClient(u)
	adminApi := admin_api.New(client, strfmt.Default)
	_, err := adminApi.Status(nil)
	if err != nil {
		fmt.Printf("server: failed to get version (%v)\n", err)
	} else {
		fmt.Printf("server: %s\n", "UNKNOWN")
	}
	return nil
}

var cmdVersion = cli.Command{
	Name:    "version",
	Aliases: []string{"v"},
	Flags:   []cli.Flag{},
	Action:  versionPrinter,
}
