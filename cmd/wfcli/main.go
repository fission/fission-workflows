package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fission/fission-workflows/pkg/version"
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
	app.Usage = "Fission Workflows CLI"
	app.Description = "CLI for Fission Workflows"
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
		cli.BoolFlag{
			Name:   "debug, d",
			EnvVar: "WFCLI_DEBUG",
		},
	}
	app.Commands = []cli.Command{
		cmdConfig,
		cmdStatus,
		cmdParse,
		cmdWorkflow,
		cmdInvocation,
		cmdValidate,
		cmdAdmin,
		cmdVersion,
		cmdParseJS, // TEMP
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
		if c.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			logrus.SetLevel(logrus.InfoLevel)
		}
		return fn(Context{c})
	}
}
