/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package metric

import "github.com/prometheus/client_golang/prometheus"

// HistogramVecOpts
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

// NewHistogramVec returns a HistogramVecOpts instance
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

// Observe add observations to Histogram.
func (h *histogramVec) Observe(v float64, labels ...string) {
	h.WithLabelValues(labels...).Observe(v)
}
