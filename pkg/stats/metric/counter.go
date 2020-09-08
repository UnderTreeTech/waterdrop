package metric

import "github.com/prometheus/client_golang/prometheus"

type CounterVecOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}

type counterVec struct {
	*prometheus.CounterVec
}

func NewCounterVec(opt *CounterVecOpts) *counterVec {
	vector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: opt.Namespace,
			Subsystem: opt.Subsystem,
			Name:      opt.Name,
			Help:      opt.Help,
		}, opt.Labels)
	prometheus.MustRegister(vector)

	return &counterVec{CounterVec: vector}
}

func (c *counterVec) Inc(labels ...string) {
	c.WithLabelValues(labels...).Inc()
}

func (c *counterVec) Add(v float64, labels ...string) {
	c.WithLabelValues(labels...).Add(v)
}
