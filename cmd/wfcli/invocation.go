package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/workflow_api"
	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/workflow_invocation_api"
	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/models"
	"github.com/fission/fission-workflows/pkg/api/workflow/parse/yaml"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli"
)

var cmdInvocation = cli.Command{
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

				var rows [][]string
				rows = collectStatus(wf.Spec.Tasks, wfi.Status.Tasks, rows)
				rows = collectStatus(wfi.Status.DynamicTasks, wfi.Status.Tasks, rows)

				table(os.Stdout, []string{"TASK", "STATUS", "STARTED", "UPDATED"}, rows)
				return nil
			},
		},
	},
}

func collectStatus(tasks map[string]models.Task, taskStatus map[string]models.TaskInvocation, rows [][]string) [][]string {
	var ids []string
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
