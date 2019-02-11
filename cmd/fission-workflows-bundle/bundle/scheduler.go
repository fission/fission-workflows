package bundle

import (
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/urfave/cli"
)

type SchedulerPolicy string

const (
	FlagSchedulerPolicy            = "scheduler.policy"
	FlagSchedulerColdStartDuration = "scheduler.coldstart"
)

var schedulerPolicies = map[string]func(time.Duration) scheduler.Policy{
	"prewarm-all": func(coldStartModel time.Duration) scheduler.Policy {
		return scheduler.Policy(scheduler.NewPrewarmAllPolicy(coldStartModel))
	},
	"prewarm-horizon": func(coldStartModel time.Duration) scheduler.Policy {
		return scheduler.Policy(scheduler.NewPrewarmHorizonPolicy(coldStartModel))
	},
	"horizon": func(_ time.Duration) scheduler.Policy { return scheduler.Policy(scheduler.NewHorizonPolicy()) },
}

func ParseSchedulerConfig(c *cli.Context) (scheduler.Policy, error) {
	policyName := c.String(FlagSchedulerPolicy)
	policy, ok := schedulerPolicies[policyName]
	if !ok {
		return nil, fmt.Errorf("unknown scheduler policy '%s'", policyName)
	}
	return policy(c.Duration(FlagSchedulerColdStartDuration)), nil
}

func RunScheduler(policy scheduler.Policy) *scheduler.InvocationScheduler {
	if policy == nil {
		panic("scheduler policy expected")
	}
	return scheduler.NewInvocationScheduler(policy)
}
