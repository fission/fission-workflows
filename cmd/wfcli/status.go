package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fission/fission-workflows/pkg/apiserver/httpclient"
	"github.com/urfave/cli"
)

var cmdStatus = cli.Command{
	Name:    "status",
	Aliases: []string{"s"},
	Usage:   "Check cluster status",
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		url := parseUrl(c.GlobalString("url"))
		admin := httpclient.NewAdminApi(url.String(), http.Client{})

		resp, err := admin.Status(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf(resp.Status)

		return nil
	},
}
