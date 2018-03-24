package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/fission/fission-workflows/pkg/apiserver/httpclient"
	"github.com/fission/fission-workflows/pkg/parse/yaml"
	"github.com/fission/fission-workflows/pkg/types"
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
					Name:  "history",
					Usage: "Amount history (non-active invocations) to show.",
					Value: time.Duration(1) * time.Hour,
				},
			},
			Action: func(c *cli.Context) error {
				//u := parseUrl(c.GlobalString("url"))
				//client := createTransportClient(u)
				//wfiApi := workflow_invocation_api.New(client, strfmt.Default)
				ctx := context.TODO()
				url := parseUrl(c.GlobalString("url"))
				wfiApi := httpclient.NewInvocationApi(url.String(), http.Client{})
				switch c.NArg() {
				case 0:
					since := c.Duration("history")
					invocationsList(os.Stdout, wfiApi, time.Now().Add(-since))
				case 1:
					// Get Workflow invocation
					wfiId := c.Args().Get(0)
					wfi, err := wfiApi.Get(ctx, wfiId)
					if err != nil {
						panic(err)
					}
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
					wfi, err := wfiApi.Get(ctx, wfiId)
					if err != nil {
						panic(err)
					}
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
				//u := parseUrl(c.GlobalString("url"))
				//client := createTransportClient(u)
				//wfiApi := workflow_invocation_api.New(client, strfmt.Default)
				ctx := context.TODO()
				url := parseUrl(c.GlobalString("url"))
				wfiApi := httpclient.NewInvocationApi(url.String(), http.Client{})
				err := wfiApi.Cancel(ctx, wfiId)
				if err != nil {
					panic(err)
				}
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
				//u := parseUrl(c.GlobalString("url"))
				//client := createTransportClient(u)
				//wfiApi := workflow_invocation_api.New(client, strfmt.Default)
				ctx := context.TODO()
				url := parseUrl(c.GlobalString("url"))
				wfiApi := httpclient.NewInvocationApi(url.String(), http.Client{})
				spec := &types.WorkflowInvocationSpec{
					WorkflowId: wfId,
					Inputs:     map[string]*types.TypedValue{},
				}
				if c.Bool("sync") {
					resp, err := wfiApi.InvokeSync(ctx, spec)
					if err != nil {
						panic(err)
					}
					bs, err := yaml.Marshal(resp)
					if err != nil {
						panic(err)
					}
					fmt.Println(string(bs))
				} else {
					resp, err := wfiApi.Invoke(ctx, spec)
					if err != nil {
						panic(err)
					}
					fmt.Println(resp.Id)
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
				//u := parseUrl(c.GlobalString("url"))
				//client := createTransportClient(u)
				//wfApi := workflow_api.New(client, strfmt.Default)
				//wfiApi := workflow_invocation_api.New(client, strfmt.Default)
				ctx := context.TODO()
				url := parseUrl(c.GlobalString("url"))
				wfiApi := httpclient.NewInvocationApi(url.String(), http.Client{})
				wfApi := httpclient.NewWorkflowApi(url.String(), http.Client{})

				wfi, err := wfiApi.Get(ctx, wfiId)
				if err != nil {
					panic(err)
				}

				wf, err := wfApi.Get(ctx, wfi.Spec.WorkflowId)
				if err != nil {
					panic(err)
				}

				wfiUpdated := wfi.Status.UpdatedAt.String()
				wfiCreated := wfi.Metadata.CreatedAt.String()
				table(os.Stdout, nil, [][]string{
					{"ID", wfi.Metadata.Id},
					{"WORKFLOW_ID", wfi.Spec.WorkflowId},
					{"CREATED", wfiCreated},
					{"UPDATED", wfiUpdated},
					{"STATUS", string(wfi.Status.Status)},
				})
				fmt.Println()

				var rows [][]string
				rows = collectStatus(wf.Spec.Tasks, wfi.Status.Tasks, rows)
				dynamicTaskSpecs := map[string]*types.TaskSpec{}
				for k, v := range wfi.Status.DynamicTasks {
					dynamicTaskSpecs[k] = v.Spec
				}
				rows = collectStatus(dynamicTaskSpecs, wfi.Status.Tasks, rows)

				table(os.Stdout, []string{"TASK", "STATUS", "STARTED", "UPDATED"}, rows)
				return nil
			},
		},
	},
}

func invocationsList(out io.Writer, wfiApi *httpclient.InvocationApi, since time.Time) {
	// List workflows invocations
	ctx := context.TODO()
	wis, err := wfiApi.List(ctx)
	if err != nil {
		panic(err)
	}
	sort.Strings(wis.Invocations)
	var rows [][]string
	for _, wfiId := range wis.Invocations {
		wi, err := wfiApi.Get(ctx, wfiId)
		if err != nil {
			panic(err)
		}
		updated := wi.Status.UpdatedAt.String()
		created := wi.Metadata.CreatedAt.String()

		// TODO add filter params to endpoint instead
		// TODO filter old invocations and system invocations

		rows = append(rows, []string{wfiId, wi.Spec.WorkflowId, string(wi.Status.Status),
			created, updated})
	}

	table(out, []string{"ID", "WORKFLOW", "STATUS", "CREATED", "UPDATED"}, rows)

}

func collectStatus(tasks map[string]*types.TaskSpec, taskStatus map[string]*types.TaskInvocation,
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
