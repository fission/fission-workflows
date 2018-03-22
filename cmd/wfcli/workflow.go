package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/fission/fission-workflows/pkg/apiserver/httpclient"
	"github.com/fission/fission-workflows/pkg/parse/yaml"
	"github.com/urfave/cli"
)

var cmdWorkflow = cli.Command{
	Name:    "workflow",
	Aliases: []string{"wf", "workflows"},
	Usage:   "Workflow-related commands",
	Subcommands: []cli.Command{
		{
			Name:  "get",
			Usage: "get <workflow-id> <task-id>",
			Action: func(c *cli.Context) error {
				//u := parseUrl(c.GlobalString("url"))
				//client := createTransportClient(u)
				//wfApi := workflow_api.New(client, strfmt.Default)
				ctx := context.TODO()
				url := parseUrl(c.GlobalString("url"))
				wfApi := httpclient.NewWorkflowApi(url.String(), http.Client{})

				switch c.NArg() {
				case 0:
					// List workflows
					resp, err := wfApi.List(ctx)
					//resp, err := wf.(workflow_api.NewWfListParams())
					if err != nil {
						panic(err)
					}
					wfs := resp.Workflows
					sort.Strings(wfs)
					var rows [][]string
					for _, wfId := range wfs {
						wf, err := wfApi.Get(ctx, wfId)
						if err != nil {
							panic(err)
						}
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
					wf, err := wfApi.Get(ctx, wfId)
					if err != nil {
						panic(err)
					}
					b, err := yaml.Marshal(wf)
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
					wf, err := wfApi.Get(ctx, wfId)
					if err != nil {
						panic(err)
					}
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
}
