package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	ticksCount prometheus.Counter
}

func New(nameSpace string) (Metrics, error) {

	m := Metrics{}
	m.ticksCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: nameSpace,
		Name:      "file_client_ticks_count",
		Help:      "file_client_ticks_count",
	})

	return m, m.registerMetrics()
}

func (m Metrics) IncTicker() {
	m.ticksCount.Inc()
}

func (m Metrics) registerMetrics() error {

	if err := ProcessPrometheusError(prometheus.Register(m.ticksCount)); err != nil {
		return err
	}

	return nil
}

func ProcessPrometheusError(err error) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case prometheus.AlreadyRegisteredError:
		return nil
	default:
		return err
	}
}
