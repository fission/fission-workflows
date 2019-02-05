package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/blang/semver"
	"github.com/fission/fission-workflows/pkg/apiserver/httpclient"
	"github.com/fission/fission-workflows/pkg/parse/yaml"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var cmdInvocation = cli.Command{
	Name:    "invocation",
	Aliases: []string{"wfi", "invocations"},
	Usage:   "Workflow invocation-related commands",
	Subcommands: []cli.Command{
		{
			Name:  "get",
			Usage: "get <invocation-id> <task-id>",
			Flags: []cli.Flag{
				cli.DurationFlag{
					Name:  "history",
					Usage: "Amount history (non-active invocations) to show.",
					Value: time.Duration(1) * time.Hour,
				},
			},
			Action: commandContext(func(ctx Context) error {
				client := getClient(ctx)
				switch ctx.NArg() {
				case 0:
					since := ctx.Duration("history")
					invocationsList(os.Stdout, client.Invocation, time.Now().Add(-since))
				case 1:
					// Get Workflow Invocation
					wfiID := ctx.Args().Get(0)
					wfi, err := client.Invocation.Get(ctx, wfiID)
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
					wfiID := ctx.Args().Get(0)
					taskID := ctx.Args().Get(1)
					wfi, err := client.Invocation.Get(ctx, wfiID)
					if err != nil {
						panic(err)
					}
					ti, ok := wfi.Status.Tasks[taskID]
					if !ok {
						fmt.Println("Task Invocation not found.")
						return nil
					}
					b, err := yaml.Marshal(ti)
					if err != nil {
						panic(err)
					}
					fmt.Printf("%v\n", string(b))
				}

				return nil
			}),
		},
		{
			Name:  "cancel",
			Usage: "cancel <invocation-id>",
			Action: commandContext(func(ctx Context) error {
				client := getClient(ctx)
				wfiID := ctx.Args().Get(0)
				err := client.Invocation.Cancel(ctx, wfiID)
				if err != nil {
					panic(err)
				}
				return nil
			}),
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
			Action: commandContext(func(ctx Context) error {
				client := getClient(ctx)
				wfID := ctx.Args().Get(0)
				spec := &types.WorkflowInvocationSpec{
					WorkflowId: wfID,
					Inputs:     map[string]*typedvalues.TypedValue{},
				}
				if ctx.Bool("sync") {
					resp, err := client.Invocation.InvokeSync(ctx, spec)
					if err != nil {
						panic(err)
					}
					bs, err := yaml.Marshal(resp)
					if err != nil {
						panic(err)
					}
					fmt.Println(string(bs))
				} else {
					resp, err := client.Invocation.Invoke(ctx, spec)
					if err != nil {
						panic(err)
					}
					fmt.Println(resp.Id)
				}
				return nil
			}),
		},
		{
			Name:  "events",
			Usage: "events <invocation-id>",
			Action: commandContext(func(ctx Context) error {
				ensureServerVersionAtLeast(ctx, semver.MustParse("0.7.0"), true)
				if !ctx.Args().Present() {
					logrus.Fatal("Usage: fission-workflows invocation events <invocation-id>")
				}
				client := getClient(ctx)
				wfiID := ctx.Args().First()

				events, err := client.Invocation.Events(ctx, wfiID)
				if err != nil {
					logrus.Fatalf("Failed to retrieve events for %s: %v", wfiID, err)
				}

				for _, event := range events.GetEvents() {
					err := (&jsonpb.Marshaler{
						Indent: "	",
					}).Marshal(os.Stdout, event)
					if err != nil {
						panic(err)
					}
				}

				return nil
			}),
		},
		{
			Name:  "status",
			Usage: "status <Workflow-Invocation-id> ",
			Action: commandContext(func(ctx Context) error {
				if !ctx.Args().Present() {
					logrus.Fatal("Usage: fission-workflows invocation status <invocation-id>")
				}
				client := getClient(ctx)
				wfiID := ctx.Args().First()

				wfi, err := client.Invocation.Get(ctx, wfiID)
				if err != nil {
					logrus.Fatalf("Failed to retrieve status for %s: %v", wfiID, err)
				}

				wf, err := client.Workflow.Get(ctx, wfi.Spec.WorkflowId)
				if err != nil {
					panic(err)
				}

				wfiUpdated := ptypes.TimestampString(wfi.Status.UpdatedAt)
				wfiCreated := ptypes.TimestampString(wfi.Metadata.CreatedAt)
				table(os.Stdout, nil, [][]string{
					{"id", wfi.Metadata.Id},
					{"WORKFLOW_ID", wfi.Spec.WorkflowId},
					{"CREATED", wfiCreated},
					{"UPDATED", wfiUpdated},
					{"STATUS", wfi.Status.Status.String()},
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
			}),
		},
	},
}

func invocationsList(out io.Writer, wfiAPI *httpclient.InvocationAPI, since time.Time) {
	// List workflows invocations
	ctx := context.TODO()
	wis, err := wfiAPI.List(ctx)
	if err != nil {
		panic(err)
	}

	var invocations []*types.WorkflowInvocation
	for _, wfiID := range wis.Invocations {
		wi, err := wfiAPI.Get(ctx, wfiID)
		if err != nil {
			panic(err)
		}
		if len(wi.Spec.ParentId) != 0 {
			continue
		}

		invocations = append(invocations, wi)
	}

	sort.Slice(invocations, func(i, j int) bool {
		tsi, err := ptypes.Timestamp(invocations[i].Status.UpdatedAt)
		if err != nil {
			panic(err)
		}
		tsj, err := ptypes.Timestamp(invocations[j].Status.UpdatedAt)
		if err != nil {
			panic(err)
		}
		return tsi.Before(tsj)
	})

	var rows [][]string
	for _, wi := range invocations {
		updated := ptypes.TimestampString(wi.Status.UpdatedAt)
		created := ptypes.TimestampString(wi.Metadata.CreatedAt)

		// TODO add filter params to endpoint instead
		// TODO filter old invocations and system invocations

		rows = append(rows, []string{wi.ID(), wi.Spec.WorkflowId, wi.Status.Status.String(),
			created, updated})
	}

	table(out, []string{"id", "WORKFLOW", "STATUS", "CREATED", "UPDATED"}, rows)

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
			status = taskStatus.Status.Status.String()
			updated = ptypes.TimestampString(taskStatus.Status.UpdatedAt)
			started = ptypes.TimestampString(taskStatus.Metadata.CreatedAt)
		}

		rows = append(rows, []string{id, status, started, updated})
	}
	return rows
}
