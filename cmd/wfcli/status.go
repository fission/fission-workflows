package main

import (
	"fmt"

	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/admin_api"
	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"
)

var cmdStatus = cli.Command{
	Name:    "status",
	Aliases: []string{"s"},
	Usage:   "Check cluster status",
	Action: func(c *cli.Context) error {
		u := parseUrl(c.GlobalString("url"))
		client := createTransportClient(u)
		adminApi := admin_api.New(client, strfmt.Default)

		resp, err := adminApi.Status(admin_api.NewStatusParams())
		if err != nil {
			panic(err)
		}
		fmt.Printf(resp.Payload.Status)

		return nil
	},
}
