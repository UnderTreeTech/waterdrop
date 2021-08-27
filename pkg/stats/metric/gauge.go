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

// GaugeVecOpts
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

// NewGaugeVec returns a GaugeVecOpts instance
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

// Inc increments the Gauge by 1
func (g *gaugeVec) Inc(labels ...string) {
	g.WithLabelValues(labels...).Inc()
}

// Dec decrements the Gauge by 1
func (g *gaugeVec) Dec(labels ...string) {
	g.WithLabelValues(labels...).Dec()
}

// Add adds the given value to the Gauge
func (g *gaugeVec) Add(v float64, labels ...string) {
	g.WithLabelValues(labels...).Add(v)
}

// Sub subtracts the given value from the Gauge
func (g *gaugeVec) Sub(v float64, labels ...string) {
	g.WithLabelValues(labels...).Sub(v)
}
