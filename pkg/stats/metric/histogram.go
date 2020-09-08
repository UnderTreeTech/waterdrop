package metric

import "github.com/prometheus/client_golang/prometheus"

type HistogramVecOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
	Buckets   []float64
}

type histogramVec struct {
	*prometheus.HistogramVec
}

func NewHistogramVec(opt *HistogramVecOpts) *histogramVec {
	vector := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: opt.Namespace,
			Subsystem: opt.Subsystem,
			Name:      opt.Name,
			Help:      opt.Help,
			Buckets:   opt.Buckets,
		}, opt.Labels)
	prometheus.MustRegister(vector)

	return &histogramVec{HistogramVec: vector}
}

func (h *histogramVec) Observe(v float64, labels ...string) {
	h.WithLabelValues(labels...).Observe(v)
}
