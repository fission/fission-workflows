package main

import (
	"fmt"

	"github.com/urfave/cli"
)

var cmdConfig = cli.Command{
	Name:  "config",
	Usage: "Print wfcli config",
	Action: func(c *cli.Context) error {
		fmt.Println("cli:")
		for _, flag := range c.GlobalFlagNames() {
			fmt.Printf("  %s: %v\n", flag, c.GlobalGeneric(flag))
		}
		for _, flag := range c.FlagNames() {
			fmt.Printf("  %s: %v\n", flag, c.Generic(flag))
		}
		return nil
	},
}
