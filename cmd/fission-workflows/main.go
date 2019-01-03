package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fission/fission-workflows/pkg/version"
	"github.com/fission/fission/fission/plugin"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// This is a prototype of the CLI (and will be integrated into the Fission CLI eventually).
func main() {

	app := cli.NewApp()
	app.Author = "Erwin van Eyk"
	app.Email = "erwin@platform9.com"
	app.Version = version.Version
	app.EnableBashCompletion = true
	app.Usage = "Inspect, manage, and debug workflow executions"
	app.Description = app.Usage
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "url, u",
			EnvVar: "FISSION_URL",
			Usage:  "URL to the Fission apiserver",
		},
		cli.StringFlag{
			Name:   "path-prefix",
			EnvVar: "FISSION_PATH_PREFIX",
			Value:  "/proxy/workflows-apiserver",
			Usage:  "The path to prepend each of the commands",
		},
		cli.IntFlag{
			Name:   "verbosity",
			Value:  1,
			Usage:  "CLI verbosity (0 is quiet, 1 is the default, 2 is verbose.)",
			EnvVar: "FISSION_VERBOSITY",
		},
		cli.BoolFlag{
			Hidden: true,
			Name:   "plugin",
			Usage:  "Show Fission plugin info",
		},
	}
	app.Commands = []cli.Command{
		cmdConfig,
		cmdStatus,
		cmdParse,
		cmdWorkflow,
		cmdInvocation,
		cmdValidate,
		cmdVersion,
	}
	app.Action = func(ctx *cli.Context) error {
		if ctx.GlobalBool("plugin") {
			bs, err := json.Marshal(plugin.Metadata{
				Name:    "workflows",
				Version: version.Version,
				Aliases: []string{"wf"},
				Usage:   ctx.App.Usage,
			})
			if err != nil {
				panic(err)
			}
			fmt.Println(string(bs))
			return nil
		}
		cli.ShowAppHelp(ctx)
		return nil
	}
	app.Run(os.Args)
}

func table(writer io.Writer, headings []string, rows [][]string) {
	w := tabwriter.NewWriter(writer, 0, 0, 5, ' ', 0)
	if headings != nil {
		fmt.Fprintln(w, strings.Join(headings, "\t")+"\t")
	}
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t")+"\t")
	}
	err := w.Flush()
	if err != nil {
		panic(err)
	}
}

func fail(msg ...interface{}) {
	for _, line := range msg {
		fmt.Fprintln(os.Stderr, line)
	}
	os.Exit(1)
}

type Context struct {
	*cli.Context
}

func (c Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c Context) Done() <-chan struct{} {
	return nil
}

func (c Context) Err() error {
	return nil
}

func (c Context) Value(key interface{}) interface{} {
	if s, ok := key.(string); ok {
		return c.Generic(s)
	}
	return nil
}

func commandContext(fn func(c Context) error) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		switch c.GlobalInt("verbosity") {
		case 0:
			logrus.SetLevel(logrus.ErrorLevel)
		case 1:
			logrus.SetLevel(logrus.InfoLevel)
		default:
			fallthrough
		case 2:
			logrus.SetLevel(logrus.DebugLevel)
		}
		return fn(Context{c})
	}
}
