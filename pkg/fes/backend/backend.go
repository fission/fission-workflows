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
	})
)

func init() {
	prometheus.MustRegister(EventsAppended, EventDelay)
}
