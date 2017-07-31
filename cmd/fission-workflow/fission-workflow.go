package main

import (
	"os"
	"sort"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "lang, l",
			Value: "english",
			Usage: "Language for the greeting",
		},
		cli.StringFlag{
			Name: "config, c",
			Usage: "Load configuration from `FILE`",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "workflow",
			Aliases: []string{"wf"},
			Usage:   "Manage Workflows.",
			Action:  func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:    "invocation",
			Aliases: []string{"iv"},
			Usage:   "Manage workflow invocations",
			Action:  func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:    "invocation",
			Aliases: []string{"iv"},
			Usage:   "add a task to the list",
			Action:  func(c *cli.Context) error {
				return nil
			},
		},
	}

	app.Run(os.Args)
}
