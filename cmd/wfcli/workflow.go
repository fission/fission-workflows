package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/workflow_api"
	"github.com/fission/fission-workflows/pkg/parse/yaml"
	"github.com/go-openapi/strfmt"
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
				u := parseUrl(c.GlobalString("url"))
				client := createTransportClient(u)
				wfApi := workflow_api.New(client, strfmt.Default)
				switch c.NArg() {
				case 0:
					// List workflows
					resp, err := wfApi.WfList(workflow_api.NewWfListParams())
					if err != nil {
						panic(err)
					}
					wfs := resp.Payload.Workflows
					sort.Strings(wfs)
					var rows [][]string
					for _, wfId := range wfs {
						resp, err := wfApi.WfGet(workflow_api.NewWfGetParams().WithID(wfId))
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
					resp, err := wfApi.WfGet(workflow_api.NewWfGetParams().WithID(wfId))
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
					resp, err := wfApi.WfGet(workflow_api.NewWfGetParams().WithID(wfId))
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
}
