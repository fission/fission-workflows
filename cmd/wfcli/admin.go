package main

import (
	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/admin_api"
	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"
)

var cmdAdmin = cli.Command{
	Name:  "admin",
	Usage: "Administrative commands",
	Subcommands: []cli.Command{
		cmdStatus,
		{
			Name: "halt",
			Action: func(c *cli.Context) error {
				u := parseUrl(c.GlobalString("url"))
				client := createTransportClient(u)
				adminApi := admin_api.New(client, strfmt.Default)
				adminApi.Status(nil)
				return nil
			},
		},
		{
			Name: "resume",
			Action: func(c *cli.Context) error {
				u := parseUrl(c.GlobalString("url"))
				client := createTransportClient(u)
				adminApi := admin_api.New(client, strfmt.Default)
				adminApi.Status(nil)
				return nil
			},
		},
	},
}
