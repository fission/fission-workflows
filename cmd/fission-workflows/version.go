package main

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/fission/fission-workflows/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var versionPrinter = commandContext(func(ctx Context) error {
	c := ctx.Bool("client")
	s := ctx.Bool("server")
	if !c && !s {
		c = true
		s = true
	}

	// Print client version
	if c {
		fmt.Printf("client: %s\n", version.VersionInfo().JSON())
	}

	// Print server version
	if s {
		client := getClient(ctx)
		resp, err := client.Admin.Version(ctx)
		if err != nil {
			fmt.Printf("server: failed to get version: %v\n", err)
		} else {
			fmt.Printf("server: %s\n", resp.JSON())
		}
	}
	return nil
})

var cmdVersion = cli.Command{
	Name:    "version",
	Usage:   "Print version of both client and server.",
	Aliases: []string{"v"},
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "c, client",
			Usage: "Only show the version of the client.",
		},
		cli.BoolFlag{
			Name:  "s, server",
			Usage: "Only show the version of the client.",
		},
	},
	Action: versionPrinter,
}

func fetchServerVersion(ctx Context) (semver.Version, error) {
	client := getClient(ctx)
	resp, err := client.Admin.Version(ctx)
	if err != nil {
		return semver.Version{}, err
	}

	return semver.Parse(resp.GetVersion())
}

func ensureServerVersionAtLeast(ctx Context, minVersion semver.Version, orFail bool) bool {
	errorPrinter := logrus.Warnf
	if orFail {
		errorPrinter = logrus.Fatalf
	}

	v, err := fetchServerVersion(ctx)
	if err != nil {
		errorPrinter("Failed fetch server version info to check if the command is supported: %v", err)
		return false
	}
	if v.LT(minVersion) {
		errorPrinter("Server is outdated; this functionality might not be supported "+
			"(server version expected: %s, but was: %s)", minVersion, v.String())
		return false
	}
	return true
}
