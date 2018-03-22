package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fission/fission-workflows/pkg/apiserver/httpclient"
	"github.com/fission/fission-workflows/pkg/version"
	"github.com/urfave/cli"
)

var versionPrinter = func(c *cli.Context) error {
	// Print client version
	fmt.Printf("client: %s\n", version.VERSION)

	// Print server version
	ctx := context.TODO()
	url := parseUrl(c.GlobalString("url"))
	admin := httpclient.NewAdminApi(url.String(), http.Client{})
	resp, err := admin.Version(ctx)
	if err != nil {
		fmt.Printf("server: failed to get version (%v)\n", err)
	} else {
		fmt.Printf("server: %s\n", resp.Version)
	}
	return nil
}

var cmdVersion = cli.Command{
	Name:    "version",
	Usage:   "Print version of both client and server",
	Aliases: []string{"v"},
	Flags:   []cli.Flag{},
	Action:  versionPrinter,
}
