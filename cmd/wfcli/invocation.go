package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/workflow_api"
	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/client/workflow_invocation_api"
	"github.com/fission/fission-workflows/cmd/wfcli/swagger-client/models"
	"github.com/fission/fission-workflows/pkg/parse/yaml"
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
			Flags: []cli.Flag{
				cli.DurationFlag{
					Name:  "history, h",
					Usage: "Amount history (non-active invocations) to show.",
					Value: time.Duration(1) * time.Hour,
				},
			},
			Action: func(c *cli.Context) error {
				u := parseUrl(c.GlobalString("url"))
				client := createTransportClient(u)
				wfiApi := workflow_invocation_api.New(client, strfmt.Default)
				switch c.NArg() {
				case 0:
					since := c.Duration("history")
					invocationsList(os.Stdout, wfiApi, time.Now().Add(-since))
				case 1:
					// Get Workflow invocation
					wfiId := c.Args().Get(0)
					resp, err := wfiApi.WfiGet(workflow_invocation_api.NewWfiGetParams().WithID(wfiId))
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
					resp, err := wfiApi.WfiGet(workflow_invocation_api.NewWfiGetParams().WithID(wfiId))
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
			Name:  "cancel",
			Usage: "cancel <workflow-invocation-id>",
			Action: func(c *cli.Context) error {
				wfiId := c.Args().Get(0)
				u := parseUrl(c.GlobalString("url"))
				client := createTransportClient(u)
				wfiApi := workflow_invocation_api.New(client, strfmt.Default)
				resp, err := wfiApi.Cancel(workflow_invocation_api.NewCancelParams().WithID(wfiId))
				if err != nil {
					panic(err)
				}
				fmt.Println(resp)
				return nil
			},
		},
		{
			// TODO support input
			Name:  "invoke",
			Usage: "invoke <workflow-id>",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "input, i",
					Usage: "Not supported!",
				},
				cli.BoolFlag{
					Name:  "sync, s",
					Usage: "Invoke synchronously",
				},
			},
			Action: func(c *cli.Context) error {
				wfId := c.Args().Get(0)
				u := parseUrl(c.GlobalString("url"))
				client := createTransportClient(u)
				wfiApi := workflow_invocation_api.New(client, strfmt.Default)
				body := &models.WorkflowInvocationSpec{
					WorkflowID: wfId,
					Inputs:     map[string]models.TypedValue{},
				}
				if c.Bool("sync") {
					params := workflow_invocation_api.NewInvokeSyncParams().WithBody(body)
					resp, err := wfiApi.InvokeSync(params)
					if err != nil {
						panic(err)
					}
					bs, err := yaml.Marshal(resp.Payload)
					if err != nil {
						panic(err)
					}
					fmt.Println(string(bs))
				} else {
					params := workflow_invocation_api.NewInvokeParams().WithBody(body)
					resp, err := wfiApi.Invoke(params)
					if err != nil {
						panic(err)
					}
					fmt.Println(resp.Payload.ID)
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

				wfiResp, err := wfiApi.WfiGet(workflow_invocation_api.NewWfiGetParams().WithID(wfiId))
				if err != nil {
					panic(err)
				}
				wfi := wfiResp.Payload

				wfResp, err := wfApi.WfGet(workflow_api.NewWfGetParams().WithID(wfi.Spec.WorkflowID))
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
				dynamicTaskSpecs := map[string]models.TaskSpec{}
				for k, v := range wfi.Status.DynamicTasks {
					dynamicTaskSpecs[k] = *v.Spec
				}
				rows = collectStatus(dynamicTaskSpecs, wfi.Status.Tasks, rows)

				table(os.Stdout, []string{"TASK", "STATUS", "STARTED", "UPDATED"}, rows)
				return nil
			},
		},
	},
}

func invocationsList(out io.Writer, wfiApi *workflow_invocation_api.Client, since time.Time) {
	// List workflows invocations
	resp, err := wfiApi.WfiList(workflow_invocation_api.NewWfiListParams())
	if err != nil {
		panic(err)
	}
	wis := resp.Payload
	sort.Strings(wis.Invocations)
	var rows [][]string
	for _, wfiId := range wis.Invocations {
		resp, err := wfiApi.WfiGet(workflow_invocation_api.NewWfiGetParams().WithID(wfiId))
		if err != nil {
			panic(err)
		}
		wi := resp.Payload
		updated := wi.Status.UpdatedAt.String()
		created := wi.Metadata.CreatedAt.String()

		// TODO add filter params to endpoint instead
		if isCompleted(wi.Status.Status) && since.Before(time.Time(wi.Status.UpdatedAt)) {
			continue
		}

		rows = append(rows, []string{wfiId, wi.Spec.WorkflowID, string(wi.Status.Status),
			created, updated})
	}

	table(out, []string{"ID", "WORKFLOW", "STATUS", "CREATED", "UPDATED"}, rows)

}

func collectStatus(tasks map[string]models.TaskSpec, taskStatus map[string]models.TaskInvocation,
	rows [][]string) [][]string {
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

func isCompleted(status models.WorkflowInvocationStatusStatus) bool {
	return status == models.WorkflowInvocationStatusStatusSUCCEEDED ||
		status == models.WorkflowInvocationStatusStatusABORTED ||
		status == models.WorkflowInvocationStatusStatusFAILED
}
