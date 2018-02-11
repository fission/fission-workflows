package main

import (
	"fmt"
	"os"

	"io"
	"strings"
	"text/tabwriter"

	"net/url"

	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/admin_api"
	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/workflow_api"
	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/workflow_invocation_api"
	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/models"
	"github.com/fission/fission-workflows/pkg/api/workflow/parse/yaml"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/version"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/urfave/cli"
	"sort"
)

// This is a prototype of the CLI (and will be integrated into the Fission CLI eventually).
func main() {
	app := cli.NewApp()
	app.Version = version.VERSION
	app.Author = "Erwin van Eyk"
	app.Email = "erwin@platform9.com"
	app.EnableBashCompletion = true
	app.Usage = "Fission Workflows CLI"
	app.Description = "CLI for Fission Workflows"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "url, u",
			Value:  "192.168.99.100:31319",
			EnvVar: "FISSION_URL",
			Usage:  "Url to the Fission apiserver",
		},
		cli.BoolFlag{
			Name:   "debug, d",
			EnvVar: "WFCLI_DEBUG",
		},
	}
	app.Commands = []cli.Command{
		{
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
		},
		{
			Name:    "config",
			Usage:   "Print wfcli config",
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
		},
		{
			Name:        "parse",
			Aliases:     []string{"p"},
			Usage:       "parse <path-to-yaml> ",
			Description: "Parse YAML definitions to the executable JSON format (deprecated)",
			Action: func(c *cli.Context) error {

				if c.NArg() == 0 {
					panic("Need a path to a yaml workflow definition")
				}

				for _, path := range c.Args() {

					fnName := strings.TrimSpace(path)

					f, err := os.Open(fnName)
					if err != nil {
						panic(err)
					}

					wfDef, err := yaml.Parse(f)
					if err != nil {
						panic(err)
					}

					wfSpec, err := yaml.Transform(wfDef)
					if err != nil {
						panic(err)
					}

					marshal := jsonpb.Marshaler{
						Indent: "  ",
					}
					jsonWf, err := marshal.MarshalToString(wfSpec)
					if err != nil {
						panic(err)
					}

					//outputFile := strings.Replace(fnName, "yaml", "json", -1)
					//
					//err = ioutil.WriteFile(outputFile, []byte(jsonWf), 0644)
					//if err != nil {
					//	panic(err)
					//}
					//
					//println(outputFile)

					fmt.Println(jsonWf)
				}
				return nil
			},
		},
		{
			Name:    "workflow",
			Aliases: []string{"wf", "workflows"},
			Usage:   "Workflow-related commands",
			Subcommands: []cli.Command{
				{
					Name:  "get",
					Usage: "get <workflow-id> <task-id>",
					Action: func(c *cli.Context) error {
						u := parseUrl(c.GlobalString("url"))
						client := createTransportClient(u)
						wfApi := workflow_api.New(client, strfmt.Default)
						switch c.NArg() {
						case 0:
							// List workflows
							resp, err := wfApi.List0(workflow_api.NewList0Params())
							if err != nil {
								panic(err)
							}
							wfs := resp.Payload.Workflows
							sort.Strings(wfs)
							rows := [][]string{}
							for _, wfId := range wfs {
								resp, err := wfApi.Get0(workflow_api.NewGet0Params().WithID(wfId))
								if err != nil {
									panic(err)
								}
								wf := resp.Payload
								updated := wf.Status.UpdatedAt.String()
								created := wf.Metadata.CreatedAt.String()

								rows = append(rows, []string{wfId, wf.Spec.Name, string(wf.Status.Status),
									created, updated})
							}
							table(os.Stdout, []string{"ID", "NAME", "STATUS", "CREATED", "UPDATED"}, rows)
						case 1:
							// Get Workflow
							wfId := c.Args().Get(0)
							println(wfId)
							resp, err := wfApi.Get0(workflow_api.NewGet0Params().WithID(wfId))
							if err != nil {
								panic(err)
							}
							b, err := yaml.Marshal(resp.Payload)
							if err != nil {
								panic(err)
							}
							fmt.Printf("%v\n", string(b))
						case 2:
							// Get Workflow task
							fallthrough
						default:
							wfId := c.Args().Get(0)
							taskId := c.Args().Get(1)
							resp, err := wfApi.Get0(workflow_api.NewGet0Params().WithID(wfId))
							if err != nil {
								panic(err)
							}
							wf := resp.Payload
							task, ok := wf.Spec.Tasks[taskId]
							if !ok {
								fmt.Println("Task not found.")
								return nil
							}
							b, err := yaml.Marshal(task)
							if err != nil {
								panic(err)
							}
							fmt.Printf("%v\n", string(b))
						}

						return nil
					},
				},
			},
		},
		{
			Name:    "invocation",
			Aliases: []string{"wi", "invocations", "workflow-invocation", "wfi"},
			Usage:   "Workflow Invocation-related commands",
			Subcommands: []cli.Command{
				{
					Name:  "get",
					Usage: "get <workflow-invocation-id> <task-invocation-id>",
					Action: func(c *cli.Context) error {
						u := parseUrl(c.GlobalString("url"))
						client := createTransportClient(u)
						wfiApi := workflow_invocation_api.New(client, strfmt.Default)
						switch c.NArg() {
						case 0:
							// List workflows invocations
							resp, err := wfiApi.List(workflow_invocation_api.NewListParams())
							if err != nil {
								panic(err)
							}
							wis := resp.Payload
							sort.Strings(wis.Invocations)
							var rows [][]string
							for _, wfiId := range wis.Invocations {
								resp, err := wfiApi.Get(workflow_invocation_api.NewGetParams().WithID(wfiId))
								if err != nil {
									panic(err)
								}
								wi := resp.Payload
								updated := wi.Status.UpdatedAt.String()
								created := wi.Metadata.CreatedAt.String()

								rows = append(rows, []string{wfiId, wi.Spec.WorkflowID, string(wi.Status.Status),
									created, updated})
							}
							table(os.Stdout, []string{"ID", "WORKFLOW", "STATUS", "CREATED", "UPDATED"}, rows)
						case 1:
							// Get Workflow invocation
							wfiId := c.Args().Get(0)
							resp, err := wfiApi.Get(workflow_invocation_api.NewGetParams().WithID(wfiId))
							if err != nil {
								panic(err)
							}
							wfi := resp.Payload
							b, err := yaml.Marshal(wfi)
							if err != nil {
								panic(err)
							}
							fmt.Printf("%v\n", string(b))
						case 2:
							fallthrough
						default:
							wfiId := c.Args().Get(0)
							taskId := c.Args().Get(1)
							resp, err := wfiApi.Get(workflow_invocation_api.NewGetParams().WithID(wfiId))
							if err != nil {
								panic(err)
							}
							wfi := resp.Payload
							ti, ok := wfi.Status.Tasks[taskId]
							if !ok {
								fmt.Println("Task invocation not found.")
								return nil
							}
							b, err := yaml.Marshal(ti)
							if err != nil {
								panic(err)
							}
							fmt.Printf("%v\n", string(b))
						}

						return nil
					},
				},
				{
					Name:  "status",
					Usage: "status <workflow-invocation-id> ",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							fmt.Println("Need workflow invocation id")
							return nil
						}
						wfiId := c.Args().Get(0)
						u := parseUrl(c.GlobalString("url"))
						client := createTransportClient(u)
						wfApi := workflow_api.New(client, strfmt.Default)
						wfiApi := workflow_invocation_api.New(client, strfmt.Default)

						wfiResp, err := wfiApi.Get(workflow_invocation_api.NewGetParams().WithID(wfiId))
						if err != nil {
							panic(err)
						}
						wfi := wfiResp.Payload

						wfResp, err := wfApi.Get0(workflow_api.NewGet0Params().WithID(wfi.Spec.WorkflowID))
						if err != nil {
							panic(err)
						}
						wf := wfResp.Payload

						wfiUpdated := wfi.Status.UpdatedAt.String()
						wfiCreated := wfi.Metadata.CreatedAt.String()
						table(os.Stdout, nil, [][]string{
							{"ID", wfi.Metadata.ID},
							{"WORKFLOW_ID", wfi.Spec.WorkflowID},
							{"CREATED", wfiCreated},
							{"UPDATED", wfiUpdated},
							{"STATUS", string(wfi.Status.Status)},
						})
						fmt.Println()

						rows := [][]string{}
						rows = collectStatus(wf.Spec.Tasks, wfi.Status.Tasks, rows)
						rows = collectStatus(wfi.Status.DynamicTasks, wfi.Status.Tasks, rows)

						table(os.Stdout, []string{"TASK", "STATUS", "STARTED", "UPDATED"}, rows)
						return nil
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
func collectStatus(tasks map[string]models.Task, taskStatus map[string]models.TaskInvocation, rows [][]string) [][]string {
	ids := []string{}
	for id := range tasks {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		status := types.TaskInvocationStatus_SCHEDULED.String()
		updated := ""
		started := ""

		taskStatus, ok := taskStatus[id]
		if ok {
			status = string(taskStatus.Status.Status)
			started = taskStatus.Metadata.CreatedAt.String()
			updated = taskStatus.Metadata.CreatedAt.String()
		}

		rows = append(rows, []string{id, status, started, updated})
	}
	return rows
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

func createTransportClient(baseUrl *url.URL) *httptransport.Runtime {
	return httptransport.New(baseUrl.Host, "/proxy/workflows-apiserver/", []string{baseUrl.Scheme})
}

func fatal(msg interface{}) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func parseUrl(rawUrl string) *url.URL {
	u, err := url.Parse(rawUrl)
	if err != nil {
		fatal(fmt.Sprintf("Invalid url '%s': %v", rawUrl, err))
	}
	return u
}
