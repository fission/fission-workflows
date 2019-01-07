package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/fission/fission-workflows/pkg/parse"
	"github.com/fission/fission-workflows/pkg/parse/yaml"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var cmdWorkflow = cli.Command{
	Name:    "workflow",
	Aliases: []string{"wf", "workflows"},
	Usage:   "Workflow-related commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "Define a workflow within the workflow engine.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "src",
					Usage: "Path to the YAML or Protobuf workflow definition file",
				},
				cli.StringFlag{
					Name:  "name",
					Usage: "Name of the workflow",
				},
			},
			Action: commandContext(func(ctx Context) error {
				client := getClient(ctx)

				// Fetch and parse the workflow
				srcPath := ctx.String("src")
				if len(srcPath) == 0 {
					logrus.Fatalf("Requires workflow definition file. Use `--src <file>`.")
				}
				fd, err := os.Open(srcPath)
				if err != nil {
					logrus.Fatalf("Failed to open workflow definition file: %v", err)
				}
				spec, err := parse.Parse(fd)
				if err != nil {
					logrus.Fatal(err)
				}
				spec.Name = ctx.String("name")

				// Create workflow
				md, err := client.Workflow.Create(ctx, spec)
				if err != nil {
					logrus.Fatalf("Failed to create workflow: %v", err)
				}
				fmt.Println(md.Id)
				return nil
			}),
		},
		{
			Name:  "delete",
			Usage: "Delete workflow within the workflow engine.",
			Action: commandContext(func(ctx Context) error {
				if !ctx.Args().Present() {
					logrus.Fatal("Usage: fission-workflows delete [workflow-id...]")
				}
				client := getClient(ctx)
				for _, wfID := range ctx.Args() {
					if err := client.Workflow.Delete(ctx, wfID); err != nil {
						logrus.Fatalf("Failed to delete %s: %v", wfID, err)
					}
					fmt.Println(wfID)
				}
				return nil
			}),
		},
		{
			Name:  "get",
			Usage: "get <Workflow-id> <task-id>",
			Action: commandContext(func(ctx Context) error {
				client := getClient(ctx)

				switch ctx.NArg() {
				case 0:
					// List workflows
					resp, err := client.Workflow.List(ctx)
					if err != nil {
						panic(err)
					}
					wfs := resp.Workflows
					sort.Strings(wfs)
					var rows [][]string
					for _, wfID := range wfs {
						wf, err := client.Workflow.Get(ctx, wfID)
						if err != nil {
							panic(err)
						}
						updated, _ := ptypes.Timestamp(wf.Status.UpdatedAt)
						created, _ := ptypes.Timestamp(wf.Metadata.CreatedAt)

						rows = append(rows, []string{wfID, wf.Spec.Name, wf.Status.Status.String(),
							created.String(), updated.String()})
					}
					table(os.Stdout, []string{"ID", "NAME", "STATUS", "CREATED", "UPDATED"}, rows)
				case 1:
					// Get Workflow
					wfID := ctx.Args().Get(0)
					wf, err := client.Workflow.Get(ctx, wfID)
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
					wfID := ctx.Args().Get(0)
					taskID := ctx.Args().Get(1)
					wf, err := client.Workflow.Get(ctx, wfID)
					if err != nil {
						panic(err)
					}
					task, ok := wf.Spec.Tasks[taskID]
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
			}),
		},
	},
}
