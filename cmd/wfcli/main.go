package main

import (
	"fmt"
	"os"

	"context"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/version"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

// This is a prototype of the CLI (standalone or integrated into Fission).
func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "url, u",
			Value:  "192.168.99.100:31319",
			EnvVar: "FISSION_WORKFLOW_APISERVER_URL",
			Usage:  "Url to workflow engine gRPC API",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "Version of this workflow CLI",
			Action: func(c *cli.Context) error {
				fmt.Printf("%s", version.VERSION)
				return nil
			},
		},
		{
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "Check cluster",
			Action: func(c *cli.Context) error {
				ctx := context.Background()
				url := c.GlobalString("url")
				fmt.Printf("Using url: '%s'\n", url)
				conn, err := grpc.Dial(url, grpc.WithInsecure())
				if err != nil {
					panic(err)
				}
				adminApi := apiserver.NewAdminAPIClient(conn)
				resp, err := adminApi.Status(ctx, &empty.Empty{})
				if err != nil {
					panic(err)
				}
				fmt.Printf(resp.Status)

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
						ctx := context.Background()
						conn, err := grpc.Dial(c.GlobalString("url"), grpc.WithInsecure())
						if err != nil {
							panic(err)
						}
						wfApi := apiserver.NewWorkflowAPIClient(conn)
						switch c.NArg() {
						case 0:
							// List workflows
							resp, err := wfApi.List(ctx, &empty.Empty{})
							if err != nil {
								panic(err)
							}
							rows := [][]string{}
							for _, wfId := range resp.Workflows {
								wf, err := wfApi.Get(ctx, &apiserver.WorkflowIdentifier{wfId})
								if err != nil {
									panic(err)
								}
								updated, _ := ptypes.Timestamp(wf.Status.UpdatedAt)
								created, _ := ptypes.Timestamp(wf.Metadata.CreatedAt)

								rows = append(rows, []string{wfId, wf.Spec.Name, wf.Status.Status.String(),
									created.String(), updated.String()})
							}
							table(os.Stdout, []string{"ID", "NAME", "STATUS", "CREATED", "UPDATED"}, rows)
						case 1:
							// Get Workflow
							wfId := c.Args().Get(0)
							println(wfId)
							wf, err := wfApi.Get(ctx, &apiserver.WorkflowIdentifier{wfId})
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
							wf, err := wfApi.Get(ctx, &apiserver.WorkflowIdentifier{wfId})
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
						ctx := context.Background()
						conn, err := grpc.Dial(c.GlobalString("url"), grpc.WithInsecure())
						if err != nil {
							panic(err)
						}
						wfiApi := apiserver.NewWorkflowInvocationAPIClient(conn)
						switch c.NArg() {
						case 0:
							// List workflows invocations
							wis, err := wfiApi.List(ctx, &empty.Empty{})
							if err != nil {
								panic(err)
							}
							rows := [][]string{}
							for _, wiId := range wis.Invocations {
								wi, err := wfiApi.Get(ctx, &apiserver.WorkflowInvocationIdentifier{wiId})
								if err != nil {
									panic(err)
								}
								updated, _ := ptypes.Timestamp(wi.Status.UpdatedAt)
								created, _ := ptypes.Timestamp(wi.Metadata.CreatedAt)

								rows = append(rows, []string{wiId, wi.Spec.WorkflowId, wi.Status.Status.String(),
									created.String(), updated.String()})
							}
							table(os.Stdout, []string{"ID", "WORKFLOW", "STATUS", "CREATED", "UPDATED"}, rows)
						case 1:
							// Get Workflow invocation
							wfiId := c.Args().Get(0)
							wfi, err := wfiApi.Get(ctx, &apiserver.WorkflowInvocationIdentifier{wfiId})
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
							wfi, err := wfiApi.Get(ctx, &apiserver.WorkflowInvocationIdentifier{wfiId})
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
					Name:  "status",
					Usage: "status <workflow-invocation-id> ",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							fmt.Println("Need workflow invocation id")
							return nil
						}
						wfiId := c.Args().Get(0)
						ctx := context.Background()
						conn, err := grpc.Dial(c.GlobalString("url"), grpc.WithInsecure())
						if err != nil {
							panic(err)
						}
						wfiApi := apiserver.NewWorkflowInvocationAPIClient(conn)
						wfApi := apiserver.NewWorkflowAPIClient(conn)

						wfi, err := wfiApi.Get(ctx, &apiserver.WorkflowInvocationIdentifier{wfiId})
						if err != nil {
							panic(err)
						}

						wf, err := wfApi.Get(ctx, &apiserver.WorkflowIdentifier{wfi.Spec.WorkflowId})
						if err != nil {
							panic(err)
						}

						wfiUpdated, _ := ptypes.Timestamp(wfi.Status.UpdatedAt)
						wfiCreated, _ := ptypes.Timestamp(wfi.Metadata.CreatedAt)
						table(os.Stdout, nil, [][]string{
							{"ID", wfi.Metadata.Id},
							{"WORKFLOW_ID", wfi.Spec.WorkflowId},
							{"CREATED", wfiCreated.String()},
							{"UPDATED", wfiUpdated.String()},
							{"STATUS", wfi.Status.Status.String()},
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
func collectStatus(tasks map[string]*types.Task, taskStatus map[string]*types.TaskInvocation, rows [][]string) [][]string {
	for id := range tasks {
		status := types.TaskInvocationStatus_UNKNOWN.String()
		updated := ""
		started := ""

		taskStatus, ok := taskStatus[id]
		if ok {
			status = taskStatus.Status.Status.String()
			tca, _ := ptypes.Timestamp(taskStatus.Metadata.CreatedAt)
			started = tca.String()
			tua, _ := ptypes.Timestamp(taskStatus.Metadata.CreatedAt)
			updated = tua.String()
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
