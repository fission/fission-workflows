package main

import (
	"context"
	"net/http"

	"github.com/fission/fission-workflows/pkg/apiserver/httpclient"
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
				ctx := context.TODO()
				url := parseUrl(c.GlobalString("url"))
				admin := httpclient.NewAdminApi(url.String(), http.Client{})
				err := admin.Halt(ctx)
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
				ctx := context.TODO()
				url := parseUrl(c.GlobalString("url"))
				admin := httpclient.NewAdminApi(url.String(), http.Client{})
				err := admin.Resume(ctx)
				if err != nil {
					panic(err)
				}
				return nil
			},
		},
	},
}
