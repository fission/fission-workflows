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
		cmdVersion,
		{
			Name:  "halt",
			Usage: "Stop the workflow engine from evaluating anything",
			Action: func(c *cli.Context) error {
				u := parseUrl(c.GlobalString("url"))
				client := createTransportClient(u)
				adminApi := admin_api.New(client, strfmt.Default)
				_, err := adminApi.Halt(admin_api.NewHaltParams())
				if err != nil {
					panic(err)
				}
				return nil
			},
		},
		{
			Name:  "resume",
			Usage: "Resume the workflow engine evaluations",
			Action: func(c *cli.Context) error {
				u := parseUrl(c.GlobalString("url"))
				client := createTransportClient(u)
				adminApi := admin_api.New(client, strfmt.Default)
				_, err := adminApi.Resume(admin_api.NewResumeParams())
				if err != nil {
					panic(err)
				}
				return nil
			},
		},
	},
}
