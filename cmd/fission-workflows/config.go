package main

import (
	"fmt"

	"github.com/urfave/cli"
)

var cmdConfig = cli.Command{
	Name:   "config",
	Hidden: true,
	Usage:  "Print fission-workflows config",
	Action: commandContext(func(ctx Context) error {
		fmt.Println("cli:")
		for _, flag := range ctx.GlobalFlagNames() {
			fmt.Printf("  %s: %v\n", flag, ctx.GlobalGeneric(flag))
		}
		for _, flag := range ctx.FlagNames() {
			fmt.Printf("  %s: %v\n", flag, ctx.Generic(flag))
		}
		return nil
	}),
}
