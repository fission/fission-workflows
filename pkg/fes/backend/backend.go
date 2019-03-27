package backend

import "github.com/prometheus/client_golang/prometheus"

var (
	EventsAppended = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fes",
		Subsystem: "backend",
		Name:      "events_appended_total",
		Help:      "Count of appended events (including any internal events).",
	}, []string{"eventType"})

	EventDelay = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "fes",
		Subsystem: "backend",
		Name:      "event_propagation_delay",
		Help:      "Delay between event publish and receive by the subscribers.",
		Objectives: map[float64]float64{
			0:     0.0001,
			0.001: 0.0001,
			0.01:  0.0001,
			0.02:  0.0001,
			0.1:   0.0001,
			0.25:  0.0001,
			0.5:   0.0001,
			0.75:  0.0001,
			0.9:   0.0001,
			0.98:  0.0001,
			0.99:  0.0001,
			0.999: 0.0001,
			1:     0.0001,
		},
	})
)

func init() {
	prometheus.MustRegister(EventsAppended, EventDelay)
}
