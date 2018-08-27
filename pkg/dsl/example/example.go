package example

import (
	"time"

	"github.com/fission/fission-workflows/pkg/dsl"
)

func Flow() dsl.Flow {
	fetchWeather := &dsl.TaskSpec{Run: "weather-fetch"}
	fetchWeather.Input("location", "Delft, The Netherlands")

	waitSeconds := dsl.Sleep(time.Second)
	waitSeconds.Require(fetchWeather)

	sendSlackMsg := &dsl.TaskSpec{
		Run: "slack-send",
		Inputs: dsl.Inputs{
			"msg": dsl.Expression("$.Tasks.FetchWeather.Output"),
		},
		Requires: []*dsl.TaskSpec{
			fetchWeather,
			waitSeconds,
		},
	}

	return sendSlackMsg
}
