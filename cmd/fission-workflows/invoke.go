package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

/*
invoke <workflow-id>  or test

#
--watch => fw invoke + fw status
--async
--debug

# Input params
--input => inputs: --input "{}"
--input-type => "json" (otherwise warn)

# fission function test: (Should basically extend upon function test)
--header => add header input
--body   => add body to input
--query  => add query to input
--method => add method to input

Whatever the method all inputs end up in a WorkflowInvocationSpec
*/
var cmdInvoke = cli.Command{
	Name:  "invoke",
	Usage: "invoke <workflow-id>",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "async",
			Usage: "Invoke workflow asynchronously; invoke without waiting for the result.",
		},
		cli.StringFlag{
			Name:  "inputs",
			Usage: "Sets the inputs to provided value. Expects a JSON object.",
		},
		cli.DurationFlag{
			Name:  "poll",
			Value: 10 * time.Millisecond,
		},
		cli.DurationFlag{
			Name:  "timeout",
			Value: 10 * time.Minute,
		},
	},
	Description: "Invoke a workflow",
	Action: commandContext(func(ctx Context) error {
		listenToEvents := ensureServerVersionAtLeast(ctx, semver.MustParse("0.7.0"), false)
		if !listenToEvents {
			logrus.Warn("Event streaming is not supported < 0.7.0")
		}
		timeout := ctx.Duration("timeout")
		if timeout <= 0 {
			logrus.Fatal("Timeout should be larger than 0")
		}
		if !ctx.Args().Present() {
			logrus.Fatal("Workflow ID is required.")
		}
		workflowID := ctx.Args().First()
		logrus.Infof("Invoking workflow: %v", workflowID)

		inputs := map[string]*typedvalues.TypedValue{}
		if jsonInputs := ctx.String("inputs"); len(jsonInputs) > 0 {
			inputMap := map[string]interface{}{}
			err := json.Unmarshal([]byte(jsonInputs), &inputMap)
			if err != nil {
				logrus.Fatalf("Failed to parse provided inputs to JSON object: %v", err)
			}
			inputs = typedvalues.MustWrapMapTypedValue(inputMap)
		}

		client := getClient(ctx)
		spec := &types.WorkflowInvocationSpec{
			WorkflowId: workflowID,
			Inputs:     inputs,
		}
		types.NewWorkflowInvocationSpec(workflowID, time.Now().Add(timeout))
		md, err := client.Invocation.Invoke(ctx, spec)
		if err != nil {
			logrus.Fatalf("Error occurred while invoking workflow: %v", err)
		}

		// Poll for updates on the progress. In the future we can replace this with a streaming API.
		outputChan := make(chan *types.WorkflowInvocation)
		go func() {
			var offset int
			ticker := time.Tick(ctx.Duration("poll"))
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker:
					pollStatus := !listenToEvents
					if listenToEvents {
						pollStatus, err = fetchAndPrintEvents(ctx, client, md.GetId(), offset, w)
						if err != nil {
							logrus.Error(err)
							logrus.Warn("Halting event streaming; falling back to polling status instead")
							listenToEvents = false
						}
						w.Flush()
					}

					if pollStatus {
						wi, err := client.Invocation.Get(ctx, md.GetId())
						if err != nil {
							logrus.Fatal("Failed to fetch invocation: ", err)
						}
						if wi.GetStatus().Finished() {
							outputChan <- wi
							close(outputChan)
							return
						}
					}
				}
			}
		}()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {
				logrus.Infof("Received signal: %v - cancelling invocation %s", sig, md.Id)
				if err := client.Invocation.Cancel(ctx, md.Id); err != nil {
					panic(err)
				}
			}
		}()

		wi, ok := <-outputChan
		if !ok {
			logrus.Fatal("No output could be retrieved.")
		}
		// Display output or error
		logrus.Debugf("Invocation status: %s", wi.GetStatus().GetStatus().String())
		if wi.GetStatus().Successful() {
			fmt.Println(typedvalues.MustUnwrap(wi.GetStatus().GetOutput()))
		} else {
			logrus.Error(wi.GetStatus().GetError().GetMessage())
			os.Exit(1)
		}
		return nil
	}),
}

func fetchAndPrintEvents(ctx context.Context, client client, invocationID string, offset int,
	w io.Writer) (finished bool,
	err error) {
	// Future: support and use offset in invocationEvents API requests
	invocationEvents, err := client.Invocation.Events(ctx, invocationID)
	if err != nil {
		return finished, fmt.Errorf("failed to fetch events: %v", err)
	}

	// Traverse all new invocationEvents
	for _, event := range invocationEvents.GetEvents()[offset:] {
		invocationEvent, err := parseInvocationEvent(event)
		if err != nil {
			return finished, fmt.Errorf("failed to parse events: %v", err)
		}

		if e, ok := invocationEvent.data.(events.Event); ok {
			switch e.Type() {
			case events.EventInvocationFailed:
				invocationEvent.subjectType = subjectTypeError
				finished = true
			case events.EventInvocationCanceled:
				invocationEvent.subjectType = subjectTypeError
				finished = true
			case events.EventInvocationCompleted:
				invocationEvent.subjectType = subjectTypeSuccess
				finished = true
			case events.EventTaskSucceeded:
				invocationEvent.subjectType = subjectTypeSuccess
				invocationEvent.target = event.GetId() // TODO can we get task name
			}
		}

		w.Write([]byte(eventToTabString(invocationEvent)))
		offset++
	}
	return finished, nil
}

type subjectType int

const (
	subjectTypeRegular = iota
	subjectTypeSuccess
	subjectTypeError
)

// Event is helper struct for pretty-printing events
type event struct {
	timestamp   time.Time
	subject     string
	subjectType subjectType
	target      string // TODO add target task
	message     string
	data        proto.Message
}

func parseInvocationEvent(fesEvent *fes.Event) (*event, error) {
	e, err := fes.ParseEventData(fesEvent)
	if err != nil {
		return nil, err
	}
	subject := reflect.TypeOf(e).String()[strings.Index(reflect.TypeOf(e).String(), ".")+1:]
	message := e.String()

	ts, err := ptypes.Timestamp(fesEvent.Timestamp)
	if err != nil {
		return nil, err
	}
	return &event{
		timestamp:   ts,
		subject:     subject,
		message:     message,
		subjectType: subjectTypeRegular,
		data:        e,
	}, nil
}

func eventToTabString(e *event) string {
	var subject string
	switch e.subjectType {
	case 1:
		subject = color.HiGreenString(e.subject)
	case 2:
		subject = color.HiRedString(e.subject)
	default:
		subject = color.HiYellowString(e.subject)
	}
	return fmt.Sprintf("%s\t%s\t%s\t%v\n", e.timestamp.String(), subject, e.target, e.message)
}
