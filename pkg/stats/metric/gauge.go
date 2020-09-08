package metric

import "github.com/prometheus/client_golang/prometheus"

type GaugeVecOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}

type gaugeVec struct {
	*prometheus.GaugeVec
}

func NewGaugeVec(opt *GaugeVecOpts) *gaugeVec {
	vector := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: opt.Namespace,
			Subsystem: opt.Subsystem,
			Name:      opt.Name,
			Help:      opt.Help,
		}, opt.Labels)
	prometheus.MustRegister(vector)

	return &gaugeVec{GaugeVec: vector}
}

func (g *gaugeVec) Inc(labels ...string) {
	g.WithLabelValues(labels...).Inc()
}

func (g *gaugeVec) Dec(labels ...string) {
	g.WithLabelValues(labels...).Dec()
}

func (g *gaugeVec) Add(v float64, labels ...string) {
	g.WithLabelValues(labels...).Add(v)
}

func (g *gaugeVec) Sub(v float64, labels ...string) {
	g.WithLabelValues(labels...).Sub(v)
}
